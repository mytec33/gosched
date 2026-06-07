package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/manifest"
	"git.sr.ht/~mytec/gosched/internal/runner"
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
	ExitWorkflowNotFoundByName int = 9
	ExitExecuteWorkflow        int = 10
	ExitScheduleValidation     int = 11
	ExitScheduleExpansion      int = 12
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
	runThisOnce := ""
	printSchedule := ""

	flag.StringVar(&manifestFlag, "manifest", "", "path to a manifest file containing schedule configuration files to load")
	flag.BoolVar(&newConfig, "new-config", false, "create an new configuration to begin with")
	flag.StringVar(&runThisOnce, "run-once", "", "run a workflow by name bypassing its scheduled time")
	flag.StringVar(&printSchedule, "print-schedule", "", "configuration summary: config, operational")
	flag.Parse()

	if newConfig {
		generateNewConfig()
		return ExitSuccess
	}

	if manifestFlag == "" {
		logging.StdOut.Error("startup", "reason", "missing required -manifest argument")
		flag.Usage()
		return ExitInvalidArgs
	}

	scheduleFiles, err := manifest.ParseManifest(manifestFlag)
	if err != nil {
		logging.StdOut.Error("startup", "reason", "no files found in manifest", "error", err)
		return ExitManifestError
	}

	if len(scheduleFiles) == 0 {
		logging.StdOut.Error("startup", "reason", "manifest contains no schedule files")
		flag.Usage()
		return ExitInvalidArgs
	}

	sched, decodeErrors, err := schedule.ReadScheduleFiles(scheduleFiles)
	if err != nil {
		logging.StdOut.Error("startup", "reason", "failed to load schedule", "error", err)
		return ExitNoConfig
	}

	// Let's user see in terminal output the files loaded that lead to this conclusion
	logging.StdOut.Info("startup", "reason", "schedule files read", "count", sched.WorkflowCount(),
		"filename", scheduleFiles,
	)

	if len(decodeErrors) > 0 {
		displayCfgErrors(decodeErrors)
		return ExitValidation
	}

	valErrors := sched.Validate()
	if len(valErrors) > 0 {
		logging.StdOut.Error("startup", "reason", "failed to validate schedule", "error(s)", valErrors)
		return ExitScheduleValidation
	}

	err = sched.ExpandSchedule()
	if err != nil {
		logging.StdOut.Error("startup", "reason", "schedule expansion failed", "error", err)
		return ExitScheduleExpansion
	}

	if printSchedule != "" {
		return printConfiguration(printSchedule, sched)
	}

	// This goes after newConfig or any other option that prints to STDOUT so only the output we
	// wish to print is shown and not logging messages. Those don't play well with JSON. :-)
	logging.StdOut.Info("startup", "reason", "scheduler service started")

	if runThisOnce != "" {
		logging.StdOut.Info("startup", "reason", "run once started", "workFlow", runThisOnce)

		exitCode, err := runScheduleOnce(sched, runThisOnce)
		if err != nil {
			logging.StdOut.Info("run once", "reason", err)
			return exitCode
		}
	} else {
		runSchedule(sched)
	}

	return 0
}

func printConfiguration(method string, s schedule.Schedule) int {
	switch method {
	case "config":
		s.PrintScheduleConfig(os.Stdout)
	case "operational":
		s.PrintScheduleOperational(os.Stdout)
	default:
		fmt.Printf("Unknown print config method: %s\n", method)
		return ExitInvalidArgs
	}

	return ExitSuccess
}

func displayCfgErrors(fileErrors []error) {
	logging.StdOut.Error("startup", "reason", "configuration invalid")

	for _, err := range fileErrors {
		var fileErr schedule.FileValidationError

		if errors.As(err, &fileErr) {
			logging.StdOut.Error("startup", "file", fileErr.File,
				"reason", fileErr.Err)
		} else {
			logging.StdOut.Error("startup", "reason", err)
		}
	}
}

func runSchedule(s schedule.Schedule) {
	// Align to the next minute boundary once, then tick.
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		currentMinute, err := types.ParseMinuteOfDay(time.Now().Format("15:04"))
		if err != nil {
			logging.StdOut.Error("run scheduler tick", "status", "failed", "reason", err)
			return
		}

		n := runSchedulerTick(currentMinute, s)
		if n > 0 {
			logging.StdOut.Info("scheduler", "scheduled workflows",
				n, "minute", currentMinute.String())
		}
	}
}

func runScheduleOnce(sched schedule.Schedule, wfName string) (int, error) {
	wf, err := sched.GetWorkflowByName(wfName)
	if err != nil {
		return ExitWorkflowNotFoundByName, err
	}

	err = executeWorkflow(wf)
	if err != nil {
		return ExitExecuteWorkflow, err
	}

	return ExitSuccess, nil
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
				logging.StdOut.Error("workflow", "status", types.WorkflowStatusFailed().String(),
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
		stdOut.Error("workflow", "status", types.WorkflowStatusSkipped().String(), "reason",
			"workflow already running", "existingRunID", existingID)
		return nil
	}
	defer schedule.RunningWorkflows.Delete(lockKey)

	stdOut.Info("workflow", "name", wf.Name, "status", "started")

	workflowStatus := types.WorkflowStatusCompleted()

	numSteps := len(wf.Steps)
	for i, step := range wf.Steps {
		result := runner.RunStepAttempt(stdOut, wf.Name, step, i)

		if result.Err != nil {
			if schedule.WorkflowAbortsOnFailure(wf) {
				stdOut.Info("step", "step", step.Name, "stepIndex", i, "status", wf.OnFailure)
				return fmt.Errorf("%w: %s", ErrRunCommandAbortsOnError, result.Err)
			}

			if schedule.WorkflowContinuesOnFailure(wf) {
				workflowStatus = types.WorkflowStatusPartial()
				continue
			}

			if schedule.WorkflowRetriesOnFailure(wf) {
				retryStatus := runner.RunStepRetries(stdOut, wf.Name, step, i, wf.Retry)
				if retryStatus == types.WorkflowStatusPartial() {
					workflowStatus = types.WorkflowStatusPartial()
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

func generateNewConfig() {
	newConfig := `
[
	{
		"name": "Workflow 1",
		"trigger": { "every": "1d", "beginAt": "10:00" },
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
		"trigger": { "every": "1d", "beginAt": "10:35" },
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
	},
	{
		"name": "Workflow 2",
		"trigger": { "every": "1d", "beginAt": "10:35" },
		"onFailure": "retry",
		"retry": {
			"numberRetries": 3,
			"pauseSeconds": 60
		},
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
