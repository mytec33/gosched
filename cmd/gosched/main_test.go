package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

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
	testprog := buildTestBinary(t, "testprog", "cmd/testprog")
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

func buildTestBinary(t *testing.T, name, rel string) string {
	t.Helper()

	root := findModuleRoot(t)
	dir := t.TempDir()
	bin := filepath.Join(dir, name)
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", bin, filepath.Join(root, rel))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	return bin
}

func findModuleRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	for {
		_, err := os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
