package schedule

import (
	"testing"

	"git.sr.ht/~mytec/gosched/internal/types"
)

const ExitSuccess = `
[
  {
    "name": "Workflow 1",
    "time": "%s",
	"onFailure": "continue",
    "steps": [
      {
        "name": "daily",
        "program": %q,
        "args": ["--sleep", "1", "--role", "daily-slot-ratings"]
      }
    ]
  }
]
`

func TestExitCode_Valid(t *testing.T) {
	testprog := buildBinary(t, "testprog", "cmd/testprog")

	tests := []struct {
		step         Step
		wantExitCode int
	}{
		{
			step: Step{
				Name:    "exit 0 succeeds",
				Program: testprog,
				Args: []types.ConfiguredArg{
					types.NewArg("-sleep"), types.NewArg("1"), types.NewArg("-role"), types.NewArg("exit 0"),
					types.NewArg("-exitCode"), types.NewArg("0"),
				},
			},
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.step.Name, func(t *testing.T) {
			result := RunStepCommand(tt.step)

			if result.ExitCode != tt.wantExitCode {
				t.Fatalf("%s: want %v, got %v", tt.step.Name, tt.wantExitCode, result.ExitCode)
			}

			if result.Failed() {
				t.Fatalf("%s: want no failure, got failure %v %v", tt.step.Name,
					result.Err, result.Output)
			}
		})
	}
}
