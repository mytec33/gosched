package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/manifest"
	"git.sr.ht/~mytec/gosched/internal/policy"
	"git.sr.ht/~mytec/gosched/internal/schedule"
	"git.sr.ht/~mytec/gosched/internal/types"
)

const (
	ExitSuccess                int = 0
	ExitNoConfig               int = 1
	ExitValidation             int = 3
	ExitInvalidArgs            int = 4
	ExitRunOnce                int = 5
	ExitDeprecatedScheduleFlag int = 7
	ExitManifestError          int = 8
)

var (
	ErrRunCommandAbortsOnError = errors.New("run command aborts on error")
)

func main() {
	os.Exit(run())
}

func run() int {
	manifestFlag := ""
	newConfig := false
	runOnce := false
	printSchedule := ""

	flag.StringVar(&manifestFlag, "manifest", "", "path to a manifest file containing schedule configuration files to load")
	flag.BoolVar(&newConfig, "new-config", false, "create an new configuration to begin with")
	flag.BoolVar(&runOnce, "run-once", false, "bypass any schedule and run now")
	flag.StringVar(&printSchedule, "print-schedule", "", "configuration summary: config, operational")
	flag.Parse()

	if newConfig {
		generateNewConfig()
		return ExitSuccess
	}

	if manifestFlag == "" {
		logging.StdErr.Error("startup", "reason", "missing required -manifest argument")
		flag.Usage()
		return ExitInvalidArgs
	}

	scheduleFiles, err := manifest.ParseManifest(manifestFlag)
	if err != nil {
		logging.StdErr.Error("startup", "reason", "no files found in manifest", "error", err)
		return ExitManifestError
	}

	if len(scheduleFiles) == 0 {
		logging.StdErr.Error("startup", "reason", "manifest contains no schedule files")
		flag.Usage()
		return ExitInvalidArgs
	}

	sched, valErrors, err := schedule.ReadScheduleFiles(scheduleFiles)
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
	logging.StdOut.Info("startup", "reason", "workflows loaded", "count", sched.WorkflowCount(), "filename", scheduleFiles)

	runSchedule(sched, runOnce)

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

func runSchedule(s schedule.Schedule, runOnce bool) {
	if runOnce {
		currentMinute, err := types.ParseMinuteOfDay(time.Now().Format("15:04"))
		if err != nil {
			logging.StdErr.Error("run scheduler tick", "status", "failed", "reason", err)
			return
		}

		runSchedulerTick(currentMinute, s)
		return
	}

	// Align to the next minute boundary once, then tick.
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		currentMinute, err := types.ParseMinuteOfDay(time.Now().Format("15:04"))
		if err != nil {
			logging.StdErr.Error("run scheduler tick", "status", "failed", "reason", err)
			return
		}

		n := runSchedulerTick(currentMinute, s)
		if n > 0 {
			logging.StdOut.Info("scheduler", "scheduled workflows",
				n, "minute", currentMinute.String())
		}
	}
}

func runSchedulerTick(currentMinute types.MinuteOfDay, s schedule.Schedule) int {
	logging.StdOut.Info("run scheduler tick", "current_minute", currentMinute)

	tasks := s.WorkflowsAtMinute(currentMinute)
	if len(tasks) == 0 {
		return 0
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
			stdErr.Error("run step", "workflow", wf.Name, "step", step.Name, "stepIndex", i, "status", "failed",
				"exitCode", result.ExitCode, "duration", stepDuration, "reason", result.Err)

			if workflowAbortsOnFailure(wf) {
				stdErr.Error("run step", "workflow", wf.Name, "step", step.Name, "stepIndex", i,
					"policy", "abort", "reason", "step failed")
				return fmt.Errorf("%w: %s", ErrRunCommandAbortsOnError, result.Err)
			}

			if workflowContinuesOnFailure(wf) {
				stdErr.Error("run step", "workflow", wf.Name, "step", step.Name, "stepIndex", i,
					"policy", "continue", "reason", "step failed")
				continue
			}
		} else {
			stdOut.Info("run step", "workflow", wf.Name, "step", step.Name, "stepIndex", i, "status", "completed",
				"stepName", step.Name, "exitCode", result.ExitCode, "duration", stepDuration)
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
	stdOut.Info("workflow", "status", "completed", "duration", workflowDuration)
	return nil
}

func workflowAbortsOnFailure(wf schedule.Workflow) bool {
	return wf.OnFailure != nil && *wf.OnFailure == policy.Abort
}

func workflowContinuesOnFailure(wf schedule.Workflow) bool {
	return wf.OnFailure != nil && *wf.OnFailure == policy.Continue
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
