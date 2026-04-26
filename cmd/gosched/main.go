package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/policy"
	"git.sr.ht/~mytec/gosched/internal/schedule"
	"git.sr.ht/~mytec/gosched/internal/types"
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
	var schedules schedule.ScheduleSliceFlag
	newConfig := false
	runOnce := false
	printSchedule := ""
	flag.Var(&schedules, "schedule", "file containing a schedule to run")
	flag.BoolVar(&newConfig, "new-config", false, "create an new configuration to begin with")
	flag.BoolVar(&runOnce, "run-once", false, "bypass any schedule and run now")
	flag.StringVar(&printSchedule, "print-schedule", "", "configuration summary: config, operational")
	flag.Parse()

	if newConfig {
		generateNewConfig()
		return ExitSuccess
	}

	if len(schedules) == 0 {
		logging.StdErr.Error("startup", "reason", "invalid args")
		flag.Usage()
		return ExitInvalidArgs
	}

	sched, valErrors, err := schedule.ReadScheduleFiles(schedules)
	if err != nil {
		logging.StdErr.Error("startup", "reason", "failed to load schedule", "error", err)
		return ExitNoConfig
	}

	if len(valErrors) > 0 {
		displayCfgErrors(valErrors)
		return ExitValidation
	}

	if printSchedule != "" {
		printConfiguration(printSchedule, sched)
		return ExitSuccess
	}

	// This goes after newConfig or any other option that prints to STDOUT so only the output we
	// wish to print is shown and not logging messages. Those don't play well with JSON. :-)
	logging.StdOut.Info("startup", "reason", "scheduler service started")
	logging.StdOut.Info("startup", "reason", "workflows loaded", "count", sched.WorkflowCount(), "filename", schedules)

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

func printConfiguration(method string, s schedule.Schedule) {
	switch method {
	case "config":
		s.PrintScheduleConfig(os.Stdout)
	case "operational":
		s.PrintScheduleOperational(os.Stdout)
	default:
		fmt.Printf("Unknown print config method: %s\n", method)
	}
}

func displayCfgErrors(errors []error) {
	logging.StdErr.Error("startup", "reason", "configuration invalid")

	for _, err := range errors {
		logging.StdErr.Error("startup", "reason", err)
	}
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
	currentMinute, err := types.ParseMinuteOfDay(now.Format("15:04"))
	if err != nil {
		logging.StdErr.Error("run scheduler tick", "status", "failed", "reason", err)
		return 1
	}
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

		stdOut.Info("step", "status", "started", "stepIndex", i, "stepName", step.Name, "args", step.Args)

		result := schedule.RunStepCommand(step)
		stepDuration := time.Since(stepStart)

		if len(result.Output) != 0 {
			stdOut.Info("step output", "data", result.Output)
		}

		if result.Err != nil {
			stdErr.Error(
				"step",
				"status", "failed",
				"exitCode", result.ExitCode,
				"duration", stepDuration,
				"reason", result.Err,
			)

			if workflowAbortsOnFailure(wf) {
				return fmt.Errorf("workflow %q step %d (%s) failed: %w", wf.Name, i, step.Name, result.Err)
			}
		} else {
			stdOut.Info(
				"step",
				"status", "completed",
				"stepIndex", i,
				"stepName", step.Name,
				"exitCode", result.ExitCode,
				"duration", stepDuration,
			)
		}

		if step.Pause.Configured() {
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
	stdOut.Info("workflow", "status", "completed", "duration", workflowDuration)
	return nil
}

func workflowAbortsOnFailure(wf schedule.Workflow) bool {
	return wf.OnFailure != nil && *wf.OnFailure == policy.Abort
}

func generateNewConfig() {
	newConfig := `
[
  {
    "name": "Workflow 1",
    "time": "10:00",
    "onFailure": "continue",
    "steps": [
      {
        "name": "step 1",
        "program": "/usr/bin/some_program",
        "args": [
            "--verbose",
            "--file",
            "some_file_name"
          ],
        "timeout": 11,
        "pause": 3
      },
      {
        "name": "step 2",
        "program": "/usr/bin/some_program",
        "args": [],
        "pause": 0
      }
    ]
  },
  {
    "name": "Workflow 2",
    "time": "10:35",
    "onFailure": "abort",
    "steps": [
      {
        "name": "daily",
        "program": "/opt/sbin/some_script",
        "args": [
          "--sleep",
          "5"
        ]
      }
    ]
  }
]	
`

	fmt.Println(newConfig)
}
