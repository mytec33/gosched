package runner

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/schedule"
	"git.sr.ht/~mytec/gosched/internal/workflow"
)

const (
	diagnosticTimeLayout = "15:04:05.000000000"
)

var (
	ErrRunCommandAbortsOnError = errors.New("run command aborts on error")
	ErrSchedulerMinuteParse    = errors.New("scheduler minute parse failed")
	ErrWorkflowNotFoundByName  = errors.New("work flow not found in file(s) loaded by name")
	ErrExecutingWorkflow       = errors.New("execute workflow failure")
	ErrSignalInterrupt         = errors.New("received interrupt signal")
)

type nextWakeTiming struct {
	Current time.Time
	Target  time.Time
	Wait    time.Duration
}

func alignToNextMinuteBoundary(ctx context.Context) error {
	wake := nextWakeDuration()
	logging.StdOut.Info("scheduler", "reason", "align to next minute boundary",
		"current", wake.Current.Format(diagnosticTimeLayout),
		"target_wake", wake.Target.Format(diagnosticTimeLayout),
		"sleep", wake.Wait.String(),
	)

	select {
	case <-ctx.Done():
		logging.StdOut.Error("scheduler", "reason", "signal interrupt received before scheduler loop started")
		return ErrSignalInterrupt
	case <-time.After(wake.Wait):
	}

	return nil
}

func RunSchedule(ctx context.Context, s schedule.Schedule) error {
	var running sync.WaitGroup

	// Capture lastProcessed before aligning so the first post-alignment minute
	// is treated as new work rather than a duplicate time.
	lastProcessed := workflow.MinuteOfDayFromTime(time.Now())

	err := alignToNextMinuteBoundary(ctx)
	if err != nil {
		return err
	}

	for {
		now := time.Now()
		currentMinute := workflow.MinuteOfDayFromTime(now)
		minuteDiff := currentMinute.MinutesSince(lastProcessed)

		switch minuteDiff {
		case 0:
			logging.StdOut.Warn("scheduler", "reason", "duplicate suppression", "currentMinute",
				currentMinute.String(), "lastProcessed", lastProcessed.String())
		default:
			if minuteDiff > 1 {
				logging.StdOut.Warn("scheduler", "reason", "skipped minute(s)", "missed", minuteDiff-1,
					"currentMinute", currentMinute.String(), "lastProcessed", lastProcessed.String())
			}

			n := runSchedulerTick(currentMinute, s, &running)
			if n > 0 {
				logging.StdOut.Info("scheduler", "scheduled workflows",
					n, "minute", currentMinute.String())
			}

			lastProcessed = currentMinute
		}

		wake := nextWakeDuration()
		logging.StdOut.Info("scheduler", "reason", "sleep diagnostic",
			"current", wake.Current.Format(diagnosticTimeLayout),
			"target_wake", wake.Target.Format(diagnosticTimeLayout),
			"sleep", wake.Wait.String(),
		)
		timer := time.NewTimer(wake.Wait)

		select {
		case <-ctx.Done():
			timer.Stop()
			logging.StdOut.Error("scheduler", "reason", "signal interrupt received; waiting for running workflows to finish")
			logging.StdOut.Error("scheduler", "reason", "shutdown in progress; not starting new work")
			running.Wait()
			return ErrSignalInterrupt
		case <-timer.C:
			logging.StdOut.Info("scheduler", "reason", "wake diagnostic", "wake", time.Now().Format(diagnosticTimeLayout))
		}
	}
}

func nextWakeDuration() nextWakeTiming {
	// Add extra milliseconds to avoid landing .999 and creating a duplicate minute occurrence
	now := time.Now()
	nextMinute := now.Truncate(time.Minute).Add(time.Minute).Add(5 * time.Millisecond)

	return nextWakeTiming{
		Current: now,
		Target:  nextMinute,
		Wait:    nextMinute.Sub(now),
	}
}

func RunScheduleOnce(sched schedule.Schedule, wfName string) error {
	wf, err := sched.GetWorkflowByName(wfName)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWorkflowNotFoundByName, err)
	}

	err = executeWorkflow(wf)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExecutingWorkflow, err)
	}

	return nil
}

func runSchedulerTick(currentMinute workflow.MinuteOfDay, s schedule.Schedule, running *sync.WaitGroup) int {
	logging.StdOut.Info("run scheduler tick", "current_minute", currentMinute.String())

	tasks := s.WorkflowsAtMinute(currentMinute)
	if len(tasks) == 0 {
		return 0
	}

	for _, task := range tasks {
		running.Add(1)
		go func(w workflow.Workflow) {
			defer running.Done()

			err := executeWorkflow(w)
			if err != nil {
				logging.StdOut.Error("workflow", "status", workflow.StatusFailed.String(),
					"error", err)
			}
		}(task)
	}
	return len(tasks)
}

func executeWorkflow(wf workflow.Workflow) error {
	workflowStart := time.Now()

	wfLog := logging.NewWorkflowLogger(wf.Name)
	stdOut := wfLog.Out

	if !wf.Enabled {
		stdOut.Info("workflow",
			"name", wf.Name,
			"status", workflow.StatusSkipped.String(),
			"reason", "disabled",
			"disabledReason", wf.DisabledReason,
		)
		return nil
	}

	lockKey := wf.Name
	existingID, acquired := workflowLocks.tryAcquire(lockKey, wfLog.WfRunID)
	if !acquired {
		stdOut.Error("workflow", "status", workflow.StatusSkipped.String(), "reason",
			"workflow already running", "existingRunID", existingID)
		return nil
	}
	defer workflowLocks.delete(lockKey)

	numSteps := len(wf.Steps)
	stdOut.Info("workflow", "name", wf.Name, "status", "started", "stepCount", numSteps)

	workflowStatus := workflow.StatusCompleted
	for i, step := range wf.Steps {
		result := runStepAttempt(stdOut, wf.Name, step, i)

		if result.Err != nil {
			if workflow.WorkflowAbortsOnFailure(wf) {
				stdOut.Info("step", "step", step.Name, "stepIndex", i, "status", wf.OnFailure)
				return fmt.Errorf("%w: %s", ErrRunCommandAbortsOnError, result.Err)
			}

			if workflow.WorkflowContinuesOnFailure(wf) {
				workflowStatus = workflow.StatusPartial
				continue
			}

			if workflow.WorkflowRetriesOnFailure(wf) {
				retryStatus := runStepRetries(stdOut, wf.Name, step, i, wf.Retry)
				if retryStatus == workflow.StatusPartial {
					workflowStatus = workflow.StatusPartial
				}
			}
		}

		if step.Pause.Duration() > 0 {
			pause := step.Pause.Duration()

			if i < numSteps-1 {
				stdOut.Info("step", "status", "paused", "duration", pause)
				time.Sleep(pause)
			} else {
				stdOut.Info("step", "status", "skipped pause", "reason", "last step")
			}
		}
	}

	workflowDuration := time.Since(workflowStart)
	stdOut.Info("workflow", "workflow", wf.Name, "status", workflowStatus.String(), "duration", workflowDuration)
	return nil
}
