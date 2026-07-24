package decode

import (
	"encoding/json"
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
		{name: "double empty objects", json: `{}{}`},
		{name: "unknown field", json: `[{"foo": "bar"}]"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)

			_, errorList, err := DecodeWorkflows(r)
			if err == nil {
				t.Fatal("expected error, got no error")
			} else if !errors.Is(err, ErrDecodeWorkflow) {
				t.Fatalf("%v: expected ErrDecodeWorkflow, got %v", tt.name, err)
			}

			if len(errorList) > 0 {
				t.Fatalf("%v: expected no validation errors, got %v", tt.name, errorList)
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
	"enabled": true,	
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
	"enabled": true,	
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
	"enabled": true,	
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
	"enabled": true,
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

			_, errorList, err := DecodeWorkflows(r)
			if err != nil {
				t.Fatalf("%v: expected no error, got %v", tt.name, err)
			}

			if len(errorList) > 0 {
				t.Fatalf("%v: expected no validation errors, got %v", tt.name, errorList)
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
	"enabled": true,	
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
	"enabled": true,
    "trigger": {"every": "1d", "beginAt": "06:30"},
    "onFailure": "retry",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

// Cadence parsing is tested in types; this fixture proves DecodeWorkflows
// routes trigger.every through that boundary and surfaces its error.
const InvalidWorkflowCadenceUnit = `
[
  {
    "name": "foo",
	"enabled": true,
    "trigger": {"every": "12s", "beginAt": "06:30"},
    "onFailure": "continue",
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

// FailureMode validation is tested in types; this fixture proves
// DecodeWorkflows routes onFailure through that boundary and surfaces its error.
const InvalidWorkflowOnFailureUnknown = `
[
  {
    "name": "foo",
	"enabled": true,
    "trigger": {"every": "1d", "beginAt": "06:30"},
    "onFailure": "stop",
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
		{name: "cadence unit invalid", json: InvalidWorkflowCadenceUnit, wantError: workflow.ErrCadenceUnitInvalid},
		{name: "onFailure unknown", json: InvalidWorkflowOnFailureUnknown, wantError: workflow.ErrOnFailureInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, errorList, err := DecodeWorkflows(r)

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
			_, validationErrors, err := DecodeWorkflows(r)

			if err != nil {
				t.Fatalf("%s: unexpected decode/system error: %v", tt.name, err)
			}

			if !hasError(validationErrors, tt.wantError) {
				t.Fatalf("%s: expected validation error %v, got %v", tt.name, tt.wantError, validationErrors)
			}
		})
	}
}

// A JSON value of the wrong type for a typed field (number where a string is
// required) must fail decoding cleanly rather than panic. This is the only
// test proving that property; the per-type unmarshal tests were consolidated
// here per the layer-ownership rule.
const InvalidWorkflowTimeNotString = `
[
  {
    "name": "foo",
	"enabled": true,
    "trigger": {"every": "1d", "beginAt": 123},
    "onFailure": "continue",
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

func TestDecode_NonStringTypedField(t *testing.T) {
	r := strings.NewReader(InvalidWorkflowTimeNotString)

	_, errorList, err := DecodeWorkflows(r)
	if err == nil {
		t.Fatal("expected decode error, got none")
	}

	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &typeErr) {
		t.Fatalf("expected json.UnmarshalTypeError, got %v", err)
	}

	if len(errorList) > 0 {
		t.Fatalf("expected no validation errors, got %v", errorList)
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
