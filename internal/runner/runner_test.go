package runner

import (
	"testing"

	"github.com/mytec33/gosched/internal/helpers"
	"github.com/mytec33/gosched/internal/workflow"
)

func TestExitCode(t *testing.T) {
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")

	timeout1s, err := workflow.ParseConfigDuration("1s")
	if err != nil {
		t.Fatalf("parse config duration 1s: %v", err)
	}

	tests := []struct {
		step         workflow.Step
		wantExitCode int
		wantFailed   bool
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
			wantFailed:   false,
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
			wantFailed:   true,
		},
		{
			step: workflow.Step{
				Name:    "program not found",
				Program: "invalid_program_name",
				Args: []string{
					"-sleep", "0", "-role", "program not found", "-exit-code", "5",
				},
			},
			wantExitCode: -1,
			wantFailed:   true,
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
			wantFailed:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.step.Name, func(t *testing.T) {
			result := runStepCommand(tt.step)

			if result.ExitCode != tt.wantExitCode {
				t.Fatalf("%s: want %v, got %v", tt.step.Name, tt.wantExitCode, result.ExitCode)
			}

			if result.Failed() != tt.wantFailed {
				t.Fatalf("%s: failure check failed: want %v got failure %v", tt.step.Name,
					tt.wantFailed, result.Failed())
			}
		})
	}
}
