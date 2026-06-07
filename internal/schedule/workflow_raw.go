package schedule

import (
	"fmt"
	"strings"

	"git.sr.ht/~mytec/gosched/internal/errs"
	"git.sr.ht/~mytec/gosched/internal/policy"
	"git.sr.ht/~mytec/gosched/internal/types"
)

type WorkflowRaw struct {
	Name      string              `json:"name"`
	Trigger   *types.Trigger      `json:"trigger"`
	OnFailure *policy.FailureMode `json:"onFailure"`
	Retry     *RetryPolicy        `json:"retry"`
	Steps     []Step              `json:"steps"`
}

func (raw WorkflowRaw) Validate() (Workflow, []error) {
	var workflow Workflow
	var errorList []error

	field := "workflow.name"
	errorList = append(errorList, validateStringValue(field, raw.Name,
		errs.MaxWorkflowNameLength)...)
	workflow.Name = raw.Name

	if raw.Trigger == nil {
		errorList = append(errorList, errs.ErrTriggerRequired)
	} else {
		if raw.Trigger.Every == nil {
			errorList = append(errorList, errs.ErrTriggerEveryRequired)
		}

		if raw.Trigger.BeginAt == nil {
			errorList = append(errorList, errs.ErrTriggerBeginAtRequired)
		}

		workflow.Trigger = *raw.Trigger
	}

	// Invalid onFailure values are rejected during JSON decoding.
	// Validation checks that the field was provided.
	if raw.OnFailure == nil {
		errorList = append(errorList, errs.ErrOnFailureRequired)
	} else if *raw.OnFailure == policy.Retry && raw.Retry == nil {
		errorList = append(errorList, errs.ErrRetryRequired)
	} else {
		workflow.OnFailure = *raw.OnFailure
	}

	if raw.Retry != nil {
		workflow.Retry = raw.Retry

		if raw.Retry.NumberRetries < 0 {
			errorList = append(errorList, errs.ErrRetryCountNegative)
		}

		if raw.Retry.PauseSeconds < 0 {
			errorList = append(errorList, errs.ErrRetryPauseNegative)
		}
	}

	if len(raw.Steps) == 0 {
		errorList = append(errorList, errs.ErrStepsRequired)
	}

	if len(raw.Steps) > errs.MaxStepsCount {
		errorList = append(errorList, errs.ErrStepCountExceeded)
	}

	for i, steps := range raw.Steps {
		field := fmt.Sprintf("workflow.steps[%d].name", i+1)
		errorList = append(errorList, validateStringValue(field, steps.Name,
			errs.MaxWorkflowStepNameLength)...)

		if steps.Timeout < 0 {
			errorList = append(errorList, errs.ErrNumberNegative)
		}

		if steps.Pause < 0 {
			errorList = append(errorList, errs.ErrNumberNegative)
		}

		field = fmt.Sprintf("workflow.steps[%d].program", i+1)
		errorList = append(errorList, validateStringValue(field, steps.Program,
			errs.MaxWorkflowStepProgramLength)...)

		if len(steps.Args) > errs.MaxStepArgsCount {
			errorList = append(errorList, errs.ErrStepArgsCountExceeded)
		}

		totalArgsLength := 0
		for j, arg := range steps.Args {
			totalArgsLength += len(arg)

			field = fmt.Sprintf("workflow.steps[%d].args[%d]", i+1, j+1)
			errorList = append(errorList, validateStringValue(field, arg,
				errs.MaxStepArgLength)...)
		}

		if totalArgsLength > errs.MaxStepArgsTotalLength {
			errorList = append(errorList, errs.ErrStepArgsTotalLengthExceeded)
		}
	}
	// Steps (and the entire schedule) are treated as immutable config after
	// validation so args won't need a deep copy
	workflow.Steps = raw.Steps

	errorList = append(errorList, validateUniqueStepNames(raw.Steps)...)

	return workflow, errorList
}

func validateStringValue(field string, s string, maxLength int) []error {
	var errorList []error
	trimmed := strings.TrimSpace(s)

	if s == "" {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, errs.ErrFieldEmpty))
	} else if trimmed == "" {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, errs.ErrFieldWhitespaceOnly))
	} else if trimmed != s {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, errs.ErrFieldWhitespacePadded))
	} else if len(trimmed) > maxLength {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, errs.ErrFieldTooLong))
	}

	return errorList
}

func validateUniqueStepNames(steps []Step) []error {
	stepNames := make(map[string]struct{})
	var errors []error

	for _, step := range steps {
		_, exists := stepNames[step.Name]
		if exists {
			errors = append(errors, fmt.Errorf("%w: %s", errs.ErrDuplicateStepName, step.Name))
		} else {
			stepNames[step.Name] = struct{}{}
		}
	}

	return errors
}
