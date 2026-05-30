package runner

import (
	"testing"
	"time"

	"git.sr.ht/~mytec/gosched/internal/helpers"
	"git.sr.ht/~mytec/gosched/internal/schedule"
)

func TestExitCode(t *testing.T) {
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")

	tests := []struct {
		step         schedule.Step
		wantExitCode int
		wantFailed   bool
	}{
		{
			step: schedule.Step{
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
			step: schedule.Step{
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
			step: schedule.Step{
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
			step: schedule.Step{
				Name:    "timeout earlier than sleep time",
				Program: testprog,
				Args: []string{
					"-sleep", "2", "-role", "timeout earlier than sleep time", "-exit-code", "10",
				},
				Timeout: 1,
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
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")

	tests := []struct {
		step       schedule.Step
		wantFailed bool
	}{
		{
			step: schedule.Step{
				Name:    "timeout earlier than sleep time",
				Program: testprog,
				Args: []string{
					"-sleep", "2", "-role", "timeout earlier than sleep time", "-exit-code", "10",
				},
				Timeout: 1,
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
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")

	step := schedule.Step{
		Name:    "duration within reasonable time",
		Program: testprog,
		Args: []string{
			"-sleep", "1", "-role", "duration within reasonable time", "-exit-code", "10",
		},
	}

	start := time.Now()
	_ = RunStepCommand(step)
	elapsed := time.Since(start)

	if elapsed > 2000*time.Millisecond {
		t.Fatalf("timeout did not trigger quickly enough, duration %q", elapsed)
	}
}
