package decode

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/schedule"
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

			sourceFileName := "not used"
			_, errorList, err := DecodeWorkflows(r, sourceFileName)
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

			sourceFileName := "not used"
			_, errorList, err := DecodeWorkflows(r, sourceFileName)
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
			sourceFileName := "not used"
			_, errorList, err := DecodeWorkflows(r, sourceFileName)

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

// Required-field fixtures. Each decodes cleanly (valid JSON, valid typed
// fields) but fails WorkflowRaw.Validate on a single missing/empty required
// field, so the error surfaces in the validation slice rather than as a decode
// error. These pin the nil-pointer gate: a workflow missing a required field
// must never cross into the trusted set.
const triggerMissing = `
[
  {
    "name": "foo",
    "enabled": true,
    "onFailure": "continue",
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const triggerEmpty = `
[
  {
    "name": "foo",
    "enabled": true,
    "trigger": {},
    "onFailure": "continue",
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const everyMissing = `
[
  {
    "name": "foo",
    "enabled": true,
    "trigger": {"beginAt": "06:30"},
    "onFailure": "continue",
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const beginAtMissing = `
[
  {
    "name": "foo",
    "enabled": true,
    "trigger": {"every": "1d"},
    "onFailure": "continue",
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const onFailureMissing = `
[
  {
    "name": "foo",
    "enabled": true,
    "trigger": {"every": "1d", "beginAt": "06:30"},
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const enabledMissing = `
[
  {
    "name": "foo",
    "trigger": {"every": "1d", "beginAt": "06:30"},
    "onFailure": "continue",
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const stepsEmpty = `
[
  {
    "name": "foo",
    "enabled": true,
    "trigger": {"every": "1d", "beginAt": "06:30"},
    "onFailure": "continue",
    "steps": []
  }
]
`

func TestDecodeReportsWorkflowValidationError(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError error
	}{
		{name: "invalid retry config", json: InvalidRetryConfiguration, wantError: workflow.ErrRetryRequired},
		{name: "trigger missing", json: triggerMissing, wantError: workflow.ErrTriggerRequired},
		{name: "trigger empty", json: triggerEmpty, wantError: workflow.ErrTriggerEveryRequired}, // also BeginAt
		{name: "every missing", json: everyMissing, wantError: workflow.ErrTriggerEveryRequired},
		{name: "beginAt missing", json: beginAtMissing, wantError: workflow.ErrTriggerBeginAtRequired},
		{name: "onFailure missing", json: onFailureMissing, wantError: workflow.ErrOnFailureRequired},
		{name: "enabled missing", json: enabledMissing, wantError: workflow.ErrEnabledRequired},
		{name: "steps empty", json: stepsEmpty, wantError: workflow.ErrStepsRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			sourceFileName := "not used"
			wfs, validationErrors, err := DecodeWorkflows(r, sourceFileName)

			if err != nil {
				t.Fatalf("%s: unexpected decode/system error: %v", tt.name, err)
			}

			if len(wfs) != 0 {
				t.Fatalf("%s: expected zero trusted workflows, got %d", tt.name, len(wfs))
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

	sourceFile := "not used"
	_, errorList, err := DecodeWorkflows(r, sourceFile)
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

// disabledReason: This logic pairs two fields whose
// validity depends on each other, so it is tested as a matrix rather than
// isolated cases. The disabled-with-reason row also asserts the reason is
// carried through into the trusted workflow, not merely accepted.
const disabledWithReason = `
[
  {
    "name": "foo",
    "enabled": false,
    "disabledReason": "under maintenance",
    "trigger": {"every": "1d", "beginAt": "06:30"},
    "onFailure": "continue",
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const disabledNoReason = `
[
  {
    "name": "foo",
    "enabled": false,
    "trigger": {"every": "1d", "beginAt": "06:30"},
    "onFailure": "continue",
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const enabledWithReason = `
[
  {
    "name": "foo",
    "enabled": true,
    "disabledReason": "should not be here",
    "trigger": {"every": "1d", "beginAt": "06:30"},
    "onFailure": "continue",
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

func TestDecodeEnabledDisabledMatrix(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		wantErr    error  // nil means the workflow must be accepted
		wantReason string // checked only when wantErr is nil
	}{
		{name: "disabled with reason is valid", json: disabledWithReason, wantErr: nil, wantReason: "under maintenance"},
		{name: "disabled without reason", json: disabledNoReason, wantErr: workflow.ErrDisabledReasonRequired},
		{name: "enabled with reason not allowed", json: enabledWithReason, wantErr: workflow.ErrDisabledReasonNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			sourceFile := "not used"
			wfs, validationErrors, err := DecodeWorkflows(r, sourceFile)

			if err != nil {
				t.Fatalf("%s: unexpected decode/system error: %v", tt.name, err)
			}

			if tt.wantErr != nil {
				if len(wfs) != 0 {
					t.Fatalf("%s: expected zero trusted workflows, got %d", tt.name, len(wfs))
				}
				if !hasError(validationErrors, tt.wantErr) {
					t.Fatalf("%s: expected validation error %v, got %v", tt.name, tt.wantErr, validationErrors)
				}
				return
			}

			if len(validationErrors) > 0 {
				t.Fatalf("%s: expected no validation errors, got %v", tt.name, validationErrors)
			}
			if len(wfs) != 1 {
				t.Fatalf("%s: expected one trusted workflow, got %d", tt.name, len(wfs))
			}
			if wfs[0].Enabled {
				t.Fatalf("%s: expected workflow to be disabled", tt.name)
			}
			if wfs[0].DisabledReason != tt.wantReason {
				t.Fatalf("%s: expected disabled reason %q, got %q", tt.name, tt.wantReason, wfs[0].DisabledReason)
			}
		})
	}
}

// a file mixing one valid and one invalid workflow must yield no
// schedule at all. Validation errors are aggregated and no partial trusted set
// is returned, so a single bad workflow can never let a half-built schedule
// through.
const oneValidOneInvalid = `
[
  {
    "name": "good",
    "enabled": true,
    "trigger": {"every": "1h", "beginAt": "00:00"},
    "onFailure": "abort",
    "steps": [{"name": "s", "program": "p"}]
  },
  {
    "name": "bad",
    "enabled": true,
    "trigger": {"beginAt": "01:00"},
    "onFailure": "abort",
    "steps": [{"name": "s", "program": "p"}]
  }
]
`

func TestDecodePartialFailureYieldsNoSchedule(t *testing.T) {
	r := strings.NewReader(oneValidOneInvalid)
	sourceFile := "not used"
	wfs, validationErrors, err := DecodeWorkflows(r, sourceFile)

	if err != nil {
		t.Fatalf("unexpected decode/system error: %v", err)
	}

	if len(wfs) != 0 {
		t.Fatalf("expected zero trusted workflows (no partial schedule), got %d", len(wfs))
	}

	if !hasError(validationErrors, workflow.ErrTriggerEveryRequired) {
		t.Fatalf("expected %v, got %v", workflow.ErrTriggerEveryRequired, validationErrors)
	}
}

// a positive end-to-end case that asserts expansion behavior, not just the
// absence of errors. An hourly trigger from midnight must expand to exactly
// 24 minute slots, be findable by name, and populate the byMinute index at the
// expected minutes only.
const validHourlyFromMidnight = `
[
  {
    "name": "test-hourly",
    "enabled": true,
    "trigger": {"every": "1h", "beginAt": "00:00"},
    "onFailure": "abort",
    "steps": [{"name": "s", "program": "p"}]
  }
]
`

func TestDecodeToReadySchedule(t *testing.T) {
	r := strings.NewReader(validHourlyFromMidnight)
	sourceFile := "not used"
	wfs, validationErrors, err := DecodeWorkflows(r, sourceFile)

	if err != nil {
		t.Fatalf("unexpected decode/system error: %v", err)
	}
	if len(validationErrors) > 0 {
		t.Fatalf("expected no validation errors, got %v", validationErrors)
	}

	sched, errorList := schedule.New(wfs)
	if len(errorList) > 0 {
		t.Fatalf("expected no errors, got %v\n", errorList)
	}

	if _, err := sched.GetWorkflowByName("test-hourly"); err != nil {
		t.Fatalf("expected to find workflow by name: %v", err)
	}

	// Hourly from 00:00 lands on minute 0, 60, 120 ... 1380: 24 slots.
	populated := 0
	for m := workflow.MinuteOfDay(0); m < workflow.MinuteOfDay(1440); m++ {
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

func TestDecodeWorkflows_UnknownTriggerField(t *testing.T) {
	jsonInput := `[
		{
			"name": "workflow",
			"enabled": true,
			"trigger": {
				"every": "1h",
				"beginAt": "10:35",
				"evrey": "2h"
			},
			"onFailure": "continue",
			"steps": [
				{
					"name": "step",
					"program": "program"
				}
			]
		}
	]`

	_, validationErrors, err := DecodeWorkflows(strings.NewReader(jsonInput),
		"source file not ")

	if !errors.Is(err, ErrDecodeWorkflow) {
		t.Fatalf("expected %v, got %T: %v", ErrDecodeWorkflow, err, err)
	}

	if len(validationErrors) > 0 {
		t.Fatalf("expected no validation errors, got %v", validationErrors)
	}
}
