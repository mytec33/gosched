package workflow

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	MaxDisabledReasonLength      int           = 256
	MaxRetryCount                int           = 8
	MaxRetryPauseLimit           int           = 7200
	MaxStepArgLength             int           = 256
	MaxStepArgsCount             int           = 64
	MaxStepArgsTotalLength       int           = 4096
	MaxStepsCount                int           = 32
	MaxStepPauseDuration         time.Duration = time.Hour * 1
	MaxWorkflowNameLength        int           = 256
	MaxWorkflowStepProgramLength int           = 256
	MaxWorkflowStepNameLength    int           = 256
)

var (
	ErrEnabledRequired          = errors.New("enabled is a required field having a value of true or false")
	ErrDisabledReasonRequired   = errors.New("disabled reason required when enabled equals false")
	ErrDisabledReasonNotAllowed = errors.New("disabled reason not allowed if enabled equals true")

	ErrFieldEmpty            = errors.New("cannot be empty")
	ErrFieldTooLong          = errors.New("too long")
	ErrFieldWhitespaceOnly   = errors.New("cannot be all whitespace")
	ErrFieldWhitespacePadded = errors.New("leading or trailing whitespace")
	ErrNumberNegative        = errors.New("number cannot be negative, must be zero (indefinite) or greater")

	ErrOnFailureRequired  = errors.New("onFailure field required")
	ErrRetryCountNegative = fmt.Errorf("retry count must be 0 to %d", MaxRetryCount)
	ErrRetryCountTooLarge = fmt.Errorf("retry count must be 0 to %d", MaxRetryCount)
	ErrRetryPauseNegative = fmt.Errorf("retry pause seconds must be 0 to %d", MaxRetryPauseLimit)
	ErrRetryPauseTooLarge = fmt.Errorf("retry pause seconds must be 0 to %d", MaxRetryPauseLimit)
	ErrRetryRequired      = errors.New("retry config required when onFailure is set to retry")

	ErrStepDuplicateName           = errors.New("step name is a duplicate")
	ErrStepArgsCountExceeded       = fmt.Errorf("too many args provided: max is %d", MaxStepArgsCount)
	ErrStepArgsTotalLengthExceeded = fmt.Errorf("total length of all args exceeds limit: max is %d", MaxStepArgsTotalLength)
	ErrStepCountExceeded           = fmt.Errorf("too many steps in workflow: max is %d", MaxStepsCount)
	ErrPauseDurationTooLarge       = fmt.Errorf("pause duration must be 1 hour or less")
	ErrStepsRequired               = fmt.Errorf("steps are required")

	ErrTriggerRequired        = fmt.Errorf("trigger field is required")
	ErrTriggerBeginAtRequired = errors.New("trigger beginAt field is required")
	ErrTriggerEveryRequired   = errors.New("trigger every field is required")
)

type WorkflowRaw struct {
	Name           string       `json:"name"`
	Enabled        *bool        `json:"enabled"`
	DisabledReason *string      `json:"disabledReason"`
	Trigger        *TriggerRaw  `json:"trigger"`
	OnFailure      *FailureMode `json:"onFailure"`
	Retry          *RetryPolicy `json:"retry"`
	Steps          []StepRaw    `json:"steps"`
}

func (raw WorkflowRaw) Validate(sourceFile string) (Workflow, []error) {
	var workflow Workflow
	var errorList []error

	workflow.SourceFile = sourceFile

	field := "workflow.name"
	errorList = append(errorList, validateStringValue(field, raw.Name,
		MaxWorkflowNameLength)...)
	workflow.Name = raw.Name

	if raw.Enabled == nil {
		errorList = append(errorList, ErrEnabledRequired)
	} else {
		workflow.Enabled = *raw.Enabled

		if !*raw.Enabled {
			if raw.DisabledReason == nil {
				errorList = append(errorList, ErrDisabledReasonRequired)
			} else {
				field := "workflow.disabledReason"
				errs := validateStringValue(field, *raw.DisabledReason, MaxDisabledReasonLength)
				if len(errs) > 0 {
					errorList = append(errorList, errs...)
				} else {
					workflow.DisabledReason = *raw.DisabledReason
				}
			}
		} else if raw.DisabledReason != nil {
			errorList = append(errorList, ErrDisabledReasonNotAllowed)
		}
	}

	if raw.Trigger == nil {
		errorList = append(errorList, ErrTriggerRequired)
	} else {
		if raw.Trigger.BeginAt == nil {
			errorList = append(errorList, ErrTriggerBeginAtRequired)
		} else {
			workflow.Trigger.BeginAt = *raw.Trigger.BeginAt
		}

		if raw.Trigger.Every == nil {
			errorList = append(errorList, ErrTriggerEveryRequired)
		} else {
			workflow.Trigger.Every = *raw.Trigger.Every
		}
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
		workflow.Retry = *raw.Retry

		if raw.Retry.NumberRetries < 0 {
			errorList = append(errorList, ErrRetryCountNegative)
		}

		if raw.Retry.NumberRetries > MaxRetryCount {
			errorList = append(errorList, ErrRetryCountTooLarge)
		}

		if raw.Retry.PauseSeconds < 0 {
			errorList = append(errorList, ErrRetryPauseNegative)
		}

		if raw.Retry.PauseSeconds > MaxRetryPauseLimit {
			errorList = append(errorList, ErrRetryPauseTooLarge)
		}
	}

	if len(raw.Steps) == 0 {
		errorList = append(errorList, ErrStepsRequired)
	}

	if len(raw.Steps) > MaxStepsCount {
		errorList = append(errorList, ErrStepCountExceeded)
	}

	defaultTimeout := ConfigDuration{duration: 30 * time.Second}
	for i := range raw.Steps {
		steps := &raw.Steps[i]

		field := fmt.Sprintf("workflow.steps[%d].name", i+1)
		errorList = append(errorList, validateStringValue(field, steps.Name,
			MaxWorkflowStepNameLength)...)

		field = fmt.Sprintf("workflow.steps[%d].pause", i+1)
		if steps.Pause.Duration() > MaxStepPauseDuration {
			errorList = append(errorList, fmt.Errorf("%s: %w", field, ErrPauseDurationTooLarge))
		}

		field = fmt.Sprintf("workflow.steps[%d].program", i+1)
		errorList = append(errorList, validateStringValue(field, steps.Program,
			MaxWorkflowStepProgramLength)...)

		field = fmt.Sprintf("workflow.steps[%d].args", i+1)
		errorList = append(errorList, validateStepArgs(field, steps.Args)...)

		if steps.Timeout == nil {
			steps.Timeout = &defaultTimeout
		}
	}

	steps := make([]Step, 0, len(raw.Steps))
	for _, rawStep := range raw.Steps {
		steps = append(steps, Step{
			Name:    rawStep.Name,
			Program: rawStep.Program,
			Args:    rawStep.Args,
			Timeout: *rawStep.Timeout,
			Pause:   rawStep.Pause,
		})
	}
	workflow.Steps = steps

	errorList = append(errorList, validateUniqueStepNames(workflow.Steps)...)

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
