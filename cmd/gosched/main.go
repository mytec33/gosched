package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/workflow"
)

const (
	ExitNoConfig  int = 1
	ExitBadConfig int = 2
)

var RunningWorkflows = workflow.NewSafeMapMutex()

func main() {
	logging.StdoutLogger.Info("startup", "Scheduler service started", "")

	workflows, err := loadWorkflows("schedule.json")
	if err != nil {
		logging.StderrLogger.Error("startup", "reason", "failed to load schedule", "error", err)
		os.Exit(ExitNoConfig)
	}

	errs := workflow.ValidateAll(workflows)
	if len(errs) > 0 {
		for _, e := range errs {
			logging.StderrLogger.Error("startup", "configuration error", e)
		}
		os.Exit(ExitBadConfig)
	}

	logging.StderrLogger.Error("startup", "Workflows loaded", len(workflows))

	lastMinute := ""

	for {
		now := time.Now()
		currentMinute := now.Format("15:04")

		if currentMinute != lastMinute {
			workflowsThisMinute := 0
			for _, wf := range workflows {
				if wf.Time == currentMinute {
					workflowsThisMinute++
					go executeWorkflow(wf)
				}
			}

			if workflowsThisMinute > 0 {
				logging.StderrLogger.Error("scheduler", "scheduled workflows", workflowsThisMinute, "minute", currentMinute)
			}

			lastMinute = currentMinute
		}

		nextMinute := now.Truncate(time.Minute).Add(time.Minute)
		time.Sleep(time.Until(nextMinute))
	}
}

func loadWorkflows(filename string) ([]workflow.Workflow, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var schedule []workflow.Workflow
	err = json.Unmarshal(data, &schedule)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
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
