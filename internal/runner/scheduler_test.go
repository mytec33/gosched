package runner

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/mytec33/gosched/internal/helpers"
	"github.com/mytec33/gosched/internal/logging"
	"github.com/mytec33/gosched/internal/schedule"
	"github.com/mytec33/gosched/internal/workflow"
)

func TestRunSchedulePrecondition(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assertPanic(t, "runner.RunSchedule: precondition failed: invalid Schedule", func() {
		_ = RunSchedule(ctx, schedule.Schedule{})
	})
}

func TestRunScheduleOncePrecondition(t *testing.T) {
	assertPanic(t, "runner.RunScheduleOnce: precondition failed: invalid Schedule", func() {
		_ = RunScheduleOnce(schedule.Schedule{}, "Workflow 1")
	})
}

func assertPanic(t *testing.T, want string, f func()) {
	t.Helper()

	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("expected panic")
		}

		if got != want {
			t.Fatalf("panic = %q, want %q", got, want)
		}
	}()

	f()
}

func TestExecuteWorkflowDisabledSkipsSteps(t *testing.T) {
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	marker := filepath.Join(t.TempDir(), "ran")
	wf := workflow.Workflow{
		Name:           "disabled workflow is skipped",
		Enabled:        false,
		DisabledReason: "disabled for skip test",
		OnFailure:      workflow.Abort,
		Steps: []workflow.Step{
			{
				Name:    "must not run",
				Program: testprog,
				Args: []string{
					"-sleep", "0",
					"-role", "must not run",
					"-exit-code", "5",
					"-fail-once-marker", marker,
				},
			},
		},
	}

	out, err := executeWorkflowWithCapturedOutput(t, wf)
	if err != nil {
		t.Fatalf("expected disabled workflow to be skipped without error: %v\n%s", err, out)
	}

	assertOutputContains(t, out, workflowStatusLog(workflow.StatusSkipped))
	assertOutputContains(t, out, []byte(`disabledReason="disabled for skip test"`))

	if _, statErr := os.Stat(marker); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected step not to run, but marker file check got: %v", statErr)
	}
}

func TestExecuteWorkflowRejectsUnknownFailurePolicy(t *testing.T) {
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	marker := filepath.Join(t.TempDir(), "later-step-ran")
	// JSON parsing rejects this policy. Supply it directly to exercise the
	// execution boundary's defensive handling of an invalid internal value.
	wf := workflow.Workflow{
		Name:      "Unknown failure policy",
		Enabled:   true,
		OnFailure: workflow.FailureMode{},
		Steps: []workflow.Step{
			{
				Name:    "Cannot start",
				Program: filepath.Join(t.TempDir(), "does-not-exist"),
			},
			{
				Name:    "Must not run",
				Program: testprog,
				Args: []string{
					"--role", "must-not-run",
					"--exit-code", "5",
					"--fail-once-marker", marker,
				},
			},
		},
	}
	out, err := executeWorkflowWithCapturedOutput(t, wf)
	if !errors.Is(err, ErrRunStepUnknownOnFailure) {
		t.Fatalf("unsupported failure policy must stop the workflow with ErrRunStepUnknownOnFailure; got %v\n%s", err, out)
	}
	assertIntegrationOutput(t, out, `reason="reached an unexpected branch on run step. Contact developer."`, "unsupported failure policy must produce a developer-contact diagnostic")
	assertIntegrationOutput(t, out, `"failure mode"=""`, "the diagnostic must identify the unsupported empty failure policy")
	assertWorkflowSummary(t, out, wf.Name, "failed")

	_, statErr := os.Stat(marker)
	if !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("unsupported failure policy must prevent the following step from running; marker check returned %v\n%s", statErr, out)
	}
	if bytes.Contains(out, []byte(`stepIndex=1`)) {
		t.Fatalf("unsupported failure policy must return before attempting the following step\n%s", out)
	}
}

func assertOutputContains(t *testing.T, out []byte, want []byte) {
	t.Helper()

	if !bytes.Contains(out, want) {
		t.Fatalf("expected output to contain %q\n%s", want, out)
	}
}

func workflowStatusLog(status workflow.WorkflowStatus) []byte {
	return []byte(`status=` + status.String())
}

func executeWorkflowWithCapturedOutput(t *testing.T, wf workflow.Workflow) ([]byte, error) {
	t.Helper()

	var buf bytes.Buffer
	original := logging.StdOut
	logging.StdOut = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	t.Cleanup(func() {
		logging.StdOut = original
	})

	err := executeWorkflow(wf)
	return buf.Bytes(), err
}
