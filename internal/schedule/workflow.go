// Package schedule provides schedule layout and locking
package schedule

import (
	"fmt"

	"git.sr.ht/~mytec/gosched/internal/errs"
	"git.sr.ht/~mytec/gosched/internal/policy"
	"git.sr.ht/~mytec/gosched/internal/types"
)

type Workflow struct {
	Name      string              `json:"name"`
	Time      types.MinuteOfDay   `json:"time"`
	OnFailure *policy.FailureMode `json:"onFailure"`
	Retry     RetryConfig
	Steps     []Step `json:"steps"`
}

type RetryConfig struct {
	NumberRetries types.ConfiguredInt `json:"numberRetries"`
	PauseSeconds  types.ConfiguredInt `json:"pauseSeconds"`
}

type Step struct {
	Name    string                `json:"name"`
	Program string                `json:"program"`
	Args    []types.ConfiguredArg `json:"args"`
	Timeout types.ConfiguredInt   `json:"timeout"`
	Pause   types.ConfiguredInt   `json:"pause"`
}

func (w Workflow) Validate() []error {
	var errorList []error

	vErrs := errs.ValidateWorkflowName(w.Name)
	if len(vErrs) != 0 {
		errorList = append(errorList, vErrs...)
	}

	// The JSON field onFailure isn't tested here because it's converted
	// from string -> enum and that boundary controls if it's valid or not
	if w.OnFailure == nil {
		errorList = append(errorList, errs.ErrOnFailureRequired)
	}

	if len(w.Steps) == 0 {
		errorList = append(errorList, errs.ValidationError{Field: "workflow.steps", Err: errs.ErrEmpty})
	}

	for _, steps := range w.Steps {
		vErrs := errs.ValidateWorkflowStepName(steps.Name)
		if len(vErrs) != 0 {
			errorList = append(errorList, vErrs...)
		}

		vErrs = errs.ValidateWorkflowStepProgram(steps.Program)
		if len(vErrs) != 0 {
			errorList = append(errorList, vErrs...)
		}
	}

	errorList = append(errorList, validateUniqueStepNames(w.Steps)...)

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
