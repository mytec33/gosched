package schedule

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// All tests in this file are testing valid workflows to ensure we can run what we say we can run

const ScheduleOneWorkflowOneStep = `
[
  {
    "name": "Workflow 1",
    "time": "%s",
    "steps": [
      {
        "name": "daily",
        "program": %q,
        "args": "--sleep 1 --role daily-slot-ratings"
      }
    ]
  }
]
`

func buildBinary(t *testing.T, name, rel string) string {
	t.Helper()

	root := findModuleRoot(t)

	dir := t.TempDir()
	bin := filepath.Join(dir, name)

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

func TestOneWorkFlowOneStep(t *testing.T) {
	t.Parallel()

	scheduler := buildBinary(t, "gosched", "cmd/gosched")
	testprog := buildBinary(t, "testprog", "cmd/testprog")
	schedule := fmt.Sprintf(ScheduleOneWorkflowOneStep, time.Now().Format("15:04"), testprog)

	dir := t.TempDir()
	file := filepath.Join(dir, "schedule.json")
	if err := os.WriteFile(file, []byte(schedule), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, scheduler, "-schedule", file, "-run-once")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scheduler failed: %v\n%s", err, out)
	}

	fmt.Printf("output: %s\n", out)
}
