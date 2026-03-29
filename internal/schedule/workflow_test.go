package schedule

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/errs"
	"git.sr.ht/~mytec/gosched/internal/platform"
)

const WorkflowNameEmpty = `
[
  {
    "name": "",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowNameWhitespace = `
[
  {
    "name": "        ",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowNameWhitespaceLeading = `
[
  {
    "name": " name",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowNameWhitespaceTrailing = `
[
  {
    "name": "name ",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowNameTooLong = `
[
  {
    "name": "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis,..",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowNoSteps = `
[
  {
    "name": "name",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": []
  }
]
`

func TestWorkflow_Invalid(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError error
	}{
		{name: "name empty", json: WorkflowNameEmpty, wantError: errs.ErrEmpty},
		{name: "name whitespace", json: WorkflowNameWhitespace, wantError: errs.ErrWhitespaceAll},
		{name: "name whitespace leading", json: WorkflowNameWhitespaceLeading, wantError: errs.ErrWhitespaceLeadingOrTrailing},
		{name: "name whitespace trailing", json: WorkflowNameWhitespaceTrailing, wantError: errs.ErrWhitespaceLeadingOrTrailing},
		{name: "name too long", json: WorkflowNameTooLong, wantError: errs.ErrTooLong},
		{name: "missing steps", json: WorkflowNoSteps, wantError: errs.ErrEmpty},
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
				if errors.Is(ve, tt.wantError) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("%s: got %v, want an error matching %v", tt.name, validationErrors, tt.wantError)
			}
		})
	}
}

const WorkflowStepNameEmpty = `
[
  {
    "name": "name",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowStepNameWhitespace = `
[
  {
    "name": "name",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "   ", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowStepNameWhitespaceLeading = `
[
  {
    "name": "name",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": " leading", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowStepNameWhitespaceTrailing = `
[
  {
    "name": "name",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "trailing ", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowStepNameTooLong = `
[
  {
    "name": "name",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis,..", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowStepProgramEmpty = `
[
  {
    "name": "name",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "", "args": ["args"]}]
  }
]
`

const WorkflowStepProgramWhitespace = `
[
  {
    "name": "name",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "     ", "args": ["args"]}]
  }
]
`

const WorkflowStepProgramWhitespaceLeading = `
[
  {
    "name": "name",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": " foo", "args": ["args"]}]
  }
]
`

const WorkflowStepProgramWhitespaceTrailing = `
[
  {
    "name": "name",
    "time": "10:35",
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "foo ", "args": ["args"]}]
  }
]
`

func TestWorkflowSteps_Invalid(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError error
	}{
		{name: "name empty", json: WorkflowStepNameEmpty, wantError: errs.ErrEmpty},
		{name: "name whitespace", json: WorkflowStepNameWhitespace, wantError: errs.ErrWhitespaceAll},
		{name: "name whitespace leading", json: WorkflowStepNameWhitespaceLeading, wantError: errs.ErrWhitespaceLeadingOrTrailing},
		{name: "name whitespace trailing", json: WorkflowStepNameWhitespaceTrailing, wantError: errs.ErrWhitespaceLeadingOrTrailing},
		{name: "name too long", json: WorkflowStepNameTooLong, wantError: errs.ErrTooLong},
		{name: "program empty", json: WorkflowStepProgramEmpty, wantError: errs.ErrEmpty},
		{name: "program whitespace", json: WorkflowStepProgramWhitespace, wantError: errs.ErrWhitespaceAll},
		{name: "program whitespace leading", json: WorkflowStepProgramWhitespaceLeading, wantError: errs.ErrWhitespaceLeadingOrTrailing},
		{name: "program whitespace trailing", json: WorkflowStepProgramWhitespaceTrailing, wantError: errs.ErrWhitespaceLeadingOrTrailing},
		{name: "program too long", json: workflowStepProgramTooLongJSON(), wantError: errs.ErrTooLong},
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
				if errors.Is(ve, tt.wantError) {
					found = true
				} else {
					t.Fatalf("%s: unexpected validation error: %v (expected only %v). Full list: %v",
						tt.name, ve, tt.wantError, validationErrors)
				}
			}
			if !found {
				t.Fatalf("%s: missing expected error %v. Full list: %v", tt.name, tt.wantError, validationErrors)
			}
		})
	}
}

func workflowStepProgramTooLongJSON() string {
	maxLength := platform.MaxPathLength()

	tooLong := strings.Repeat("a", maxLength+1)

	return fmt.Sprintf(`
      [
        {
          "name": "name",
          "time": "10:35",
          "onFailure": "continue",          
          "steps": [{"name": "step name", "program": "%s", "args": ["args"]}]
        }
      ]`, tooLong)
}

const WorkflowNamesNotUnique = `
[
  {
    "name": "workflow 1",
    "time": "10:35",
    "onFailure": "continue",
    "steps": [{"name": "step name", "program": "program", "args": ["args"], "timeout": 43200, "pause": 3600}]
  },
  {
    "name": "workflow 1",
    "onFailure": "continue",    
    "time": "11:35",
    "steps": [{"name": "step name 2", "program": "program 1", "args": ["args 1"], "timeout": 0, "pause": 0}]
  }
]
`

func TestWorkflowNamesNotUnique(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError error
	}{
		{name: "duplicate workflow names", json: WorkflowNamesNotUnique, wantError: errs.ErrDuplicateWorkflowName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, validationErrors, _ := DecodeSchedule(r)

			found := false
			for _, ve := range validationErrors {
				if errors.Is(ve, tt.wantError) {
					found = true
				} else {
					t.Fatalf("%s: unexpected validation error: %v (expected only %v). Full list: %v",
						tt.name, ve, tt.wantError, validationErrors)
				}
			}
			if !found {
				t.Fatalf("%s: missing expected error %v. Full list: %v", tt.name, tt.wantError, validationErrors)
			}
		})
	}
}

const StepNamesNotUnique = `
[
  {
    "name": "workflow 1",
    "time": "10:35",
    "onFailure": "continue",
    "steps": [
      {"name": "step name", "program": "program", "args": ["args"], "timeout": 43200, "pause": 3600},
      {"name": "step name", "program": "program", "args": ["args"], "timeout": 43200, "pause": 3600}      
    ]
  },
  {
    "name": "workflow 2",
    "onFailure": "continue",    
    "time": "11:35",
    "steps": [{"name": "step name 2", "program": "program 1", "args": ["args 1"], "timeout": 0, "pause": 0}]
  }
]
`

func TestStepNamesNotUnique(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError error
	}{
		{name: "duplicate step names", json: StepNamesNotUnique, wantError: errs.ErrDuplicateStepName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, validationErrors, _ := DecodeSchedule(r)

			found := false
			for _, ve := range validationErrors {
				if errors.Is(ve, tt.wantError) {
					found = true
				} else {
					t.Fatalf("%s: unexpected validation error: %v (expected only %v). Full list: %v",
						tt.name, ve, tt.wantError, validationErrors)
				}
			}
			if !found {
				t.Fatalf("%s: missing expected error %v. Full list: %v", tt.name, tt.wantError, validationErrors)
			}
		})
	}
}

const WorkflowValid = `
[
  {
    "name": "workflow 1",
    "time": "10:35",
    "onFailure": "continue",
    "steps": [{"name": "step name", "program": "program", "args": ["args"], "timeout": 43200, "pause": 3600}]
  },
  {
    "name": "workflow 2",
    "onFailure": "continue",    
    "time": "11:35",
    "steps": [{"name": "step name 2", "program": "program 1", "args": ["args 1"], "timeout": 0, "pause": 0}]
  }
]
`

func TestWorkflowSteps_Valid(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "valid full config", json: WorkflowValid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, validationErrors, err := DecodeSchedule(r)

			if err != nil {
				t.Fatalf("%s: unexpected decode/IO error: %v, expected no error", tt.name, err)
			}

			if len(validationErrors) != 0 {
				t.Fatalf("%s: unexpected validation error(s): %v, expected no error", tt.name, validationErrors)
			}
		})
	}
}
