package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/workflow"
)

const (
	ExitNoConfig     int = 1
	ExitDecodeConfig int = 2
	ExitValidation   int = 3
	ExitInvalidArgs  int = 4
)

var RunningWorkflows = workflow.NewSafeMapMutex()

func main() {
	os.Exit(run())
}

func run() int {
	logging.StdoutLogger.Info("startup", "Scheduler service started", "")

	filename := ""
	flag.StringVar(&filename, "schedule", "", "file containing a schedule to run")
	flag.Parse()

	if filename == "" {
		logging.StderrLogger.Error("invalid args")
		flag.Usage()
		os.Exit(ExitInvalidArgs)
	}

	data, err := loadWorkflows(filename)
	if err != nil {
		logging.StderrLogger.Error("startup", "reason", "failed to load schedule", "error", err)
		return ExitNoConfig
	}

	workflows, err := decodeWorkflows(filename, data)
	if err != nil {
		logging.StderrLogger.Error("startup", "reason", "failed to decode schedule", "error", err)
		return ExitDecodeConfig
	}

	if errs := workflow.ValidateAll(workflows); len(errs) > 0 {
		for _, e := range errs {
			logging.StderrLogger.Error("startup", "configuration error", e)
		}
		return ExitValidation
	}

	logging.StdoutLogger.Info("startup", "Workflows loaded", len(workflows))

	// Long-running scheduler loop (effects)
	runScheduler(workflows)

	return 0
}

func runScheduler(workflows []workflow.Workflow) {
	// Align to the next minute boundary once, then tick.
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for t := range ticker.C {
		n := runSchedulerTick(t, workflows)
		if n > 0 {
			logging.StdoutLogger.Info("scheduler", "scheduled workflows",
				n, "minute", t.Format("15:04"),
			)
		}
	}
}

func runSchedulerTick(now time.Time, workflows []workflow.Workflow) int {
	currentMinute := now.Format("15:04")

	count := 0
	for _, wf := range workflows {
		if wf.Time == currentMinute {
			count++
			go executeWorkflow(wf)
		}
	}
	return count
}

func loadWorkflows(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading workflows file %q: %w", filename, err)
	}

	return data, nil
}

func decodeWorkflows(filename string, data []byte) ([]workflow.Workflow, error) {
	var schedule []workflow.Workflow
	err := json.Unmarshal(data, &schedule)
	if err != nil {
		return nil, fmt.Errorf("error parsing %q: %w", filename, err)
	}

	return schedule, nil
}

func executeWorkflow(wf workflow.Workflow) {
	workflowStart := time.Now()

	lockKey := wf.Name
	_, running := RunningWorkflows.Get(lockKey)
	if running {
		logging.StderrLogger.Error("execute", "workflow already running", lockKey)
		return
	}
	RunningWorkflows.Set(lockKey, "running")
	defer RunningWorkflows.Delete(lockKey)

	for i, step := range wf.Steps {
		stepStart := time.Now()

		args := strings.Fields(step.Args)
		logging.StdoutLogger.Info("execute", wf.Name, "starting", "stepIndex", i, "stepName", step.Name, "args", step.Args)

		cmd := exec.Command(step.Program, args...)
		output, err := cmd.CombinedOutput()
		stepDuration := time.Since(stepStart)

		if err != nil {
			logging.StderrLogger.Error("execute", wf.Name, "failed", "stepIndex", i, "stepName", step.Name, "args", step.Args, "duration", stepDuration, "reason", err, "output", string(output))
			return
		} else {
			logging.StdoutLogger.Info("execute", wf.Name, "completed", "stepIndex", i, "stepName", step.Name, "args", step.Args, "duration", stepDuration)
		}
	}

	workflowDuration := time.Since(workflowStart)
	logging.StdoutLogger.Info("execute", wf.Name, "completed", "duration", workflowDuration)
}
