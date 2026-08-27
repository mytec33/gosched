package schedule

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/workflow"
)

func TestNewExpandsHourlyWorkflowFromMidnight(t *testing.T) {
	manifestFile := filepath.Join("testdata", "hourly_manifest.txt")
	sched, errorList := New(manifestFile)
	if len(errorList) > 0 {
		t.Fatalf("expected no errors, got %v", errorList)
	}

	if _, err := sched.GetWorkflowByName("test-hourly"); err != nil {
		t.Fatalf("expected to find workflow by name: %v", err)
	}

	populated := 0
	for m := workflow.MinuteOfDay(0); m < workflow.MinuteOfDay(MinutesPerDay); m++ {
		if len(sched.WorkflowsAtMinute(m)) > 0 {
			populated++
		}
	}
	if populated != 24 {
		t.Fatalf("expected 24 populated minute slots, got %d", populated)
	}

	if got := len(sched.WorkflowsAtMinute(workflow.MinuteOfDay(0))); got != 1 {
		t.Fatalf("expected 1 workflow at minute 0, got %d", got)
	}
	if got := len(sched.WorkflowsAtMinute(workflow.MinuteOfDay(60))); got != 1 {
		t.Fatalf("expected 1 workflow at minute 60, got %d", got)
	}
	if got := len(sched.WorkflowsAtMinute(workflow.MinuteOfDay(30))); got != 0 {
		t.Fatalf("expected 0 workflows at minute 30, got %d", got)
	}
}

func TestScheduleValidate_DuplicateWorkflowNames_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
	}{
		{
			name: "Identical",
			schedule: Schedule{
				workflows: []workflow.Workflow{
					{Name: "Workflow 1"},
					{Name: "Workflow 2"},
					{Name: "Workflow 1"},
				},
			},
		},
		{
			name: "Leading/trailing whitespace",
			schedule: Schedule{
				workflows: []workflow.Workflow{
					{Name: "Workflow 1\t"},
					{Name: " Workflow 1"},
				},
			},
		},
		{
			name: "Mixed case",
			schedule: Schedule{
				workflows: []workflow.Workflow{
					{Name: "Workflow 1"},
					{Name: "worKfLow 1"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			errorList := tt.schedule.validate()

			found := false
			for _, e := range errorList {
				if errors.Is(e, ErrDuplicateWorkflowName) {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("expected %v, got %v", ErrDuplicateWorkflowName, errorList)
			}
		})
	}
}

func TestScheduleValidate_DuplicateWorkflowNames_Valid(t *testing.T) {
	s := Schedule{
		workflows: []workflow.Workflow{
			{Name: "Workflow 1"},
			{Name: "Workflow 2"},
			{Name: "workflow 3"},
		},
	}

	errorList := s.validate()

	found := false
	for _, e := range errorList {
		if errors.Is(e, ErrDuplicateWorkflowName) {
			found = true
			break
		}
	}

	if found {
		t.Fatalf("expected %v", ErrDuplicateWorkflowName)
	}
}

func TestScheduleValidate_WorkflowCount_Valid(t *testing.T) {
	// Valid range is between 1 and max workflow count
	s := Schedule{workflows: makeWorkflows(MinWorkflowCount)}
	err := s.validateWorkflowCount()
	if len(err) > 0 {
		t.Fatalf("expected no error, got %v", err)
	}

	s = Schedule{workflows: makeWorkflows(MaxWorkflowCount)}
	err = s.validateWorkflowCount()
	if len(err) > 0 {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestScheduleValidate_WorkflowCount_Invalid(t *testing.T) {
	expectedError := ErrWorkflowsEmpty
	s := Schedule{workflows: makeWorkflows(0)}
	err := s.validateWorkflowCount()
	if !hasError(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	expectedError = ErrWorkflowCountExceeded
	s = Schedule{workflows: makeWorkflows(MaxWorkflowCount + 1)}
	err = s.validateWorkflowCount()
	if !hasError(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}
}

func hasError(errorList []error, target error) bool {
	for _, e := range errorList {
		if errors.Is(e, target) {
			return true
		}
	}

	return false
}

func makeWorkflows(n int) []workflow.Workflow {
	workflows := make([]workflow.Workflow, 0, n)
	for i := range n {
		workflows = append(workflows, workflow.Workflow{
			Name: fmt.Sprintf("Workflow %d", i),
		})
	}
	return workflows
}

func TestScheduleStepCount(t *testing.T) {
	s := Schedule{
		workflows: []workflow.Workflow{
			{Steps: []workflow.Step{{}}},
			{Steps: []workflow.Step{{}, {}}},
		},
	}

	if got := s.StepCount(); got != 3 {
		t.Fatalf("StepCount() = %d, want 3", got)
	}
}

func TestScheduleValid(t *testing.T) {
	if (Schedule{}).Valid() {
		t.Fatal("expected zero value Schedule to be invalid")
	}

	missingName := Schedule{
		workflows: []workflow.Workflow{
			{Steps: []workflow.Step{{}}},
		},
	}
	if missingName.Valid() {
		t.Fatal("expected Schedule with unnamed Workflow to be invalid")
	}

	missingSteps := Schedule{
		workflows: []workflow.Workflow{
			{Name: "Workflow 1"},
		},
	}
	if missingSteps.Valid() {
		t.Fatal("expected Schedule with no Workflow steps to be invalid")
	}

	valid := Schedule{
		workflows: []workflow.Workflow{
			{Name: "Workflow 1", Steps: []workflow.Step{{}}},
		},
	}
	if !valid.Valid() {
		t.Fatal("expected named Workflow with a step to produce a valid Schedule")
	}
}

func TestExpandCadence(t *testing.T) {
	tests := []struct {
		name      string
		every     string
		beginAt   string
		wantCount int
		wantFirst string
		wantLast  string
	}{
		{"daily pinned", "1d", "03:15", 1, "03:15", "03:15"},
		{"15 minutes all day", "15m", "00:00", 96, "00:00", "23:45"},
		{"15 minutes half day", "15m", "12:00", 48, "12:00", "23:45"},
		{"6 hours all day", "6h", "00:00", 4, "00:00", "18:00"},
		{"6 hours late start", "6h", "21:00", 1, "21:00", "21:00"},
		{"7 hours uneven day", "7h", "00:00", 4, "00:00", "21:00"},
		{"1 minute all day", "1m", "00:00", 1440, "00:00", "23:59"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cadence, err := workflow.ParseCadence(tt.every)
			if err != nil {
				t.Fatalf("parse cadence %q: %v", tt.every, err)
			}

			beginAt, err := workflow.ParseMinuteOfDay(tt.beginAt)
			if err != nil {
				t.Fatalf("parse beginAt %q: %v", tt.beginAt, err)
			}

			trigger := workflow.Trigger{
				Every:   cadence,
				BeginAt: beginAt,
			}

			got, err := expandCadence(trigger)
			if err != nil {
				t.Fatalf("expand cadence: %v", err)
			}

			if len(got) != tt.wantCount {
				t.Fatalf("got %d minutes, want %d: %v", len(got), tt.wantCount, got)
			}

			if got[0].String() != tt.wantFirst {
				t.Fatalf("first minute = %s, want %s", got[0].String(), tt.wantFirst)
			}

			last := got[len(got)-1]
			if last.String() != tt.wantLast {
				t.Fatalf("last minute = %s, want %s", last.String(), tt.wantLast)
			}
		})
	}
}

func TestExpandSchedule_Idempotent(t *testing.T) {
	cadence, err := workflow.ParseCadence("15m")
	if err != nil {
		t.Fatalf("parse cadence: %v", err)
	}

	beginAt, err := workflow.ParseMinuteOfDay("00:00")
	if err != nil {
		t.Fatalf("parse beginAt: %v", err)
	}

	s := Schedule{
		workflows: []workflow.Workflow{
			{
				Name:    "Jackpots",
				Enabled: true,
				Trigger: workflow.Trigger{
					Every:   cadence,
					BeginAt: beginAt,
				},
				OnFailure: workflow.Retry,
				Steps: []workflow.Step{
					{Name: "step 1", Program: "program"},
				},
			},
		},
	}

	if err := s.expandSchedule(); err != nil {
		t.Fatalf("first expansion: %v", err)
	}

	firstCount := 0
	for _, workflows := range s.byMinute {
		firstCount += len(workflows)
	}

	if err := s.expandSchedule(); err != nil {
		t.Fatalf("second expansion: %v", err)
	}

	secondCount := 0
	for _, workflows := range s.byMinute {
		secondCount += len(workflows)
	}

	if firstCount != 96 {
		t.Fatalf("first expansion count = %d, want 96", firstCount)
	}

	if secondCount != firstCount {
		t.Fatalf("second expansion count = %d, want %d", secondCount, firstCount)
	}
}
