package runner

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mytec33/gosched/internal/helpers"
	"github.com/mytec33/gosched/internal/workflow"
)

// TestRunStepCommandResults checks the command result contract directly:
// exit codes, error types, and captured output. Integration tests verify
// how workflow execution responds to those results.
func TestRunStepCommandResults(t *testing.T) {
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")

	timeout1s, err := workflow.ParseConfigDuration("1s")
	if err != nil {
		t.Fatalf("parse config duration 1s: %v", err)
	}
	timeout10s, err := workflow.ParseConfigDuration("10s")
	if err != nil {
		t.Fatalf("parse config duration 10s: %v", err)
	}

	tests := []struct {
		step         workflow.Step
		wantExitCode int
		wantErr      error
		wantExitErr  bool
		wantOutput   string
	}{
		{
			step: workflow.Step{
				Name:    "exit 0 succeeds",
				Program: testprog,
				Args: []string{
					"-sleep", "0", "-role", "exit 0 succeeds", "-exit-code", "0",
				},
			},
			wantExitCode: 0,
			wantOutput:   "EXIT  | role=exit 0 succeeds | 0",
		},
		{
			step: workflow.Step{
				Name:    "succeeds before timeout",
				Program: testprog,
				Args: []string{
					"-sleep", "0", "-role", "succeeds before timeout", "-exit-code", "0",
				},
				Timeout: timeout10s,
			},
			wantExitCode: 0,
			wantOutput:   "EXIT  | role=succeeds before timeout | 0",
		},
		{
			step: workflow.Step{
				Name:    "exit 5 fails",
				Program: testprog,
				Args: []string{
					"-sleep", "0", "-role", "exit 5 fails", "-exit-code", "5",
				},
			},
			wantExitCode: 5,
			wantExitErr:  true,
			wantOutput:   "EXIT  | role=exit 5 fails | 5",
		},
		{
			step: workflow.Step{
				Name:    "program not found",
				Program: filepath.Join(t.TempDir(), "does-not-exist"),
				Args: []string{
					"-sleep", "0", "-role", "program not found", "-exit-code", "5",
				},
			},
			wantExitCode: -1,
			wantErr:      os.ErrNotExist,
		},
		{
			step: workflow.Step{
				Name:    "timeout earlier than sleep time",
				Program: testprog,
				Args: []string{
					"-sleep", "2", "-role", "timeout earlier than sleep time", "-exit-code", "10",
				},
				Timeout: timeout1s,
			},
			wantExitCode: -1,
			wantErr:      context.DeadlineExceeded,
			wantOutput:   "START | role=timeout earlier than sleep time |",
		},
	}

	for _, tt := range tests {
		t.Run(tt.step.Name, func(t *testing.T) {
			result := runStepCommand(tt.step)

			if result.ExitCode != tt.wantExitCode {
				t.Errorf("command result must preserve exit code %d; got %d", tt.wantExitCode, result.ExitCode)
			}

			switch {
			case tt.wantExitErr:
				exitErr, ok := errors.AsType[*exec.ExitError](result.Err)
				if !ok {
					t.Errorf("nonzero command exit must return an exec.ExitError; got %T: %v", result.Err, result.Err)
				} else if exitErr.ExitCode() != tt.wantExitCode {
					t.Errorf("exit error must preserve child exit code %d; got %d", tt.wantExitCode, exitErr.ExitCode())
				}
			case tt.wantErr != nil:
				if !errors.Is(result.Err, tt.wantErr) {
					t.Errorf("command failure must identify %v; got %T: %v", tt.wantErr, result.Err, result.Err)
				}
			default:
				if result.Err != nil {
					t.Errorf("successful command must return a nil error; got %v", result.Err)
				}
			}

			if tt.wantOutput == "" {
				if len(result.Output) != 0 {
					t.Errorf("a program that could not start must not produce child output; got %q", result.Output)
				}
			} else if !bytes.Contains(result.Output, []byte(tt.wantOutput)) {
				t.Errorf("command result must retain child output, including output before a failure or timeout; missing %q in %q", tt.wantOutput, result.Output)
			}
		})
	}
}
