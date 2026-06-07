package main

import (
	"bytes"
	"log/slog"
	"path/filepath"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/helpers"
	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/policy"
	"git.sr.ht/~mytec/gosched/internal/schedule"
	"git.sr.ht/~mytec/gosched/internal/types"
)

func TestExecuteWorkflowAbortOnMissingProgram(t *testing.T) {
	wf := schedule.Workflow{
		Name:      "missing program aborts",
		OnFailure: policy.Abort,
		Steps: []schedule.Step{
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
	wf := schedule.Workflow{
		Name:      "command failure aborts",
		OnFailure: policy.Abort,
		Steps: []schedule.Step{
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
	wf := schedule.Workflow{
		Name:      "continue failure is partial",
		OnFailure: policy.Continue,
		Steps: []schedule.Step{
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

	assertOutputContains(t, out, workflowStatusLog(types.WorkflowStatusPartial))
	assertOutputContains(t, out, []byte(`role=continues`))
}

func TestExecuteWorkflowRetryExhaustionLogsPartial(t *testing.T) {
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	wf := schedule.Workflow{
		Name:      "retry exhaustion is partial",
		OnFailure: policy.Retry,
		Retry: &schedule.RetryPolicy{
			NumberRetries: 1,
		},
		Steps: []schedule.Step{
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
	assertOutputContains(t, out, workflowStatusLog(types.WorkflowStatusPartial))
}

func TestExecuteWorkflowRetrySuccessLogsCompleted(t *testing.T) {
	dir := t.TempDir()
	markerFile := filepath.Join(dir, "retried")
	wf := schedule.Workflow{
		Name:      "retry success is completed",
		OnFailure: policy.Retry,
		Retry: &schedule.RetryPolicy{
			NumberRetries: 1,
		},
		Steps: []schedule.Step{
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

	assertOutputContains(t, out, workflowStatusLog(types.WorkflowStatusCompleted))
	if bytes.Contains(out, workflowStatusLog(types.WorkflowStatusPartial)) {
		t.Fatalf("expected output not to contain partial status\n%s", out)
	}
}

func executeWorkflowWithCapturedOutput(t *testing.T, wf schedule.Workflow) ([]byte, error) {
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

func assertOutputContains(t *testing.T, out []byte, want []byte) {
	t.Helper()

	if !bytes.Contains(out, want) {
		t.Fatalf("expected output to contain %q\n%s", want, out)
	}
}

func workflowStatusLog(status types.WorkflowStatus) []byte {
	return []byte(`status=` + status.String())
}
