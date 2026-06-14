package schedule

import (
	"errors"
	"strings"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/types"
)

const WorkflowNameEmpty = `
[
  {
    "name": "",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowNameWhitespace = `
[
  {
    "name": "        ",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowNameWhitespaceLeading = `
[
  {
    "name": " name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowNameWhitespaceTrailing = `
[
  {
    "name": "name ",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowNameTooLong = `
[
  {
    "name": "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis,..",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "daily", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowNoSteps = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
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
		{name: "name empty", json: WorkflowNameEmpty, wantError: ErrFieldEmpty},
		{name: "name whitespace", json: WorkflowNameWhitespace, wantError: ErrFieldWhitespaceOnly},
		{name: "name whitespace leading", json: WorkflowNameWhitespaceLeading, wantError: ErrFieldWhitespacePadded},
		{name: "name whitespace trailing", json: WorkflowNameWhitespaceTrailing, wantError: ErrFieldWhitespacePadded},
		{name: "name too long", json: WorkflowNameTooLong, wantError: ErrFieldTooLong},
		{name: "missing steps", json: WorkflowNoSteps, wantError: ErrStepsRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, validationErrors, err := DecodeWorkflows(r)

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
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowStepNameWhitespace = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "   ", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowStepNameWhitespaceLeading = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": " leading", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowStepNameWhitespaceTrailing = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "trailing ", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowStepNameTooLong = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis,..", "program": "program", "args": ["args"]}]
  }
]
`

const WorkflowStepProgramEmpty = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "", "args": ["args"]}]
  }
]
`

const WorkflowStepProgramWhitespace = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "     ", "args": ["args"]}]
  }
]
`

const WorkflowStepProgramWhitespaceLeading = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": " foo", "args": ["args"]}]
  }
]
`

const WorkflowStepProgramWhitespaceTrailing = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "foo ", "args": ["args"]}]
  }
]
`

// 1. Tests what happens when the entire "trigger" block is missing entirely
const WorkflowMissingTriggerBlock = `
[
  {
    "name": "Missing Trigger Test",
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "foo", "args": ["args"]}]
  }
]
`

// 2. Tests when the "trigger" block exists, but the "every" field is missing
const WorkflowMissingTriggerEvery = `
[
  {
    "name": "Missing Every Test",
    "trigger": { "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "foo", "args": ["args"]}]
  }
]
`

// 3. Tests when the "trigger" block exists, but the "beginAt" field is missing
const WorkflowMissingTriggerBeginAt = `
[
  {
    "name": "Missing BeginAt Test",
    "trigger": { "every": "1d" },
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "foo", "args": ["args"]}]
  }
]
`

func TestWorkflowSteps_Invalid(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError error
	}{
		{name: "name empty", json: WorkflowStepNameEmpty, wantError: ErrFieldEmpty},
		{name: "name whitespace", json: WorkflowStepNameWhitespace, wantError: ErrFieldWhitespaceOnly},
		{name: "name whitespace leading", json: WorkflowStepNameWhitespaceLeading, wantError: ErrFieldWhitespacePadded},
		{name: "name whitespace trailing", json: WorkflowStepNameWhitespaceTrailing, wantError: ErrFieldWhitespacePadded},
		{name: "name too long", json: WorkflowStepNameTooLong, wantError: ErrFieldTooLong},
		{name: "program empty", json: WorkflowStepProgramEmpty, wantError: ErrFieldEmpty},
		{name: "program whitespace", json: WorkflowStepProgramWhitespace, wantError: ErrFieldWhitespaceOnly},
		{name: "program whitespace leading", json: WorkflowStepProgramWhitespaceLeading, wantError: ErrFieldWhitespacePadded},
		{name: "program whitespace trailing", json: WorkflowStepProgramWhitespaceTrailing, wantError: ErrFieldWhitespacePadded},

		{name: "trigger block missing", json: WorkflowMissingTriggerBlock, wantError: ErrTriggerRequired},
		{name: "trigger 'every' field missing", json: WorkflowMissingTriggerEvery, wantError: ErrTriggerEveryRequired},
		{name: "trigger 'beginAt' field missing", json: WorkflowMissingTriggerBeginAt, wantError: ErrTriggerBeginAtRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, validationErrors, err := DecodeWorkflows(r)

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

const StepNamesNotUnique = `
[
  {
    "name": "workflow 1",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",
    "steps": [
      {"name": "step name", "program": "program", "args": ["args"], "timeout": 43200, "pause": 3600},
      {"name": "step name", "program": "program", "args": ["args"], "timeout": 43200, "pause": 3600}      
    ]
  },
  {
    "name": "workflow 2",
    "onFailure": "continue",    
    "trigger": { "every": "1d", "beginAt": "11:35" },
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
		{name: "duplicate step names", json: StepNamesNotUnique, wantError: ErrStepDuplicateName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			_, validationErrors, _ := DecodeWorkflows(r)

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
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",
    "steps": [{"name": "step name", "program": "program", "args": ["args"], "timeout": 43200, "pause": 3600}]
  },
  {
    "name": "workflow 2",
    "onFailure": "continue",    
    "trigger": { "every": "1d", "beginAt": "11:35" },
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
			_, validationErrors, err := DecodeWorkflows(r)

			if err != nil {
				t.Fatalf("%s: unexpected decode/IO error: %v, expected no error", tt.name, err)
			}

			if len(validationErrors) != 0 {
				t.Fatalf("%s: unexpected validation error(s): %v, expected no error", tt.name, validationErrors)
			}
		})
	}
}

func TestWorkflowRetryConfiguredNumbers_Invalid(t *testing.T) {
	tests := []struct {
		name          string
		wf            WorkflowRaw
		expectedError error
	}{
		{
			name: "retry number negative",
			wf: WorkflowRaw{
				Name:  "workflow",
				Retry: &RetryPolicy{NumberRetries: -1},
				Steps: []Step{{Name: "step", Program: "program"}},
			},
			expectedError: ErrRetryCountNegative,
		},
		{
			name: "retry pause negative",
			wf: WorkflowRaw{
				Name:  "workflow",
				Retry: &RetryPolicy{PauseSeconds: -1},
				Steps: []Step{{Name: "step", Program: "program"}},
			},
			expectedError: ErrRetryPauseNegative,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, validationErrors := tt.wf.Validate()

			found := false
			for _, ve := range validationErrors {
				if errors.Is(ve, tt.expectedError) {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("%s: missing expected error %v. Full list: %v",
					tt.name, tt.expectedError, validationErrors)
			}
		})
	}
}

func TestWorkflowStepsConfiguredNumbers_Invalid(t *testing.T) {
	tests := []struct {
		name          string
		wf            WorkflowRaw
		expectedError error
	}{
		{
			name: "step timeout negative",
			wf: WorkflowRaw{
				Name:      "workflow",
				OnFailure: &types.Continue,
				Steps:     []Step{{Name: "step", Program: "program", Timeout: -1}},
			},
			expectedError: ErrNumberNegative,
		},
		{
			name: "step pause negative",
			wf: WorkflowRaw{
				Name:      "workflow",
				OnFailure: &types.Continue,
				Steps:     []Step{{Name: "step", Program: "program", Pause: -1}},
			},
			expectedError: ErrNumberNegative,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, validationErrors := tt.wf.Validate()

			found := false
			for _, ve := range validationErrors {
				if errors.Is(ve, ErrNumberNegative) {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("%s: missing expected error %v. Full list: %v", tt.name, ErrNumberNegative, validationErrors)
			}
		})
	}
}

const WorkflowStepArgsWhitespace = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "program", "args": ["         "]}]
  }
]
`

const WorkflowStepArgsWhitespaceLeading = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "program", "args": [" leading"]}]
  }
]
`

const WorkflowStepArgsWhitespaceTrailing = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "program", "args": ["trailing "]}]
  }
]
`

const WorkflowStepArgsTooLong = `
[
  {
    "name": "name",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "continue",    
    "steps": [{"name": "step name", "program": "program", "args": ["Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis,.."]}]
  }
]
`

func TestDecodeArgs_Invalid(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError error
	}{
		{name: "args whitespace", json: WorkflowStepArgsWhitespace, wantError: ErrFieldWhitespaceOnly},
		{name: "args whitespace leading", json: WorkflowStepArgsWhitespaceLeading, wantError: ErrFieldWhitespacePadded},
		{name: "args whitespace trailing", json: WorkflowStepArgsWhitespaceTrailing, wantError: ErrFieldWhitespacePadded},
		{name: "args too long", json: WorkflowStepArgsTooLong, wantError: ErrFieldTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)

			_, validationErrors, _ := DecodeWorkflows(r)

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

const MissingRetryConfigOnPolicyRetry = `
[
  {
    "name": "workflow 1",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "retry",
    "steps": [
      {"name": "step name", "program": "program", "args": ["args"], "timeout": 43200, "pause": 3600}     
    ]
  }
]
`

const ValidateRetryConfigOnPolicyRetry = `
[
  {
    "name": "workflow 1",
    "trigger": { "every": "1d", "beginAt": "10:35" },
    "onFailure": "retry",
	"retry": {
		"numberRetries": 0,
		"pauseSeconds": 60
	},
    "steps": [
      {"name": "step name", "program": "program", "args": ["args"], "timeout": 43200, "pause": 3600}     
    ]
  }
]
`

func TestMissingRetryConfigOnPolicyRetry(t *testing.T) {
	r := strings.NewReader(MissingRetryConfigOnPolicyRetry)

	_, validationErrors, err := DecodeWorkflows(r)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	if !hasError(validationErrors, ErrRetryRequired) {
		t.Fatalf("expected %v, got %v", ErrRetryRequired, validationErrors)
	}
}

func TestValidateRetryConfigOnPolicyRetry(t *testing.T) {
	r := strings.NewReader(ValidateRetryConfigOnPolicyRetry)

	_, validationErrors, err := DecodeWorkflows(r)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	if len(validationErrors) != 0 {
		t.Fatalf("unexpected validation errors: %v", validationErrors)
	}

	if hasError(validationErrors, ErrRetryRequired) {
		t.Fatalf("unexpected %v, got %v", ErrRetryRequired, validationErrors)
	}
}
