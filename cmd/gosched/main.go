package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/schedule"
)

const (
	ExitSuccess      int = 0
	ExitNoConfig     int = 1
	ExitDecodeConfig int = 2
	ExitValidation   int = 3
	ExitInvalidArgs  int = 4
	ExitInvalidLock  int = 5
)

type TickFunc func(now time.Time, data schedule.Schedule) int

func main() {
	os.Exit(run())
}

func run() int {
	logging.StdOut.Info("startup", "reason", "scheduler service started")

	filename := ""
	summarizeConfig := false
	flag.StringVar(&filename, "schedule", "", "file containing a schedule to run")
	flag.BoolVar(&summarizeConfig, "summarize-config", false, "show concise summary of configuration file schedule")
	flag.Parse()

	if filename == "" {
		logging.StdErr.Error("startup", "reason", "invalid args")
		flag.Usage()
		return ExitInvalidArgs
	}

	schedule, valErrors, err := schedule.ReadScheduleFile(filename)
	if err != nil {
		logging.StdErr.Error("startup", "reason", "failed to load schedule", "error", err)
		return ExitNoConfig
	}

	if len(valErrors) > 0 {
		displayErrors(valErrors)
		return ExitValidation
	}

	if summarizeConfig {
		displayConfigSummarization()
		return ExitSuccess
	}

	logging.StdOut.Info("startup", "reason", "workflows loaded", "count", schedule.WorkflowCount(), "filename", filename)

	runScheduler(runSchedulerTick, schedule)

	return 0
}

func displayConfigSummarization() {
	fmt.Println("display configuration summarization")
}

func displayErrors(errors []error) {
	fmt.Println("configuration invalid, errors found:")
	for _, err := range errors {
		fmt.Println(err)
	}
}

func runScheduler(tick TickFunc, s schedule.Schedule) {
	// Align to the next minute boundary once, then tick.
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().Truncate(time.Minute)
		n := tick(now, s)
		if n > 0 {
			logging.StdOut.Info("scheduler", "scheduled workflows",
				n, "minute", now.Format("15:04"),
			)
		}
	}
}

func runSchedulerTick(now time.Time, s schedule.Schedule) int {
	currentMinute := schedule.MinuteKey(now.Format("15:04"))
	logging.StdOut.Info("run scheduler tick", "current_minute", currentMinute)

	tasks := s.WorkflowsAtMinute(currentMinute)
	if len(tasks) == 0 {
		return 0
	}

	for _, task := range tasks {
		go executeWorkflow(task)
	}
	return len(tasks)
}

func executeWorkflow(wf schedule.Workflow) {
	workflowStart := time.Now()

	wfLog := logging.NewWorkflowLogger(wf.Name)
	stdOut := wfLog.Out
	stdErr := wfLog.Err

	lockKey := wf.Name
	existingID, running := schedule.RunningWorkflows.Get(lockKey)
	if running {
		stdErr.Error("workflow", "status", "skipped", "reason", "workflow already running", "existingWfRunID", existingID)
		return
	}
	schedule.RunningWorkflows.Set(lockKey, wfLog.WfRunID)
	defer schedule.RunningWorkflows.Delete(lockKey)

	stdOut.Info("workflow", "status", "started")

	for i, step := range wf.Steps {
		stepStart := time.Now()

		args := strings.Fields(step.Args)
		stdOut.Info("workflow step", "status", "started", "stepIndex", i, "stepName", step.Name, "args", step.Args)

		cmd := exec.Command(step.Program, args...)
		output, err := cmd.CombinedOutput()
		stepDuration := time.Since(stepStart)

		if err != nil {
			stdErr.Error("workflow step", "status", "failed", "stepIndex", i, "stepName", step.Name, "duration", stepDuration, "reason", err, "output", string(output))
			return
		} else {
			stdOut.Info("workflow step", "status", "completed", "stepIndex", i, "stepName", step.Name, "duration", stepDuration)
		}
	}

	workflowDuration := time.Since(workflowStart)
	stdOut.Info("workflow", "status", "completed", "duration", workflowDuration)
}
