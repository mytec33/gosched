package runner

import (
	"errors"
	"fmt"
	"time"

	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/schedule"
	"git.sr.ht/~mytec/gosched/internal/types"
)

var (
	ErrRunCommandAbortsOnError = errors.New("run command aborts on error")
	ErrSchedulerMinuteParse    = errors.New("scheduler minute parse failed")
	ErrWorkflowNotFoundByName  = errors.New("work flow not found in file(s) loaded by name")
	ErrExecutingWorkflow       = errors.New("execute workflow failure")
)

func RunSchedule(s schedule.Schedule) error {
	scheduleBegan := time.Now()

	lastProcessed, err := types.ParseMinuteOfDay(scheduleBegan.Format("15:04"))
	if err != nil {
		return fmt.Errorf("%w: %q: %w", ErrSchedulerMinuteParse, scheduleBegan, err)
	}

	// Align to the next clean minute boundary once
	nextBoundary := scheduleBegan.Truncate(time.Minute).Add(time.Minute)
	logging.StdOut.Info("scheduler syncing to next clean minute boundary", "next_boundary", nextBoundary)
	time.Sleep(time.Until(nextBoundary))

	for {
		logging.StdOut.Info("scheduler", "reason", "wake diagnostic",
			"wake", time.Now().Format(time.RFC3339Nano),
		)

		now := time.Now()
		currentMinute, err := types.ParseMinuteOfDay(now.Format("15:04"))
		if err != nil {
			return fmt.Errorf("%w: %q: %w", ErrSchedulerMinuteParse, now, err)
		}

		// duplicate = 0, normal = 1, skipped = > 1
		minuteDiff := currentMinute.MinutesSince(lastProcessed)

		switch {
		case minuteDiff == 0:
			logging.StdOut.Warn("scheduler", "reason", "duplicate suppression", "currentMinute",
				currentMinute.String(), "lastProcessed", lastProcessed.String())
		case minuteDiff > 1:
			logging.StdOut.Warn("scheduler", "reason", "skipped minute(s)", "missed", minuteDiff-1,
				"currentMinute", currentMinute.String(), "lastProcessed", lastProcessed.String())

			// We have skipped one or more minutes but we can still run the current minute
			fallthrough
		default:
			n := runSchedulerTick(currentMinute, s)
			if n > 0 {
				logging.StdOut.Info("scheduler", "scheduled workflows",
					n, "minute", currentMinute.String())
			}

			lastProcessed = currentMinute
		}

		// Fresh time so we sleep as close to the next minute boundary as possible.
		nextMinute := time.Now().Truncate(time.Minute).Add(time.Minute).Add(5 * time.Millisecond)
		logging.StdOut.Info("scheduler", "reason", "sleep diagnostic", "current", time.Now().String(),
			"sleep", nextMinute.String())
		time.Sleep(time.Until(nextMinute))
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

func runSchedulerTick(currentMinute types.MinuteOfDay, s schedule.Schedule) int {
	logging.StdOut.Info("run scheduler tick", "current_minute", currentMinute.String())

	tasks := s.WorkflowsAtMinute(currentMinute)
	if len(tasks) == 0 {
		return 0
	}

	for _, task := range tasks {
		go func(w schedule.Workflow) {
			err := executeWorkflow(w)
			if err != nil {
				logging.StdOut.Error("workflow", "status", types.WorkflowStatusFailed.String(),
					"error", err)
			}
		}(task)
	}
	return len(tasks)
}

func executeWorkflow(wf schedule.Workflow) error {
	workflowStart := time.Now()

	wfLog := logging.NewWorkflowLogger(wf.Name)
	stdOut := wfLog.Out

	lockKey := wf.Name
	existingID, acquired := schedule.RunningWorkflows.TryAcquire(lockKey, wfLog.WfRunID)
	if !acquired {
		stdOut.Error("workflow", "status", types.WorkflowStatusSkipped.String(), "reason",
			"workflow already running", "existingRunID", existingID)
		return nil
	}
	defer schedule.RunningWorkflows.Delete(lockKey)

	numSteps := len(wf.Steps)
	stdOut.Info("workflow", "name", wf.Name, "status", "started", "stepCount", numSteps)

	workflowStatus := types.WorkflowStatusCompleted
	for i, step := range wf.Steps {
		result := RunStepAttempt(stdOut, wf.Name, step, i)

		if result.Err != nil {
			if schedule.WorkflowAbortsOnFailure(wf) {
				stdOut.Info("step", "step", step.Name, "stepIndex", i, "status", wf.OnFailure)
				return fmt.Errorf("%w: %s", ErrRunCommandAbortsOnError, result.Err)
			}

			if schedule.WorkflowContinuesOnFailure(wf) {
				workflowStatus = types.WorkflowStatusPartial
				continue
			}

			if schedule.WorkflowRetriesOnFailure(wf) {
				retryStatus := RunStepRetries(stdOut, wf.Name, step, i, wf.Retry)
				if retryStatus == types.WorkflowStatusPartial {
					workflowStatus = types.WorkflowStatusPartial
				}
			}
		}

		if step.Pause > 0 {
			pause := time.Duration(step.Pause) * time.Second

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
