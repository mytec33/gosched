package main

import (
	"path/filepath"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/helpers"
	"git.sr.ht/~mytec/gosched/internal/policy"
	"git.sr.ht/~mytec/gosched/internal/schedule"
)

func TestExecuteWorkflowAbortOnMissingProgram(t *testing.T) {
	abort := policy.Abort
	wf := schedule.Workflow{
		Name:      "missing program aborts",
		OnFailure: &abort,
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
	abort := policy.Abort
	wf := schedule.Workflow{
		Name:      "command failure aborts",
		OnFailure: &abort,
		Steps: []schedule.Step{
			{
				Name:    "exit 5",
				Program: testprog,
				Args: []string{
					"-sleep", "0",
					"-role", "exit 5",
					"-exitCode", "5",
				},
			},
		},
	}

	if err := executeWorkflow(wf); err == nil {
		t.Fatal("expected abort workflow to fail on command error")
	}
}
