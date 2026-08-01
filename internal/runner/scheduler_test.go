package runner

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/helpers"
	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/workflow"
)

func TestExecuteWorkflowDisabledSkipsSteps(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	wf := workflow.Workflow{
		Name:           "disabled workflow is skipped",
		Enabled:        false,
		DisabledReason: "disabled for skip test",
		OnFailure:      workflow.Abort,
		Steps: []workflow.Step{
			{
				Name:    "must not run",
				Program: "/bin/sh",
				Args:    []string{"-c", "touch \"$1\"", "skip-script", marker},
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

func TestExecuteWorkflowAbortOnMissingProgram(t *testing.T) {
	wf := workflow.Workflow{
		Name:      "missing program aborts",
		Enabled:   true,
		OnFailure: workflow.Abort,
		Steps: []workflow.Step{
			{
				Name:    "missing",
				Program: filepath.Join(t.TempDir(), "does-not-exist"),
			},
			{
				Name:    "would continue",
				Program: filepath.Join(t.TempDir(), "also-does-not-exist"),
			},
		},
	}

	if err := executeWorkflow(wf); err == nil {
		t.Fatal("expected abort workflow to fail on missing program")
	}
}

func TestExecuteWorkflowAbortUsesPolicyValue(t *testing.T) {
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	wf := workflow.Workflow{
		Name:      "command failure aborts",
		Enabled:   true,
		OnFailure: workflow.Abort,
		Steps: []workflow.Step{
			{
				Name:    "exit 5",
				Program: testprog,
				Args: []string{
					"-sleep", "0",
					"-role", "exit 5",
					"-exit-code", "5",
				},
			},
		},
	}

	if err := executeWorkflow(wf); err == nil {
		t.Fatal("expected abort workflow to fail on command error")
	}
}

func TestExecuteWorkflowContinueFailureLogsPartial(t *testing.T) {
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	wf := workflow.Workflow{
		Name:      "continue failure is partial",
		Enabled:   true,
		OnFailure: workflow.Continue,
		Steps: []workflow.Step{
			{
				Name:    "missing",
				Program: filepath.Join(t.TempDir(), "does-not-exist"),
			},
			{
				Name:    "continues",
				Program: testprog,
				Args: []string{
					"-sleep", "0",
					"-role", "continues",
					"-exit-code", "0",
				},
			},
		},
	}

	out, err := executeWorkflowWithCapturedOutput(t, wf)
	if err != nil {
		t.Fatalf("expected continue workflow to finish: %v\n%s", err, out)
	}

	assertOutputContains(t, out, workflowStatusLog(workflow.StatusPartial))
	assertOutputContains(t, out, []byte(`role=continues`))
}

func TestExecuteWorkflowRetryExhaustionLogsPartial(t *testing.T) {
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	wf := workflow.Workflow{
		Name:      "retry exhaustion is partial",
		Enabled:   true,
		OnFailure: workflow.Retry,
		Retry: workflow.RetryPolicy{
			NumberRetries: 1,
		},
		Steps: []workflow.Step{
			{
				Name:    "exhausts retries",
				Program: testprog,
				Args: []string{
					"-sleep", "0",
					"-role", "exhausts-retries",
					"-exit-code", "5",
				},
			},
		},
	}

	out, err := executeWorkflowWithCapturedOutput(t, wf)
	if err != nil {
		t.Fatalf("expected retry workflow to finish: %v\n%s", err, out)
	}

	assertOutputContains(t, out, []byte(`status=exhausted`))
	assertOutputContains(t, out, workflowStatusLog(workflow.StatusPartial))
}

func TestExecuteWorkflowRetrySuccessLogsCompleted(t *testing.T) {
	dir := t.TempDir()
	markerFile := filepath.Join(dir, "retried")
	wf := workflow.Workflow{
		Name:      "retry success is completed",
		Enabled:   true,
		OnFailure: workflow.Retry,
		Retry: workflow.RetryPolicy{
			NumberRetries: 1,
		},
		Steps: []workflow.Step{
			{
				Name:    "succeeds on retry",
				Program: "/bin/sh",
				Args: []string{
					"-c",
					"if [ -f \"$1\" ]; then exit 0; fi; touch \"$1\"; exit 5",
					"retry-script",
					markerFile,
				},
			},
		},
	}

	out, err := executeWorkflowWithCapturedOutput(t, wf)
	if err != nil {
		t.Fatalf("expected retry workflow to finish: %v\n%s", err, out)
	}

	assertOutputContains(t, out, workflowStatusLog(workflow.StatusCompleted))
	if bytes.Contains(out, workflowStatusLog(workflow.StatusPartial)) {
		t.Fatalf("expected output not to contain partial status\n%s", out)
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
