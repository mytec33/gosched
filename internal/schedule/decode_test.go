package schedule

import (
	"errors"
	"strings"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/errs"
)

const validOneWorkflowOneStep = `
[
  {
    "name": "Workflow 1",
    "time": "10:35",
    "onFailure": "abort",      
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "10", "--role", "daily-slot-ratings"]
      }
    ]
  }
]
`

const validOneWorkflowTwoSteps = `
[
  {
    "name": "Workflow 1",
    "time": "10:35",
    "onFailure": "abort",      
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "10", "--role", "daily-slot-ratings"]
      },
      {
        "name": "modified",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "50", "--role", "modified-slot-ratings"]
      }
    ]
  }
]
`

const validTwoWorkflows = `
[
  {
    "name": "Workflow 1",
    "time": "10:35",
    "onFailure": "abort",      
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "10", "--role", "daily-slot-ratings"]
      },
      {
        "name": "modified",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "50", "--role", "modified-slot-ratings"]
      }
    ]
  },
  {
    "name": "Workflow 2",
    "time": "10:35",
    "onFailure": "abort",      
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "5", "--role", "daily-table-ratings"]
      },
      {
        "name": "modified",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "25", "--role", "modified-table-ratings"]
      }
    ]
  }
]
`

func TestDecode_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "empty", json: ``},
		{name: "syntax error 1", json: `{`},
		{name: "syntax error 2", json: `{}{}`},
		{name: "unknown field", json: `[{"foo": "bar"}]"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)

			_, errorList, err := DecodeSchedule(r)
			if err == nil {
				t.Fatal("expected error, got no error")
			} else if !errors.Is(err, ErrDecodeSchedule) {
				t.Fatalf("%v: expected ErrDecodeSchedule, got %v", tt.name, err)
			}

			if len(errorList) > 0 {
				t.Fatalf("%v: expected no validation errors, got %v", tt.name, errorList)
				t.Fatalf("%v", errorList)
			}
		})
	}
}

func TestDecode_ValidInput(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "One work flow, one step", json: validOneWorkflowOneStep},
		{name: "One work flow, two steps", json: validOneWorkflowTwoSteps},
		{name: "Two work flows", json: validTwoWorkflows},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)

			_, errorList, err := DecodeSchedule(r)
			if err != nil {
				t.Fatalf("%v: expected no error, got %v", tt.name, err)
			}

			if len(errorList) > 0 {
				t.Fatalf("%v: expected no validation errors, got %v", tt.name, errorList)
				t.Fatalf("%v", errorList)
			}
		})
	}
}

const WorkflowTimeEmpty = `
[
  {
    "name": "foo",
    "time": "",
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowTimeBadHour = `
[
  {
    "name": "foo",
    "time": "99:35",
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowTimeBadMinute = `
[
  {
    "name": "foo",
    "time": "10:123",
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowTimeMissingColon = `
[
  {
    "name": "name",
    "time": "1001",
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
  `

const WorkflowTimeWhitespace = `
[
  {
    "name": "name",
    "time": " ",
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

func TestWorkflowTimes_Invalid(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError error
	}{
		{name: "time empty", json: WorkflowTimeEmpty, wantError: errs.ErrInvalidTimeFormat},
		{name: "time bad hour", json: WorkflowTimeBadHour, wantError: errs.ErrInvalidTimeFormat},
		{name: "time bad minute", json: WorkflowTimeBadMinute, wantError: errs.ErrInvalidTimeFormat},
		{name: "time missing colon", json: WorkflowTimeMissingColon, wantError: errs.ErrInvalidTimeFormat},
		{name: "time bad whitespace", json: WorkflowTimeWhitespace, wantError: errs.ErrInvalidTimeFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, errorList, err := DecodeSchedule(r)

			if err == nil {
				t.Fatalf("%s: expected decode/system error: got %v, want %v", tt.name, err, tt.wantError)
			}

			if !errors.Is(err, tt.wantError) {
				t.Fatalf("%s: got %v, want %v", tt.name, err, tt.wantError)
			}

			if len(errorList) > 0 {
				t.Fatalf("%v: expected no validation errors, got %v", tt.name, errorList)
			}

		})
	}
}

const InvalidTwoDuplicateNamedWorkflows = `
[
  {
    "name": "Workflow 1",
    "time": "10:35",
    "onFailure": "abort",
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "10", "--role", "daily-slot-ratings"]
      },
      {
        "name": "modified",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "50", "--role", "modified-slot-ratings"]
      }
    ]
  },
  {
    "name": "Workflow 1",
    "time": "10:35",
    "onFailure": "abort",    
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "5", "--role", "daily-table-ratings"]
      }
    ]
  }
]
`

const InvalidTwoDuplicateNamedWorkflowsLeadingWhitespace = `
[
  {
    "name": "Workflow 1",
    "time": "10:35",
    "onFailure": "abort",
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "10", "--role", "daily-slot-ratings"]
      },
      {
        "name": "modified",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "50", "--role", "modified-slot-ratings"]
      }
    ]
  },
  {
    "name": "\tWorkflow 1",
    "time": "10:35",
    "onFailure": "abort",    
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "5", "--role", "daily-table-ratings"]
      }
    ]
  }
]
`

const InvalidTwoDuplicateNamedWorkflowsMixedCase = `
[
  {
    "name": "Workflow 1",
    "time": "10:35",
    "onFailure": "abort",
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "10", "--role", "daily-slot-ratings"]
      },
      {
        "name": "modified",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "50", "--role", "modified-slot-ratings"]
      }
    ]
  },
  {
    "name": "workFloW 1",
    "time": "10:35",
    "onFailure": "abort",    
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "5", "--role", "daily-table-ratings"]
      }
    ]
  }
]
`

func TestDecodeDuplicateWorkflowNames(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "duplicate workflow names - identical",
			json: InvalidTwoDuplicateNamedWorkflows,
		},
		{
			name: "duplicate workflow names - leading whitespace",
			json: InvalidTwoDuplicateNamedWorkflowsLeadingWhitespace,
		},
		{
			name: "duplicate workflow names - mixed case",
			json: InvalidTwoDuplicateNamedWorkflowsMixedCase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)

			_, errorList, err := DecodeSchedule(r)

			if err != nil {
				t.Fatalf("%v: expected no error, got %v", tt.name, err)
			}

			found := false
			for _, e := range errorList {
				if errors.Is(e, errs.ErrDuplicateWorkflowName) {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("%v: expected %v", tt.name, errs.ErrDuplicateWorkflowName)
			}
		})
	}
}
