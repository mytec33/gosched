package decode

import (
	"errors"
	"strings"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/workflow"
)

func TestDecode_InvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "empty", json: ``},
		{name: "missing closing brace", json: `{`},
		{name: "wrong top-level type object", json: `{}`},
		{name: "double empty opjects", json: `{}{}`},
		{name: "unknown field", json: `[{"foo": "bar"}]"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)

			_, errorList, err := DecodeWorkflowFile(r)
			if err == nil {
				t.Fatal("expected error, got no error")
			} else if !errors.Is(err, ErrDecodeWorkflow) {
				t.Fatalf("%v: expected ErrDecodeSchedule, got %v", tt.name, err)
			}

			if len(errorList) > 0 {
				t.Fatalf("%v: expected no validation errors, got %v", tt.name, errorList)
				t.Fatalf("%v", errorList)
			}
		})
	}
}

// Demonstrate a single workflow with a single step. This is the simplest configuration.
// Note, this demonstrates retry: {} is optional.
const validOneWorkflowOneStep = `
[
  {
    "name": "Workflow 1",
    "trigger": { "every": "1d", "beginAt": "10:35" },
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

const validFullExample = `
[
  {
    "name": "Workflow 1",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "abort",      
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": ["--sleep", "10", "--role", "daily-slot-ratings"]
      }
    ]
  },
  {
    "name": "Workflow 2",
    "trigger": { "every": "1m", "beginAt": "08:35" },
    "onFailure": "continue",      
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
  },
  {
    "name": "Workflow 3",
    "trigger": { "every": "10h", "beginAt": "7:35" },
    "retry": {
      "numberRetries": 1,
      "pauseSeconds": 30
    },
    "onFailure": "retry",      
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

func TestDecode_ValidInput(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "One work flow, one step", json: validOneWorkflowOneStep},
		{name: "Full example", json: validFullExample},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)

			_, errorList, err := DecodeWorkflowFile(r)
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

// MinuteOfDay parsing is tested in types; this fixture proves DecodeWorkflows
// routes trigger.beginAt through that boundary and surfaces its error.
const InvalidWorkflowTimeEmpty = `
[
  {
    "name": "foo",
    "trigger": {"every": "1d", "beginAt": ""},
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

// Retry validation is tested at the WorkflowRaw level; this fixture proves
// DecodeWorkflows returns workflow validation errors from the JSON path.
const InvalidRetryConfiguration = `
[
  {
    "name": "foo",
    "trigger": {"every": "1d", "beginAt": "06:30"},
    "onFailure": "retry",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

func TestDecodeRejectsInvalidTypedField(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError error
	}{
		{name: "time empty", json: InvalidWorkflowTimeEmpty, wantError: workflow.ErrTimeFormatInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, errorList, err := DecodeWorkflowFile(r)

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

func TestDecodeReportsWorkflowValidationError(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError error
	}{
		{name: "invalid retry config", json: InvalidRetryConfiguration, wantError: workflow.ErrRetryRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, validationErrors, err := DecodeWorkflowFile(r)

			if err != nil {
				t.Fatalf("%s: unexpected decode/system error: %v", tt.name, err)
			}

			if !hasError(validationErrors, tt.wantError) {
				t.Fatalf("%s: expected validation error %v, got %v", tt.name, tt.wantError, validationErrors)
			}
		})
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
