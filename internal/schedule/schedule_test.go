package schedule

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
	"git.sr.ht/~mytec/gosched/internal/policy"
	"git.sr.ht/~mytec/gosched/internal/types"
)

// All tests in this file are testing valid workflows to ensure we can run what we say we can run

const ScheduleOneWorkflowOneStep = `
[
  {
    "name": "Workflow 1",
    "time": "%s",
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

const ScheduleTwoWorkflowOneStep = `
[
  {
    "name": "Workflow 1",
    "time": "%s",
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
    "time": "%s",
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

func TestOneWorkFlowOneStep(t *testing.T) {
	t.Parallel()

	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	schedule := fmt.Sprintf(ScheduleOneWorkflowOneStep, time.Now().Format("15:04"), testprog)

	dir := t.TempDir()
	scheduleFile := filepath.Join(dir, "schedule.json")
	if err := os.WriteFile(scheduleFile, []byte(schedule), 0o600); err != nil {
		t.Fatal(err)
	}

	manifestFile := filepath.Join(dir, "manifest.txt")
	if err := os.WriteFile(manifestFile, []byte(scheduleFile), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Workflow 1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scheduler failed: %v\n%s", err, out)
	}

	expected := [][]byte{
		[]byte(`reason="run once started"`),
		[]byte(`workFlow="Workflow 1"`),
		[]byte(`msg=workflow`),
		[]byte(`name="Workflow 1"`),
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
	schedule := fmt.Sprintf(ScheduleOneWorkflowOneStep, time.Now().Format("15:04"), testprog)

	dir := t.TempDir()
	scheduleFile := filepath.Join(dir, "schedule.json")
	if err := os.WriteFile(scheduleFile, []byte(schedule), 0o600); err != nil {
		t.Fatal(err)
	}

	manifestFile := filepath.Join(dir, "manifest.txt")
	if err := os.WriteFile(manifestFile, []byte(scheduleFile), 0o600); err != nil {
		t.Fatal(err)
	}

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
		[]byte(`workflows loaded" count=1`),
		[]byte(`reason="run once started"`),
		[]byte(`workFlow="Workflow 11"`),
		[]byte(`workflow not found by name: Workflow 11`),
	}

	for _, want := range expected {
		if !bytes.Contains(out, want) {
			t.Fatalf("expected output to contain %q\n%s", want, out)
		}
	}
}

func TestTwoWorkFlowOneStep(t *testing.T) {
	t.Parallel()

	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")

	now := time.Now().Format("15:04")
	schedule := fmt.Sprintf(ScheduleTwoWorkflowOneStep, now, testprog, now, testprog)

	dir := t.TempDir()
	scheduleFile := filepath.Join(dir, "schedule.json")
	if err := os.WriteFile(scheduleFile, []byte(schedule), 0o600); err != nil {
		t.Fatal(err)
	}

	manifestFile := filepath.Join(dir, "manifest.txt")
	err := os.WriteFile(manifestFile, []byte(scheduleFile), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Workflow 2")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scheduler failed: %v\n%s", err, out)
	}

	expected := [][]byte{
		[]byte(`reason="workflows loaded" count=2`),
		[]byte(`reason="run once started"`),
		[]byte(`workFlow="Workflow 2"`),
		[]byte(`msg=workflow`),
		[]byte(`name="Workflow 2"`),
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

func TestPrintScheduleConfig(t *testing.T) {
	m1145, err := types.ParseMinuteOfDay("11:45")
	if err != nil {
		t.Fatalf("parse 11:45: %v", err)
	}

	m1146, err := types.ParseMinuteOfDay("11:46")
	if err != nil {
		t.Fatalf("parse 11:46: %v", err)
	}

	m1247, err := types.ParseMinuteOfDay("12:47")
	if err != nil {
		t.Fatalf("parse 12:47: %v", err)
	}

	s := Schedule{
		workflows: []Workflow{
			{
				Name:      "Workflow 1",
				Time:      m1146,
				OnFailure: &policy.Abort,
				Steps: []Step{
					{Name: "step 1", Timeout: 30},
				},
			},
			{
				Name:      "Workflow 1",
				Time:      m1145,
				OnFailure: &policy.Abort,
				Steps: []Step{
					{Name: "step 1", Timeout: 30, Pause: 5},
					{Name: "step 2", Timeout: 30},
				},
			},
			{
				Name:      "Workflow 2",
				Time:      m1145,
				OnFailure: &policy.Continue,
				Steps: []Step{
					{Name: "step 1", Timeout: 30},
				},
			},
			{
				Name:      "Workflow 1",
				Time:      m1247,
				OnFailure: &policy.Retry,
				Steps: []Step{
					{Name: "step 1", Timeout: 1800},
				},
			},
		},
	}

	var buf bytes.Buffer
	s.PrintScheduleConfig(&buf)

	got := buf.String()
	want := `1: 11:46  Workflow 1 (abort)
		1: step 1 (timeout 30s)
2: 11:45  Workflow 1 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)
3: 11:45  Workflow 2 (continue)
		1: step 1 (timeout 30s)
4: 12:47  Workflow 1 (retry)
		1: step 1 (timeout 30m0s)
`

	if got != want {
		t.Fatalf("unexpected output\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestPrintScheduleOperational(t *testing.T) {
	m1145, err := types.ParseMinuteOfDay("11:45")
	if err != nil {
		t.Fatalf("parse 11:45: %v", err)
	}

	m1146, err := types.ParseMinuteOfDay("11:46")
	if err != nil {
		t.Fatalf("parse 11:46: %v", err)
	}

	m1247, err := types.ParseMinuteOfDay("12:47")
	if err != nil {
		t.Fatalf("parse 12:47: %v", err)
	}

	s := Schedule{
		byMinute: map[types.MinuteOfDay][]Workflow{
			m1146: {
				{
					Name:      "Workflow 3",
					Time:      m1146,
					OnFailure: &policy.Abort,
					Steps: []Step{
						{Name: "step 1", Timeout: 30},
					},
				},
			},
			m1145: {
				{
					Name:      "Workflow 1",
					Time:      m1145,
					OnFailure: &policy.Abort,
					Steps: []Step{
						{Name: "step 1", Timeout: 30, Pause: 5},
						{Name: "step 2", Timeout: 30},
					},
				},
				{
					Name:      "Workflow 2",
					Time:      m1145,
					OnFailure: &policy.Continue,
					Steps: []Step{
						{Name: "step 1", Timeout: 30},
					},
				},
			},
			m1247: {
				{
					Name:      "Workflow 1",
					Time:      m1247,
					OnFailure: &policy.Retry,
					Steps: []Step{
						{Name: "step 1", Timeout: 1800},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	s.PrintScheduleOperational(&buf)

	got := buf.String()
	want := `1: 11:45  Workflow 1 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)
2: 11:45  Workflow 2 (continue)
		1: step 1 (timeout 30s)

1: 11:46  Workflow 3 (abort)
		1: step 1 (timeout 30s)

1: 12:47  Workflow 1 (retry)
		1: step 1 (timeout 30m0s)
`

	if got != want {
		t.Fatalf("unexpected output\ngot:\n%s\nwant:\n%s", got, want)
	}
}
