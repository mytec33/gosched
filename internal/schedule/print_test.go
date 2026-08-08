package schedule

import (
	"bytes"
	"strings"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/workflow"
)

func TestPrintScheduleConfig(t *testing.T) {
	m1145, err := workflow.ParseMinuteOfDay("11:45")
	if err != nil {
		t.Fatalf("parse 11:45: %v", err)
	}

	m1146, err := workflow.ParseMinuteOfDay("11:46")
	if err != nil {
		t.Fatalf("parse 11:46: %v", err)
	}

	m1247, err := workflow.ParseMinuteOfDay("12:47")
	if err != nil {
		t.Fatalf("parse 12:47: %v", err)
	}

	c15m, err := workflow.ParseCadence("1h")
	if err != nil {
		t.Fatalf("parse cadence 1h: %v", err)
	}

	pause5s, err := workflow.ParseConfigDuration("5s")
	if err != nil {
		t.Fatalf("parse config duration 5s: %v", err)
	}

	timeout30s, err := workflow.ParseConfigDuration("30s")
	if err != nil {
		t.Fatalf("parse config duration 30s: %v", err)
	}

	timeout1800s, err := workflow.ParseConfigDuration("1800s")
	if err != nil {
		t.Fatalf("parse config duration 1800s: %v", err)
	}

	workflows := []workflow.Workflow{
		{
			SourceFile: "/home/foo/etl.json",
			Name:       "Workflow 1",
			Enabled:    true,
			Trigger: workflow.Trigger{
				Every:   c15m,
				BeginAt: m1146,
			},
			OnFailure: workflow.Abort,
			Steps: []workflow.Step{
				{Name: "step 1", Timeout: timeout30s},
			},
		},
		{
			SourceFile: "/home/foo/etl.json",
			Name:       "Workflow 2",
			Enabled:    true,
			Trigger: workflow.Trigger{
				Every:   c15m,
				BeginAt: m1145,
			},
			OnFailure: workflow.Abort,
			Steps: []workflow.Step{
				{Name: "step 1", Timeout: timeout30s, Pause: pause5s},
				{Name: "step 2", Timeout: timeout30s},
			},
		},
		{
			SourceFile: "/home/foo/csv_exports.json",
			Name:       "Workflow 3",
			Enabled:    true,
			Trigger: workflow.Trigger{
				Every:   c15m,
				BeginAt: m1145,
			},
			OnFailure: workflow.Continue,
			Steps: []workflow.Step{
				{Name: "step 1", Timeout: timeout30s},
			},
		},
		{
			SourceFile: "/home/foo/csv_exports.json",
			Name:       "Workflow 4",
			Enabled:    true,
			Trigger: workflow.Trigger{
				Every:   c15m,
				BeginAt: m1247,
			},
			OnFailure: workflow.Retry,
			Steps: []workflow.Step{
				{Name: "step 1", Timeout: timeout1800s},
			},
		},
		{
			SourceFile:     "/home/foo/csv_exports.json",
			Name:           "Workflow 5",
			Enabled:        false,
			DisabledReason: "demonstrating disabled display",
			Trigger: workflow.Trigger{
				Every:   c15m,
				BeginAt: m1247,
			},
			OnFailure: workflow.Retry,
			Steps: []workflow.Step{
				{Name: "step 1", Timeout: timeout1800s},
			},
		},
	}

	s, errorList := New(workflows)
	if len(errorList) > 0 {
		t.Fatalf("expected no errors, got %v\n", errorList)
	}

	var buf bytes.Buffer
	s.printScheduleConfig(&buf)

	got := buf.String()
	want := `
/home/foo/etl.json:
1: 1h 11:46  Workflow 1 (abort)
		1: step 1 (timeout 30s)
2: 1h 11:45  Workflow 2 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)

/home/foo/csv_exports.json:
3: 1h 11:45  Workflow 3 (continue)
		1: step 1 (timeout 30s)
4: 1h 12:47  Workflow 4 (retry)
		1: step 1 (timeout 30m0s)
5: 1h 12:47  Workflow 5 (retry) <--- disabled: demonstrating disabled display
		1: step 1 (timeout 30m0s)

`

	if got != want {
		t.Fatalf("unexpected output\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestDisplayReason(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "short reason unchanged", input: "maintenance", want: "maintenance"},
		{name: "exactly max unchanged", input: strings.Repeat("a", 40), want: strings.Repeat("a", 40)},
		{name: "over max truncated with ellipsis", input: strings.Repeat("a", 41), want: strings.Repeat("a", 40) + "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := displayReason(tt.input)
			if got != tt.want {
				t.Fatalf("displayReason(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPrintStep(t *testing.T) {
	timeout30s, err := workflow.ParseConfigDuration("30s")
	if err != nil {
		t.Fatalf("parse config duration 30s: %v", err)
	}

	pause5s, err := workflow.ParseConfigDuration("5s")
	if err != nil {
		t.Fatalf("parse config duration 5s: %v", err)
	}

	tests := []struct {
		name string
		step workflow.Step
		want string
	}{
		{
			name: "timeout and pause",
			step: workflow.Step{Name: "step 1", Timeout: timeout30s, Pause: pause5s},
			want: "\t\t1: step 1 (timeout 30s, pause 5s)\n",
		},
		{
			name: "timeout only",
			step: workflow.Step{Name: "step 1", Timeout: timeout30s},
			want: "\t\t1: step 1 (timeout 30s)\n",
		},
		{
			name: "no timeout",
			step: workflow.Step{Name: "step 1"},
			want: "\t\t1: step 1 (no timeout)\n",
		},
		{
			name: "no timeout with pause",
			step: workflow.Step{Name: "step 1", Pause: pause5s},
			want: "\t\t1: step 1 (no timeout, pause 5s)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printStep(1, 0, tt.step, &buf)

			if got := buf.String(); got != tt.want {
				t.Errorf("unexpected output\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestPrintScheduleOperational(t *testing.T) {
	c1d, err := workflow.ParseCadence("1d")
	if err != nil {
		t.Fatalf("parse cadence 1d: %v", err)
	}

	m1145, err := workflow.ParseMinuteOfDay("11:45")
	if err != nil {
		t.Fatalf("parse 11:45: %v", err)
	}

	m1146, err := workflow.ParseMinuteOfDay("11:46")
	if err != nil {
		t.Fatalf("parse 11:46: %v", err)
	}

	m1247, err := workflow.ParseMinuteOfDay("12:47")
	if err != nil {
		t.Fatalf("parse 12:47: %v", err)
	}

	pause5s, err := workflow.ParseConfigDuration("5s")
	if err != nil {
		t.Fatalf("parse config duration 5s: %v", err)
	}

	timeOut30s, err := workflow.ParseConfigDuration("30s")
	if err != nil {
		t.Fatalf("parse config duration 30s: %v", err)
	}

	timeOut1800s, err := workflow.ParseConfigDuration("1800s")
	if err != nil {
		t.Fatalf("parse config duration 1800s: %v", err)
	}

	workflows := []workflow.Workflow{
		{
			Name:    "Workflow 1",
			Enabled: true,
			Trigger: workflow.Trigger{
				Every:   c1d,
				BeginAt: m1145,
			},
			OnFailure: workflow.Abort,
			Steps: []workflow.Step{
				{Name: "step 1", Timeout: timeOut30s, Pause: pause5s},
				{Name: "step 2", Timeout: timeOut30s},
			},
		},
		{
			Name:    "Workflow 2",
			Enabled: true,
			Trigger: workflow.Trigger{
				Every:   c1d,
				BeginAt: m1145,
			},
			OnFailure: workflow.Continue,
			Steps: []workflow.Step{
				{Name: "step 1", Timeout: timeOut30s},
			},
		},
		{
			Name:    "Workflow 3",
			Enabled: true,
			Trigger: workflow.Trigger{
				Every:   c1d,
				BeginAt: m1146,
			},
			OnFailure: workflow.Abort,
			Steps: []workflow.Step{
				{Name: "step 1", Timeout: timeOut30s},
			},
		},
		{
			Name:    "Workflow 1 at 12:47",
			Enabled: true,
			Trigger: workflow.Trigger{
				Every:   c1d,
				BeginAt: m1247,
			},
			OnFailure: workflow.Retry,
			Steps: []workflow.Step{
				{Name: "step 1", Timeout: timeOut1800s},
			},
		},
		{
			Name:           "Workflow 4",
			Enabled:        false,
			DisabledReason: "disabled to test enable functionality",
			Trigger: workflow.Trigger{
				Every:   c1d,
				BeginAt: m1247,
			},
			OnFailure: workflow.Retry,
			Steps: []workflow.Step{
				{Name: "step 1", Timeout: timeOut1800s},
			},
		},
	}

	s, errorList := New(workflows)
	if len(errorList) > 0 {
		t.Fatalf("expected no errors, got %v\n", errorList)
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

1: 12:47  Workflow 1 at 12:47 (retry)
		1: step 1 (timeout 30m0s)
2: 12:47  Workflow 4 (retry) <--- disabled: disabled to test enable functionality
		1: step 1 (timeout 30m0s)
`

	if got != want {
		t.Fatalf("unexpected output\ngot:\n%s\nwant:\n%s", got, want)
	}
}
