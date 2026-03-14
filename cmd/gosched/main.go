package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/policy"
	"git.sr.ht/~mytec/gosched/internal/schedule"
)

const (
	ExitSuccess     int = 0
	ExitNoConfig    int = 1
	ExitValidation  int = 3
	ExitInvalidArgs int = 4
	ExitRunOnce     int = 5
)

type TickFunc func(now time.Time, data schedule.Schedule, runOnce bool) int

func main() {
	os.Exit(run())
}

func run() int {
	logging.StdOut.Info("startup", "reason", "scheduler service started")

	filename := ""
	runOnce := false
	summarizeConfig := false
	flag.StringVar(&filename, "schedule", "", "file containing a schedule to run")
	flag.BoolVar(&runOnce, "run-once", false, "bypass any schedule and run now")
	flag.BoolVar(&summarizeConfig, "summarize-config", false, "show concise summary of configuration file schedule")
	flag.Parse()

	if filename == "" {
		logging.StdErr.Error("startup", "reason", "invalid args")
		flag.Usage()
		return ExitInvalidArgs
	}

	sched, valErrors, err := schedule.ReadScheduleFile(filename)
	if err != nil {
		logging.StdErr.Error("startup", "reason", "failed to load schedule", "error", err)
		return ExitNoConfig
	}

	if len(valErrors) > 0 {
		displayCfgErrors(valErrors)
		return ExitValidation
	}

	if summarizeConfig {
		displayConfigSummarization()
		return ExitSuccess
	}

	logging.StdOut.Info("startup", "reason", "workflows loaded", "count", sched.WorkflowCount(), "filename", filename)

	if runOnce {
		failures := runSchedulerTick(time.Now(), sched, true)
		logging.StdOut.Info("scheduler", "event", "run-once-finished", "failures", failures)
		if failures > 0 {
			return ExitRunOnce
		}
		return ExitSuccess
	}

	runScheduler(runSchedulerTick, sched, runOnce)

	return 0
}

func displayConfigSummarization() {
	fmt.Println("display configuration summarization")
}

func displayCfgErrors(errors []error) {
	logging.StdErr.Error("startup", "reason", "configuration invalid")

	for _, err := range errors {
		logging.StdErr.Error("startup", "reason", err)
	}
}

func programExists(prog string) error {
	_, err := os.Stat(prog)
	if err != nil {
		return fmt.Errorf("program not found: %s", prog)
	}

	return nil
}

func runScheduler(tick TickFunc, s schedule.Schedule, runOnce bool) {
	if runOnce {
		tick(time.Now().Truncate(time.Minute), s, runOnce)
	}

	// Align to the next minute boundary once, then tick.
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().Truncate(time.Minute)
		n := tick(now, s, runOnce)
		if n > 0 {
			logging.StdOut.Info("scheduler", "scheduled workflows",
				n, "minute", now.Format("15:04"),
			)
		}
	}
}

func runSchedulerTick(now time.Time, s schedule.Schedule, runOnce bool) int {
	currentMinute := schedule.MinuteKey(now.Format("15:04"))
	logging.StdOut.Info("run scheduler tick", "current_minute", currentMinute)

	tasks := s.WorkflowsAtMinute(currentMinute)
	if len(tasks) == 0 {
		return 0
	}

	if runOnce {
		failures := 0
		for _, task := range tasks {
			err := executeWorkflow(task)
			if err != nil {
				failures++
				logging.StdErr.Error("workflow", "status", "failed", "workflow", task.Name, "error", err)
			}
		}
		return failures
	}

	for _, task := range tasks {
		go func(w schedule.Workflow) {
			err := executeWorkflow(w)
			if err != nil {
				logging.StdErr.Error("workflow", "status", "failed", "error", err)
			}
		}(task)
	}
	return len(tasks)
}

func executeWorkflow(wf schedule.Workflow) error {
	workflowStart := time.Now()

	wfLog := logging.NewWorkflowLogger(wf.Name)
	stdOut := wfLog.Out
	stdErr := wfLog.Err

	lockKey := wf.Name
	existingID, running := schedule.RunningWorkflows.Get(lockKey)
	if running {
		stdErr.Error("workflow", "status", "skipped", "reason", "workflow already running", "existingRunID", existingID)
		return nil
	}
	schedule.RunningWorkflows.Set(lockKey, wfLog.WfRunID)
	defer schedule.RunningWorkflows.Delete(lockKey)

	stdOut.Info("workflow", "status", "started")

	numSteps := len(wf.Steps)
	for i, step := range wf.Steps {
		stepStart := time.Now()

		stdOut.Info("workflow step", "status", "started", "stepIndex", i, "stepName", step.Name, "args", step.Args)

		err := programExists(step.Program)
		if err != nil {
			stdErr.Error("workflow step", "status", "cannot find program", "stepIndex", i, "stepName", step.Name, "program", step.Program)
			continue
		}

		var cmd *exec.Cmd
		var cancel context.CancelFunc

		args := make([]string, 0, len(step.Args))
		for _, a := range step.Args {
			args = append(args, a.String())
		}

		if step.Timeout.Configured() {
			ctx, c := context.WithTimeout(context.Background(), step.Timeout.Duration())
			cancel = c
			cmd = exec.CommandContext(ctx, step.Program, args...)
		} else {
			cmd = exec.Command(step.Program, args...)
		}
		output, err := cmd.CombinedOutput()
		stdOut.Info("workflow step", "status", "output", "stepIndex", i, "stepName", step.Name,
			"output", output)

		if cancel != nil {
			cancel()
		}

		stepDuration := time.Since(stepStart)

		if err != nil && wf.OnFailure == &policy.Abort {
			stdErr.Error("workflow step", "status", "failed", "stepIndex", i, "stepName", step.Name,
				"duration", stepDuration, "reason", err)
			return fmt.Errorf("workflow %q step %d (%s) failed: %w", wf.Name, i, step.Name, err)
		}

		stdOut.Info("workflow step", "status", "completed", "stepIndex", i, "stepName", step.Name, "duration", stepDuration)

		if step.Pause.Configured() {
			pause := step.Pause.Duration()

			if i < numSteps-1 {
				stdOut.Info("workflow step", "status", "pause started", "duration", pause, "stepIndex", i, "stepName", step.Name)
				time.Sleep(pause)
			} else {
				stdOut.Info("workflow step", "status", "pause skipped", "reason", "last step", "stepIndex", i, "stepName", step.Name)
			}
		}
	}

	workflowDuration := time.Since(workflowStart)
	stdOut.Info("workflow", "status", "completed", "duration", workflowDuration)
	return nil
}
