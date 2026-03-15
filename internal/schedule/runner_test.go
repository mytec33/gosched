package schedule

import (
	"testing"
	"time"

	"git.sr.ht/~mytec/gosched/internal/types"
)

func TestExitCode(t *testing.T) {
	testprog := buildBinary(t, "testprog", "cmd/testprog")

	tests := []struct {
		step         Step
		wantExitCode int
		wantFailed   bool
	}{
		{
			step: Step{
				Name:    "exit 0 succeeds",
				Program: testprog,
				Args: []types.ConfiguredArg{
					types.NewArg("-sleep"), types.NewArg("0"), types.NewArg("-role"), types.NewArg("exit 0 succeeds"),
					types.NewArg("-exitCode"), types.NewArg("0"),
				},
			},
			wantExitCode: 0,
			wantFailed:   false,
		},
		{
			step: Step{
				Name:    "exit 5 fails",
				Program: testprog,
				Args: []types.ConfiguredArg{
					types.NewArg("-sleep"), types.NewArg("0"), types.NewArg("-role"), types.NewArg("exit 5 fails"),
					types.NewArg("-exitCode"), types.NewArg("5"),
				},
			},
			wantExitCode: 5,
			wantFailed:   true,
		},
		{
			step: Step{
				Name:    "program not found",
				Program: "invalid_program_name",
				Args: []types.ConfiguredArg{
					types.NewArg("-sleep"), types.NewArg("0"), types.NewArg("-role"), types.NewArg("program not found"),
					types.NewArg("-exitCode"), types.NewArg("5"),
				},
			},
			wantExitCode: -1,
			wantFailed:   true,
		},
		{
			step: Step{
				Name:    "timeout earlier than sleep time",
				Program: testprog,
				Args: []types.ConfiguredArg{
					types.NewArg("-sleep"), types.NewArg("2"), types.NewArg("-role"), types.NewArg("timeout earlier than sleep time"),
					types.NewArg("-exitCode"), types.NewArg("10"),
				},
				Timeout: types.NewConfiguredInt(1),
			},
			wantExitCode: -1,
			wantFailed:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.step.Name, func(t *testing.T) {
			result := RunStepCommand(tt.step)

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

func TestTimeout(t *testing.T) {
	testprog := buildBinary(t, "testprog", "cmd/testprog")

	tests := []struct {
		step       Step
		wantFailed bool
	}{
		{
			step: Step{
				Name:    "timeout earlier than sleep time",
				Program: testprog,
				Args: []types.ConfiguredArg{
					types.NewArg("-sleep"), types.NewArg("2"), types.NewArg("-role"), types.NewArg("timeout earlier than sleep time"),
					types.NewArg("-exitCode"), types.NewArg("10"),
				},
				Timeout: types.NewConfiguredInt(1),
			},
			wantFailed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.step.Name, func(t *testing.T) {
			result := RunStepCommand(tt.step)

			if result.Failed() != tt.wantFailed {
				t.Fatalf("%s: failure check failed: want %v got failure %v", tt.step.Name,
					tt.wantFailed, result.Failed())
			}
		})
	}
}

func TestDuration(t *testing.T) {
	testprog := buildBinary(t, "testprog", "cmd/testprog")

	step := Step{
		Name:    "duration within reasonable time",
		Program: testprog,
		Args: []types.ConfiguredArg{
			types.NewArg("-sleep"), types.NewArg("1"), types.NewArg("-role"), types.NewArg("duration within reasonable time"),
			types.NewArg("-exitCode"), types.NewArg("10"),
		},
	}

	start := time.Now()
	_ = RunStepCommand(step)
	elapsed := time.Since(start)

	if elapsed > 2000*time.Millisecond {
		t.Fatalf("timeout did not trigger quickly enough, duration %q", elapsed)
	}
}
