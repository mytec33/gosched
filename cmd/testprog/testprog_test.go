package main

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"
	"time"

	"git.sr.ht/~mytec/gosched/internal/helpers"
)

func assertTestProgramOutput(t *testing.T, out []byte, role string) {
	roleName := "role=" + role
	expected := [][]byte{
		[]byte(`| START |`),
		[]byte(`| DONE  |`),
		[]byte(`| EXIT  |`),
		[]byte(`| sleep=`),
		[]byte(`| elapsed=`),
		[]byte(roleName),
	}

	for _, want := range expected {
		if !bytes.Contains(out, want) {
			t.Fatalf("expected output to contain %q\n%s", want, out)
		}
	}
}

func TestSleepDuration(t *testing.T) {
	t.Parallel()

	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")

	tests := []struct {
		name         string
		sleepSecs    string
		minDuration  time.Duration
		maxDuration  time.Duration
		wantExitCode int
	}{
		{
			name:         "sleep 0 exits quickly",
			sleepSecs:    "0",
			minDuration:  0,
			maxDuration:  1 * time.Second,
			wantExitCode: 0,
		},
		{
			name:         "sleep 2 waits long enough",
			sleepSecs:    "2",
			minDuration:  2 * time.Second,
			maxDuration:  5 * time.Second,
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			start := time.Now()

			cmd := exec.Command(
				testprog,
				"-sleep", tt.sleepSecs,
				"-role", tt.name,
				"-exit-code", "0",
			)

			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("command failed: %v\n%s", err, out)
			}

			elapsed := time.Since(start)

			if elapsed < tt.minDuration {
				t.Fatalf(
					"elapsed time too short: want >= %v got %v", tt.minDuration, elapsed)
			}

			if elapsed > tt.maxDuration {
				t.Fatalf(
					"elapsed time too long: want <= %v got %v", tt.maxDuration, elapsed)
			}

			assertTestProgramOutput(t, out, tt.name)

			if cmd.ProcessState.ExitCode() != tt.wantExitCode {
				t.Fatalf("exit code: want %v, got %v", tt.wantExitCode, cmd.ProcessState.ExitCode())
			}
		})
	}
}

func TestRole(t *testing.T) {
	t.Parallel()

	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")

	// Single-word roles are exercised by every other test in this file; the
	// only input this test uniquely owns is a role containing spaces.
	tests := []struct {
		name         string
		sleepSecs    string
		role         string
		wantExitCode int
	}{
		{
			name:         "multi word role",
			sleepSecs:    "0",
			role:         "role with spaces",
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := exec.Command(
				testprog,
				"-sleep", tt.sleepSecs,
				"-role", tt.role,
				"-exit-code", "0",
			)

			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("command failed: %v\n%s", err, out)
			}

			assertTestProgramOutput(t, out, tt.role)

			if cmd.ProcessState.ExitCode() != tt.wantExitCode {
				t.Fatalf("exit code: want %v, got %v", tt.wantExitCode, cmd.ProcessState.ExitCode())
			}
		})
	}
}

func TestExit(t *testing.T) {
	t.Parallel()

	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")

	tests := []struct {
		name         string
		sleepSecs    string
		role         string
		exitCode     string
		wantExitCode int
	}{
		{
			name:         "exit code success",
			sleepSecs:    "0",
			role:         "test success exit code",
			exitCode:     "0",
			wantExitCode: 0,
		},
		{
			name:         "exit code non-zero",
			sleepSecs:    "0",
			role:         "test non-zero exit code",
			exitCode:     "10",
			wantExitCode: 10,
		},
		{
			name:         "exit code max",
			sleepSecs:    "0",
			role:         "test max exit code",
			exitCode:     "255",
			wantExitCode: 255,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := exec.Command(
				testprog,
				"-sleep", tt.sleepSecs,
				"-role", tt.role,
				"-exit-code", tt.exitCode,
			)

			out, err := cmd.CombinedOutput()
			if err != nil {
				var exitErr *exec.ExitError

				if !errors.As(err, &exitErr) {
					t.Fatalf("command failed: %v\n%s", err, out)
				}
			}

			assertTestProgramOutput(t, out, tt.role)

			if cmd.ProcessState.ExitCode() != tt.wantExitCode {
				t.Fatalf("exit code: want %v, got %v", tt.wantExitCode, cmd.ProcessState.ExitCode())
			}
		})
	}
}
