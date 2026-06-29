package schedule

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/types"
)

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

	c15m, err := types.ParseCadence("1h")
	if err != nil {
		t.Fatalf("parse cadence 1h: %v", err)
	}

	pause5s, err := types.ParseConfigDuration("5s")
	if err != nil {
		t.Fatalf("parse config duration 5s: %v", err)
	}

	timeout30s, err := types.ParseConfigDuration("30s")
	if err != nil {
		t.Fatalf("parse config duration 30s: %v", err)
	}

	timeout1800s, err := types.ParseConfigDuration("1800s")
	if err != nil {
		t.Fatalf("parse config duration 1800s: %v", err)
	}

	s := Schedule{
		workflows: []Workflow{
			{
				Name: "Workflow 1",
				Trigger: types.Trigger{
					Every:   &c15m,
					BeginAt: &m1146,
				},
				OnFailure: types.Abort,
				Steps: []Step{
					{Name: "step 1", Timeout: timeout30s},
				},
			},
			{
				Name: "Workflow 1",
				Trigger: types.Trigger{
					Every:   &c15m,
					BeginAt: &m1145,
				},
				OnFailure: types.Abort,
				Steps: []Step{
					{Name: "step 1", Timeout: timeout30s, Pause: pause5s},
					{Name: "step 2", Timeout: timeout30s},
				},
			},
			{
				Name: "Workflow 2",
				Trigger: types.Trigger{
					Every:   &c15m,
					BeginAt: &m1145,
				},
				OnFailure: types.Continue,
				Steps: []Step{
					{Name: "step 1", Timeout: timeout30s},
				},
			},
			{
				Name: "Workflow 1",
				Trigger: types.Trigger{
					Every:   &c15m,
					BeginAt: &m1247,
				},
				OnFailure: types.Retry,
				Steps: []Step{
					{Name: "step 1", Timeout: timeout1800s},
				},
			},
		},
	}

	var buf bytes.Buffer
	s.printScheduleConfig(&buf)

	got := buf.String()
	want := `1: 1h 11:46  Workflow 1 (abort)
		1: step 1 (timeout 30s)
2: 1h 11:45  Workflow 1 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)
3: 1h 11:45  Workflow 2 (continue)
		1: step 1 (timeout 30s)
4: 1h 12:47  Workflow 1 (retry)
		1: step 1 (timeout 30m0s)
`

	if got != want {
		t.Fatalf("unexpected output\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestPrintScheduleOperational(t *testing.T) {
	c15m, err := types.ParseCadence("1h")
	if err != nil {
		t.Fatalf("parse cadence 1h: %v", err)
	}

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

	pause5s, err := types.ParseConfigDuration("5s")
	if err != nil {
		t.Fatalf("parse config duration 5s: %v", err)
	}

	timeOut30s, err := types.ParseConfigDuration("30s")
	if err != nil {
		t.Fatalf("parse config duration 30s: %v", err)
	}

	timeOut1800s, err := types.ParseConfigDuration("1800s")
	if err != nil {
		t.Fatalf("parse config duration 1800s: %v", err)
	}

	s := Schedule{
		byMinute: map[types.MinuteOfDay][]Workflow{
			m1146: {
				{
					Name: "Workflow 3",
					Trigger: types.Trigger{
						Every:   &c15m,
						BeginAt: &m1146,
					},
					OnFailure: types.Abort,
					Steps: []Step{
						{Name: "step 1", Timeout: timeOut30s},
					},
				},
			},
			m1145: {
				{
					Name: "Workflow 1",
					Trigger: types.Trigger{
						Every:   &c15m,
						BeginAt: &m1145,
					},
					OnFailure: types.Abort,
					Steps: []Step{
						{Name: "step 1", Timeout: timeOut30s, Pause: pause5s},
						{Name: "step 2", Timeout: timeOut30s},
					},
				},
				{
					Name: "Workflow 2",
					Trigger: types.Trigger{
						Every:   &c15m,
						BeginAt: &m1145,
					},
					OnFailure: types.Continue,
					Steps: []Step{
						{Name: "step 1", Timeout: timeOut30s},
					},
				},
			},
			m1247: {
				{
					Name: "Workflow 1",
					Trigger: types.Trigger{
						Every:   &c15m,
						BeginAt: &m1247,
					},
					OnFailure: types.Retry,
					Steps: []Step{
						{Name: "step 1", Timeout: timeOut1800s},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	s.printScheduleOperational(&buf)

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

func TestScheduleValidate_DuplicateWorkflowNames_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
	}{
		{
			name: "Identical",
			schedule: Schedule{
				workflows: []Workflow{
					{Name: "Workflow 1"},
					{Name: "Workflow 2"},
					{Name: "Workflow 1"},
				},
			},
		},
		{
			name: "Leading/trailing whitespace",
			schedule: Schedule{
				workflows: []Workflow{
					{Name: "Workflow 1\t"},
					{Name: " Workflow 1"},
				},
			},
		},
		{
			name: "Mixed case",
			schedule: Schedule{
				workflows: []Workflow{
					{Name: "Workflow 1"},
					{Name: "worKfLow 1"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			errorList := tt.schedule.Validate()

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
		workflows: []Workflow{
			{Name: "Workflow 1"},
			{Name: "Workflow 2"},
			{Name: "workflow 3"},
		},
	}

	errorList := s.Validate()

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

func TestScheduleValidate_WorkflowCount_ValidAtLimit(t *testing.T) {
	s := Schedule{workflows: makeWorkflows(MaxWorkflowCount)}

	errorList := s.Validate()

	if hasError(errorList, ErrWorkflowCountExceeded) {
		t.Fatalf("unexpected %v in %v", ErrWorkflowCountExceeded, errorList)
	}
}

func TestScheduleValidate_WorkflowCount_InvalidOverLimit(t *testing.T) {
	s := Schedule{workflows: makeWorkflows(MaxWorkflowCount + 1)}

	errorList := s.Validate()

	if !hasError(errorList, ErrWorkflowCountExceeded) {
		t.Fatalf("expected %v, got %v", ErrWorkflowCountExceeded, errorList)
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

func makeWorkflows(n int) []Workflow {
	workflows := make([]Workflow, 0, n)
	for i := range n {
		workflows = append(workflows, Workflow{
			Name: fmt.Sprintf("Workflow %d", i),
		})
	}
	return workflows
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
			cadence, err := types.ParseCadence(tt.every)
			if err != nil {
				t.Fatalf("parse cadence %q: %v", tt.every, err)
			}

			beginAt, err := types.ParseMinuteOfDay(tt.beginAt)
			if err != nil {
				t.Fatalf("parse beginAt %q: %v", tt.beginAt, err)
			}

			trigger := types.Trigger{
				Every:   &cadence,
				BeginAt: &beginAt,
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
	cadence, err := types.ParseCadence("15m")
	if err != nil {
		t.Fatalf("parse cadence: %v", err)
	}

	beginAt, err := types.ParseMinuteOfDay("00:00")
	if err != nil {
		t.Fatalf("parse beginAt: %v", err)
	}

	s := Schedule{
		workflows: []Workflow{
			{
				Name: "Jackpots",
				Trigger: types.Trigger{
					Every:   &cadence,
					BeginAt: &beginAt,
				},
				OnFailure: types.Retry,
				Steps: []Step{
					{Name: "step 1", Program: "program"},
				},
			},
		},
	}

	if err := s.ExpandSchedule(); err != nil {
		t.Fatalf("first expansion: %v", err)
	}

	firstCount := 0
	for _, workflows := range s.byMinute {
		firstCount += len(workflows)
	}

	if err := s.ExpandSchedule(); err != nil {
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
