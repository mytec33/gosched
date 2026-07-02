package workflow

import (
	"errors"
	"fmt"
	"strings"
)

const (
	MaxStepArgLength             int = 256
	MaxStepArgsCount             int = 64
	MaxStepArgsTotalLength       int = 4096
	MaxStepsCount                int = 32
	MaxWorkflowNameLength        int = 256
	MaxWorkflowStepProgramLength int = 256
	MaxWorkflowStepNameLength    int = 256
)

var (
	ErrFieldEmpty            = errors.New("cannot be empty")
	ErrFieldTooLong          = errors.New("too long")
	ErrFieldWhitespaceOnly   = errors.New("cannot be all whitespace")
	ErrFieldWhitespacePadded = errors.New("leading or trailing whitespace")
	ErrNumberNegative        = errors.New("number cannot be negative, must be zero (indefinite) or greater")

	ErrOnFailureRequired  = errors.New("onFailure field required")
	ErrRetryCountNegative = errors.New("retry count must be 0 or greater")
	ErrRetryPauseNegative = errors.New("retry pause seconds must be 0 or greater")
	ErrRetryRequired      = errors.New("retry config required when onFailure is set to retry")

	ErrStepDuplicateName           = errors.New("step name is a duplicate")
	ErrStepArgsCountExceeded       = fmt.Errorf("too many args provided: max is %d", MaxStepArgsCount)
	ErrStepArgsTotalLengthExceeded = fmt.Errorf("total length of all args exceeds limit: max is %d", MaxStepArgsTotalLength)
	ErrStepCountExceeded           = fmt.Errorf("too many steps in workflow: max is %d", MaxStepsCount)
	ErrStepsRequired               = fmt.Errorf("steps are required")

	ErrTriggerRequired        = fmt.Errorf("trigger field is required")
	ErrTriggerBeginAtRequired = errors.New("trigger beginAt field is required")
	ErrTriggerEveryRequired   = errors.New("trigger every field is required")
)

type WorkflowRaw struct {
	Name      string       `json:"name"`
	Trigger   *Trigger     `json:"trigger"`
	OnFailure *FailureMode `json:"onFailure"`
	Retry     *RetryPolicy `json:"retry"`
	Steps     []Step       `json:"steps"`
}

func (raw WorkflowRaw) Validate() (Workflow, []error) {
	var workflow Workflow
	var errorList []error

	field := "workflow.name"
	errorList = append(errorList, validateStringValue(field, raw.Name,
		MaxWorkflowNameLength)...)
	workflow.Name = raw.Name

	if raw.Trigger == nil {
		errorList = append(errorList, ErrTriggerRequired)
	} else {
		if raw.Trigger.Every == nil {
			errorList = append(errorList, ErrTriggerEveryRequired)
		}

		if raw.Trigger.BeginAt == nil {
			errorList = append(errorList, ErrTriggerBeginAtRequired)
		}

		workflow.Trigger = *raw.Trigger
	}

	// Invalid onFailure values are rejected during JSON decoding.
	// Validation checks that the field was provided.
	if raw.OnFailure == nil {
		errorList = append(errorList, ErrOnFailureRequired)
	} else if *raw.OnFailure == Retry && raw.Retry == nil {
		errorList = append(errorList, ErrRetryRequired)
	} else {
		workflow.OnFailure = *raw.OnFailure
	}

	if raw.Retry != nil {
		workflow.Retry = raw.Retry

		if raw.Retry.NumberRetries < 0 {
			errorList = append(errorList, ErrRetryCountNegative)
		}

		if raw.Retry.PauseSeconds < 0 {
			errorList = append(errorList, ErrRetryPauseNegative)
		}
	}

	if len(raw.Steps) == 0 {
		errorList = append(errorList, ErrStepsRequired)
	}

	if len(raw.Steps) > MaxStepsCount {
		errorList = append(errorList, ErrStepCountExceeded)
	}

	for i, steps := range raw.Steps {
		field := fmt.Sprintf("workflow.steps[%d].name", i+1)
		errorList = append(errorList, validateStringValue(field, steps.Name,
			MaxWorkflowStepNameLength)...)

		if steps.Pause.Duration() < 0 {
			errorList = append(errorList, ErrNumberNegative)
		}

		field = fmt.Sprintf("workflow.steps[%d].program", i+1)
		errorList = append(errorList, validateStringValue(field, steps.Program,
			MaxWorkflowStepProgramLength)...)

		field = fmt.Sprintf("workflow.steps[%d].args", i+1)
		errorList = append(errorList, validateStepArgs(field, steps.Args)...)
	}
	// Steps (and the entire schedule) are treated as immutable config after
	// validation so args won't need a deep copy
	workflow.Steps = raw.Steps

	errorList = append(errorList, validateUniqueStepNames(raw.Steps)...)

	return workflow, errorList
}

func validateStepArgs(field string, args []string) []error {
	var errorList []error

	if len(args) > MaxStepArgsCount {
		errorList = append(errorList, ErrStepArgsCountExceeded)
	}

	totalArgsLength := 0
	for j, arg := range args {
		totalArgsLength += len(arg)

		argField := fmt.Sprintf("%s[%d]", field, j+1)
		errorList = append(errorList, validateStringValue(argField, arg, MaxStepArgLength)...)
	}

	if totalArgsLength > MaxStepArgsTotalLength {
		errorList = append(errorList, ErrStepArgsTotalLengthExceeded)
	}

	return errorList
}

func validateStringValue(field string, s string, maxLength int) []error {
	var errorList []error
	trimmed := strings.TrimSpace(s)

	if s == "" {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, ErrFieldEmpty))
	} else if trimmed == "" {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, ErrFieldWhitespaceOnly))
	} else if trimmed != s {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, ErrFieldWhitespacePadded))
	} else if len(trimmed) > maxLength {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, ErrFieldTooLong))
	}

	return errorList
}

func validateUniqueStepNames(steps []Step) []error {
	stepNames := make(map[string]struct{})
	var errors []error

	for _, step := range steps {
		_, exists := stepNames[step.Name]
		if exists {
			errors = append(errors, fmt.Errorf("%w: %s", ErrStepDuplicateName, step.Name))
		} else {
			stepNames[step.Name] = struct{}{}
		}
	}

	return errors
}
