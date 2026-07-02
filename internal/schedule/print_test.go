package schedule

import (
	"bytes"
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

	s := Schedule{
		workflows: []workflow.Workflow{
			{
				Name: "Workflow 1",
				Trigger: workflow.Trigger{
					Every:   &c15m,
					BeginAt: &m1146,
				},
				OnFailure: workflow.Abort,
				Steps: []workflow.Step{
					{Name: "step 1", Timeout: timeout30s},
				},
			},
			{
				Name: "Workflow 1",
				Trigger: workflow.Trigger{
					Every:   &c15m,
					BeginAt: &m1145,
				},
				OnFailure: workflow.Abort,
				Steps: []workflow.Step{
					{Name: "step 1", Timeout: timeout30s, Pause: pause5s},
					{Name: "step 2", Timeout: timeout30s},
				},
			},
			{
				Name: "Workflow 2",
				Trigger: workflow.Trigger{
					Every:   &c15m,
					BeginAt: &m1145,
				},
				OnFailure: workflow.Continue,
				Steps: []workflow.Step{
					{Name: "step 1", Timeout: timeout30s},
				},
			},
			{
				Name: "Workflow 1",
				Trigger: workflow.Trigger{
					Every:   &c15m,
					BeginAt: &m1247,
				},
				OnFailure: workflow.Retry,
				Steps: []workflow.Step{
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
	c15m, err := workflow.ParseCadence("1h")
	if err != nil {
		t.Fatalf("parse cadence 1h: %v", err)
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

	s := Schedule{
		byMinute: map[workflow.MinuteOfDay][]workflow.Workflow{
			m1146: {
				{
					Name: "Workflow 3",
					Trigger: workflow.Trigger{
						Every:   &c15m,
						BeginAt: &m1146,
					},
					OnFailure: workflow.Abort,
					Steps: []workflow.Step{
						{Name: "step 1", Timeout: timeOut30s},
					},
				},
			},
			m1145: {
				{
					Name: "Workflow 1",
					Trigger: workflow.Trigger{
						Every:   &c15m,
						BeginAt: &m1145,
					},
					OnFailure: workflow.Abort,
					Steps: []workflow.Step{
						{Name: "step 1", Timeout: timeOut30s, Pause: pause5s},
						{Name: "step 2", Timeout: timeOut30s},
					},
				},
				{
					Name: "Workflow 2",
					Trigger: workflow.Trigger{
						Every:   &c15m,
						BeginAt: &m1145,
					},
					OnFailure: workflow.Continue,
					Steps: []workflow.Step{
						{Name: "step 1", Timeout: timeOut30s},
					},
				},
			},
			m1247: {
				{
					Name: "Workflow 1",
					Trigger: workflow.Trigger{
						Every:   &c15m,
						BeginAt: &m1247,
					},
					OnFailure: workflow.Retry,
					Steps: []workflow.Step{
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
