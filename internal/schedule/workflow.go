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
	Time      types.MinuteOfDay   `json:"time"`
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

	errorList = append(errorList, validateRequiredName(w.Name, errs.MaxWorkflowNameLength)...)

	// The JSON field onFailure isn't tested here because it's converted
	// from string -> enum and that boundary controls if it's valid or not
	if w.OnFailure == nil {
		errorList = append(errorList, errs.ErrOnFailureRequired)
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
		errorList = append(errorList, errs.ErrEmpty)
	}

	for _, steps := range w.Steps {
		errorList = append(errorList, validateRequiredName(steps.Name,
			errs.MaxWorkflowStepNameLength)...)

		if steps.Timeout < 0 {
			errorList = append(errorList, errs.ErrNegativeNumber)
		}

		if steps.Pause < 0 {
			errorList = append(errorList, errs.ErrNegativeNumber)
		}

		errorList = append(errorList, validateRequiredName(steps.Program,
			errs.MaxWorkflowStepProgramLength)...)

		for _, arg := range steps.Args {
			errorList = append(errorList, validateRequiredName(arg,
				errs.MaxProgramArgsLength)...)
		}

	}

	errorList = append(errorList, validateUniqueStepNames(w.Steps)...)

	return errorList
}

func validateRequiredName(s string, maxLength int) []error {
	var errorList []error
	trimmed := strings.TrimSpace(s)

	if s == "" {
		errorList = append(errorList, errs.ErrEmpty)
	} else if trimmed == "" {
		errorList = append(errorList, errs.ErrWhitespaceAll)
	} else if trimmed != s {
		errorList = append(errorList, errs.ErrWhitespaceLeadingOrTrailing)
	} else if len(trimmed) > maxLength {
		errorList = append(errorList, errs.ErrTooLong)
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
