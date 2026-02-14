package schedule

import (
	"errors"
	"strings"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/platform"
)

type ErrorCode string

const WorkflowNameEmpty = `
[
  {
    "name": "",
    "time": "10:35",
    "steps": [{"name": "daily", "program": "program", "args": "args"}]
  }
]
`

const WorkflowNameWhitespace = `
[
  {
    "name": "        ",
    "time": "10:35",
    "steps": [{"name": "daily", "program": "program", "args": "args"}]
  }
]
`

const WorkflowNameWhitespaceLeading = `
[
  {
    "name": " name",
    "time": "10:35",
    "steps": [{"name": "daily", "program": "program", "args": "args"}]
  }
]
`

const WorkflowNameWhitespaceTrailing = `
[
  {
    "name": "name ",
    "time": "10:35",
    "steps": [{"name": "daily", "program": "program", "args": "args"}]
  }
]
`

const WorkflowNameTooLong = `
[
  {
    "name": "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis,..",
    "time": "10:35",
    "steps": [{"name": "daily", "program": "program", "args": "args"}]
  }
]
`

const WorkflowTimeEmpty = `
[
  {
    "name": "foo",
    "time": "",
    "steps": [{"name": "daily", "program": "program", "args": "args"}]
  }
]
`

const WorkflowTimeBadHour = `
[
  {
    "name": "foo",
    "time": "99:35",
    "steps": [{"name": "daily", "program": "program", "args": "args"}]
  }
]
`

const WorkflowTimeWhitespace = `
[
  {
    "name": "     ",
    "time": "99:35",
    "steps": [{"name": "daily", "program": "program", "args": "args"}]
  }
]
`

func TestWorkflow_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantCode error
	}{
		{name: "name empty", json: WorkflowNameEmpty, wantCode: ErrEmpty},
		{name: "name whitespace", json: WorkflowNameWhitespace, wantCode: ErrWhitespaceAll},
		{name: "name whitespace leading", json: WorkflowNameWhitespaceLeading, wantCode: ErrWhitespaceLeadingOrTrailing},
		{name: "name whitespace trailing", json: WorkflowNameWhitespaceTrailing, wantCode: ErrWhitespaceLeadingOrTrailing},
		{name: "name too long", json: WorkflowNameTooLong, wantCode: ErrTooLong},
		{name: "time empty", json: WorkflowTimeEmpty, wantCode: ErrEmpty},
		{name: "time bad hour", json: WorkflowTimeBadHour, wantCode: ErrInvalidTime},
		{name: "time bad whitespace", json: WorkflowTimeWhitespace, wantCode: ErrInvalidTime},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, validationErrors, err := DecodeSchedule(r)

			if err != nil {
				t.Fatalf("%s: unexpected decode/system error: %v", tt.name, err)
			}

			if len(validationErrors) == 0 {
				t.Fatalf("%v: expected validation error(s), got none", tt.name)
			}

			found := false
			for _, ve := range validationErrors {
				if errors.Is(ve, tt.wantCode) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("%s: got %v, want an error matching %v", tt.name, validationErrors, tt.wantCode)
			}
		})
	}
}

const WorkflowStepNameEmpty = `
[
  {
    "name": "name",
    "time": "10:35",
    "steps": [{"name": "", "program": "program", "args": "args"}]
  }
]
`

const WorkflowStepNameWhitespace = `
[
  {
    "name": "name",
    "time": "10:35",
    "steps": [{"name": "   ", "program": "program", "args": "args"}]
  }
]
`

const WorkflowStepNameWhitespaceLeading = `
[
  {
    "name": "name",
    "time": "10:35",
    "steps": [{"name": " leading", "program": "program", "args": "args"}]
  }
]
`

const WorkflowStepNameWhitespaceTrailing = `
[
  {
    "name": "name",
    "time": "10:35",
    "steps": [{"name": "trailing ", "program": "program", "args": "args"}]
  }
]
`

const WorkflowStepNameTooLong = `
[
  {
    "name": "name",
    "time": "10:35",
    "steps": [{"name": "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis,..", "program": "program", "args": "args"}]
  }
]
`

const WorkflowStepProgramEmpty = `
[
  {
    "name": "name",
    "time": "10:35",
    "steps": [{"name": "step name", "program": "", "args": "args"}]
  }
]
`

const WorkflowStepProgramWhitespace = `
[
  {
    "name": "name",
    "time": "10:35",
    "steps": [{"name": "step name", "program": "     ", "args": "args"}]
  }
]
`

const WorkflowStepProgramWhitespaceLeading = `
[
  {
    "name": "name",
    "time": "10:35",
    "steps": [{"name": "step name", "program": " foo", "args": "args"}]
  }
]
`

const WorkflowStepProgramWhitespaceTrailing = `
[
  {
    "name": "name",
    "time": "10:35",
    "steps": [{"name": "step name", "program": "foo ", "args": "args"}]
  }
]
`

const WorkflowStepProgramTooLong = `
[
  {
    "name": "name",
    "time": "10:35",
    "steps": [{"name": "step name", "program": "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis,..", "args": "args"}]
  }
]
`

func TestWorkflowSteps_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantCode error
	}{
		{name: "name empty", json: WorkflowStepNameEmpty, wantCode: ErrEmpty},
		{name: "name whitespace", json: WorkflowStepNameWhitespace, wantCode: ErrWhitespaceAll},
		{name: "name whitespace leading", json: WorkflowStepNameWhitespaceLeading, wantCode: ErrWhitespaceLeadingOrTrailing},
		{name: "name whitespace trailing", json: WorkflowStepNameWhitespaceTrailing, wantCode: ErrWhitespaceLeadingOrTrailing},
		{name: "name too long", json: WorkflowStepNameTooLong, wantCode: ErrTooLong},
		{name: "program empty", json: WorkflowStepProgramEmpty, wantCode: ErrEmpty},
		{name: "program whitespace", json: WorkflowStepProgramWhitespace, wantCode: platform.ErrPlatformWhitespaceAll},
		{name: "program whitespace leading", json: WorkflowStepProgramWhitespaceLeading, wantCode: platform.ErrPlatformWhitespaceLeadingOrTrailing},
		{name: "program whitespace trailing", json: WorkflowStepProgramWhitespaceTrailing, wantCode: platform.ErrPlatformWhitespaceLeadingOrTrailing},
		{name: "program too long", json: WorkflowStepProgramTooLong, wantCode: platform.ErrPlatformTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, validationErrors, err := DecodeSchedule(r)

			if err != nil {
				t.Fatalf("%s: unexpected decode/system error: %v", tt.name, err)
			}

			if len(validationErrors) == 0 {
				t.Fatalf("%v: expected validation error(s), got none", tt.name)
			}

			found := false
			for _, ve := range validationErrors {
				if errors.Is(ve, tt.wantCode) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("%s: got %v, want an error matching %v", tt.name, validationErrors, tt.wantCode)
			}
		})
	}
}
