// Package runner provides methods to run workflow steps
package runner

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"
	"time"

	"git.sr.ht/~mytec/gosched/internal/helpers"
	"git.sr.ht/~mytec/gosched/internal/schedule"
)

type StepExecutionResult struct {
	Output   []byte
	Err      error
	ExitCode int
}

func RunStepCommand(step schedule.Step) StepExecutionResult {
	var stepResult StepExecutionResult

	var cmd *exec.Cmd
	var cancel context.CancelFunc
	var ctx context.Context

	if step.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), helpers.SecondsDuration(step.Timeout))
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

func RunStepAttempt(stdOut *slog.Logger, step schedule.Step, index int) StepExecutionResult {
	stdOut.Info("step", "status", "started", "stepIndex", index, "stepName", step.Name, "args", step.Args)
	stepStart := time.Now()

	result := RunStepCommand(step)
	stepDuration := time.Since(stepStart)

	if len(result.Output) != 0 {
		stdOut.Info("step", "output", result.Output)
	}

	if result.Err != nil {
		stdOut.Error("step", "step", step.Name, "stepIndex", index, "status", "failed",
			"exitCode", result.ExitCode, "duration", stepDuration, "reason", result.Err)
	} else {
		stdOut.Info("step", "step", step.Name, "stepIndex", index, "status", "completed",
			"exitCode", result.ExitCode, "duration", stepDuration)
	}

	return result
}

func RunStepRetries(stdOut *slog.Logger, step schedule.Step, stepIndex int, retry *schedule.RetryPolicy) {
	var retryResult StepExecutionResult

	for attempt := 1; attempt <= retry.NumberRetries; attempt++ {
		retryPause := time.Duration(retry.PauseSeconds) * time.Second

		stdOut.Info("step retry", "stepName", step.Name, "stepIndex", stepIndex,
			"retryAttempt", attempt, "maxRetries", retry.NumberRetries,
			"status", "paused", "duration", retryPause,
		)

		time.Sleep(retryPause)

		retryResult = RunStepAttempt(stdOut, step, stepIndex)
		if retryResult.Err == nil {
			return
		}
	}

	stdOut.Error("step retry", "stepName", step.Name, "stepIndex", stepIndex,
		"status", "exhausted", "retries", retry.NumberRetries)
}

func (s StepExecutionResult) Failed() bool {
	return s.Err != nil || s.ExitCode != 0
}
