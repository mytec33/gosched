package schedule

import (
	"context"
	"os/exec"
	"time"
)

type StepExecutionResult struct {
	Output   []byte
	Err      error
	ExitCode int
}

func RunStepCommand(step Step) StepExecutionResult {
	var stepResult StepExecutionResult

	var cmd *exec.Cmd
	var cancel context.CancelFunc

	if step.Timeout > 0 {
		ctx, c := context.WithTimeout(context.Background(), secondsDuration(step.Timeout))
		cancel = c
		cmd = exec.CommandContext(ctx, step.Program, step.Args...)
	} else {
		cmd = exec.Command(step.Program, step.Args...)
	}
	output, err := cmd.CombinedOutput()
	stepResult.Err = err
	stepResult.Output = output

	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	stepResult.ExitCode = exitCode

	if cancel != nil {
		cancel()
	}

	return stepResult
}

func (s StepExecutionResult) Failed() bool {
	return s.Err != nil || s.ExitCode != 0
}

func secondsDuration(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}
