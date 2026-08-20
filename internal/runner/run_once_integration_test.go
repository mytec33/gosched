package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"git.sr.ht/~mytec/gosched/internal/helpers"
)

const runOnceOneWorkflowOneStep = `
[
  {
    "name": "Workflow 1",
	"enabled": true,
    "trigger": { "every": "1d", "beginAt": "%s" },
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

const runOnceTwoWorkflowsOneStep = `
[
  {
    "name": "Workflow 1",
	"enabled": true,
    "trigger": { "every": "1d", "beginAt": "%s" },
	"onFailure": "continue",	
    "steps": [
      {
        "name": "daily",
        "program": %q,
        "args": ["--sleep", "1", "--role", "workflow-1-step-1"]
      }
    ]
  },
  {
    "name": "Workflow 2",
	"enabled": true,
    "trigger": { "every": "1d", "beginAt": "%s" },
    "onFailure": "continue",	
    "steps": [
      {
        "name": "daily",
        "program": %q,
        "args": ["--sleep", "2", "--role", "workflow-2-step-1"]
      }
    ]
  }
]
`

func TestRunOnceOneWorkflowOneStep(t *testing.T) {
	t.Parallel()

	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	schedule := fmt.Sprintf(runOnceOneWorkflowOneStep, time.Now().Format("15:04"), testprog)

	manifestFile := writeRunOnceManifest(t, schedule)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Workflow 1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scheduler failed: %v\n%s", err, out)
	}

	expected := [][]byte{
		[]byte(`fileCount=1`),
		[]byte(`workflowCount=1`),
		[]byte(`stepCount=1`),
		[]byte(`reason="run once started"`),
		[]byte(`workflow="Workflow 1"`),
		[]byte(`msg=workflow`),
		[]byte(`status=completed`),
		[]byte(`role=daily-slot-ratings`),
	}

	for _, want := range expected {
		if !bytes.Contains(out, want) {
			t.Fatalf("expected output to contain %q\n%s", want, out)
		}
	}
}

func TestRunOnceWorkflowNotFound(t *testing.T) {
	t.Parallel()

	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	schedule := fmt.Sprintf(runOnceOneWorkflowOneStep, time.Now().Format("15:04"), testprog)

	manifestFile := writeRunOnceManifest(t, schedule)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Workflow 11")

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected scheduler to fail\n%s", out)
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exec.ExitError, got %T: %v\n%s", err, err, out)
	}

	if exitErr.ExitCode() != 9 {
		t.Fatalf("expected exit code 9, got %d\n%s", exitErr.ExitCode(), out)
	}

	expected := [][]byte{
		[]byte(`fileCount=1`),
		[]byte(`reason="run once started"`),
		[]byte(`workflow="Workflow 11"`),
		[]byte(`workflow not found by name: Workflow 11`),
	}

	for _, want := range expected {
		if !bytes.Contains(out, want) {
			t.Fatalf("expected output to contain %q\n%s", want, out)
		}
	}
}

func TestRunOnceTwoWorkflowsOneStep(t *testing.T) {
	t.Parallel()

	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")

	now := time.Now().Format("15:04")
	schedule := fmt.Sprintf(runOnceTwoWorkflowsOneStep, now, testprog, now, testprog)

	manifestFile := writeRunOnceManifest(t, schedule)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Workflow 2")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scheduler failed: %v\n%s", err, out)
	}

	expected := [][]byte{
		[]byte(`reason="scheduler service started"`),
		[]byte(`reason="run once started"`),
		[]byte(`workflow="Workflow 2"`),
		[]byte(`msg=workflow`),
		[]byte(`status=completed`),
		[]byte(`role=workflow-2-step-1`),
	}

	for _, want := range expected {
		if !bytes.Contains(out, want) {
			t.Fatalf("expected output to contain %q\n%s", want, out)
		}
	}

	if bytes.Contains(out, []byte(`role=workflow-1-step-1`)) {
		t.Fatalf("unexpected Workflow 1 execution\n%s", out)
	}
}

func writeRunOnceManifest(t *testing.T, schedule string) string {
	t.Helper()

	dir := t.TempDir()
	scheduleFile := filepath.Join(dir, "schedule.json")
	if err := os.WriteFile(scheduleFile, []byte(schedule), 0o600); err != nil {
		t.Fatal(err)
	}

	manifestFile := filepath.Join(dir, "manifest.txt")
	if err := os.WriteFile(manifestFile, []byte(scheduleFile), 0o600); err != nil {
		t.Fatal(err)
	}

	return manifestFile
}
