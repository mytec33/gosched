// Package runner provides methods to run workflow steps
package runner

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"
	"time"

	"git.sr.ht/~mytec/gosched/internal/workflow"
)

type StepExecutionResult struct {
	Output   []byte
	Err      error
	ExitCode int
}

func runStepCommand(step workflow.Step) StepExecutionResult {
	var stepResult StepExecutionResult

	var cmd *exec.Cmd
	var cancel context.CancelFunc
	var ctx context.Context

	if step.Timeout.Duration() > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), step.Timeout.Duration())
		cmd = exec.CommandContext(ctx, step.Program, step.Args...)
	} else {
		cmd = exec.Command(step.Program, step.Args...)
	}
	output, err := cmd.CombinedOutput()
	stepResult.Err = err
	stepResult.Output = output

	if cancel != nil {
		defer cancel()
	}

	if ctx != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		stepResult.ExitCode = -1
		stepResult.Err = ctx.Err()
		return stepResult
	}

	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	stepResult.ExitCode = exitCode

	return stepResult
}

func RunStepAttempt(stdOut *slog.Logger, wfName string, step workflow.Step, index int) StepExecutionResult {
	stdOut.Info("step", "status", "started", "workflow", wfName, "stepIndex", index, "stepName", step.Name,
		"args", step.Args)
	stepStart := time.Now()

	stepResult := runStepCommand(step)
	stepDuration := time.Since(stepStart)

	if len(stepResult.Output) != 0 {
		stdOut.Info("step", "output", stepResult.Output)
	}

	if stepResult.Err != nil {
		stdOut.Error("step", "workflow", wfName, "step", step.Name, "stepIndex", index, "status", "failed",
			"exitCode", stepResult.ExitCode, "duration", stepDuration, "reason", stepResult.Err)
	} else {
		stdOut.Info("step", "workflow", wfName, "step", step.Name, "stepIndex", index, "status", "completed",
			"exitCode", stepResult.ExitCode, "duration", stepDuration)
	}

	return stepResult
}

func RunStepRetries(stdOut *slog.Logger, wfName string, step workflow.Step,
	stepIndex int, retry workflow.RetryPolicy) workflow.WorkflowStatus {
	var retryResult StepExecutionResult

	for attempt := 1; attempt <= retry.NumberRetries; attempt++ {
		retryPause := time.Duration(retry.PauseSeconds) * time.Second

		stdOut.Info("step retry", "stepName", step.Name, "stepIndex", stepIndex,
			"retryAttempt", attempt, "maxRetries", retry.NumberRetries,
			"status", "paused", "duration", retryPause,
		)

		time.Sleep(retryPause)

		retryResult = RunStepAttempt(stdOut, wfName, step, stepIndex)
		if retryResult.Err == nil {
			return workflow.StatusCompleted
		}
	}

	stdOut.Error("step retry", "stepName", step.Name, "stepIndex", stepIndex,
		"status", "exhausted", "retries", retry.NumberRetries)

	return workflow.StatusPartial
}

// Failed receiver function only used in tests
func (s StepExecutionResult) Failed() bool {
	return s.Err != nil || s.ExitCode != 0
}
