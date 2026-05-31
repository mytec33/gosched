// Package schedule provides schedule layout and locking
package schedule

import (
	"fmt"
	"strings"

	"git.sr.ht/~mytec/gosched/internal/errs"
	"git.sr.ht/~mytec/gosched/internal/policy"
	"git.sr.ht/~mytec/gosched/internal/types"
)

type Workflow struct {
	Name      string              `json:"name"`
	Time      *types.MinuteOfDay  `json:"time"`
	OnFailure *policy.FailureMode `json:"onFailure"`
	Retry     *RetryPolicy        `json:"retry"`
	Steps     []Step              `json:"steps"`
}

type RetryPolicy struct {
	NumberRetries int `json:"numberRetries"`
	PauseSeconds  int `json:"pauseSeconds"`
}

type Step struct {
	Name    string   `json:"name"`
	Program string   `json:"program"`
	Args    []string `json:"args"`
	Timeout int      `json:"timeout"`
	Pause   int      `json:"pause"`
}

func (w Workflow) Validate() []error {
	var errorList []error

	field := "workflow.name"
	errorList = append(errorList, validateStringValue(field, w.Name,
		errs.MaxWorkflowNameLength)...)

	if w.Time == nil {
		errorList = append(errorList, errs.ErrTimeFieldNotPresent)
	}

	// Invalid onFailure values are rejected during JSON decoding.
	// Validation checks that the field was provided.
	if w.OnFailure == nil {
		errorList = append(errorList, errs.ErrOnFailureRequired)
	} else if *w.OnFailure == policy.Retry && w.Retry == nil {
		errorList = append(errorList, errs.ErrRetryRequired)
	}

	if w.Retry != nil {
		if w.Retry.NumberRetries < 0 {
			errorList = append(errorList, errs.ErrRetryCountNegative)
		}

		if w.Retry.PauseSeconds < 0 {
			errorList = append(errorList, errs.ErrRetryPauseNegative)
		}
	}

	if len(w.Steps) == 0 {
		errorList = append(errorList, errs.ErrStepsMissing)
	}

	if len(w.Steps) > errs.MaxStepsCount {
		errorList = append(errorList, errs.ErrStepsCount)
	}

	for i, steps := range w.Steps {
		field := fmt.Sprintf("workflow.steps[%d].name", i+1)
		errorList = append(errorList, validateStringValue(field, steps.Name,
			errs.MaxWorkflowStepNameLength)...)

		if steps.Timeout < 0 {
			errorList = append(errorList, errs.ErrNegativeNumber)
		}

		if steps.Pause < 0 {
			errorList = append(errorList, errs.ErrNegativeNumber)
		}

		field = fmt.Sprintf("workflow.steps[%d].program", i+1)
		errorList = append(errorList, validateStringValue(field, steps.Program,
			errs.MaxWorkflowStepProgramLength)...)

		if len(steps.Args) > errs.MaxStepArgsCount {
			errorList = append(errorList, errs.ErrStepArgsTooMany)
		}

		totalArgsLength := 0
		for j, arg := range steps.Args {
			totalArgsLength += len(arg)

			field = fmt.Sprintf("workflow.steps[%d].args[%d]", i+1, j+1)
			errorList = append(errorList, validateStringValue(field, arg,
				errs.MaxStepArgLength)...)
		}

		if totalArgsLength > errs.MaxStepArgsTotalLength {
			errorList = append(errorList, errs.ErrStepArgsTotalLength)
		}
	}

	errorList = append(errorList, validateUniqueStepNames(w.Steps)...)

	return errorList
}

func validateStringValue(field string, s string, maxLength int) []error {
	var errorList []error
	trimmed := strings.TrimSpace(s)

	if s == "" {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, errs.ErrEmpty))
	} else if trimmed == "" {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, errs.ErrWhitespaceAll))
	} else if trimmed != s {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, errs.ErrWhitespaceLeadingOrTrailing))
	} else if len(trimmed) > maxLength {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, errs.ErrTooLong))
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

func WorkflowAbortsOnFailure(wf Workflow) bool {
	return wf.OnFailure != nil && *wf.OnFailure == policy.Abort
}

func WorkflowContinuesOnFailure(wf Workflow) bool {
	return wf.OnFailure != nil && *wf.OnFailure == policy.Continue
}

func WorkflowRetriesOnFailure(wf Workflow) bool {
	return wf.OnFailure != nil && *wf.OnFailure == policy.Retry
}
