// Package schedule provides schedule layout and locking
package schedule

import (
	"git.sr.ht/~mytec/gosched/internal/errs"
	"git.sr.ht/~mytec/gosched/internal/policy"
	"git.sr.ht/~mytec/gosched/internal/types"
)

type Workflow struct {
	Name      string              `json:"name"`
	Time      string              `json:"time"`
	OnFailure *policy.FailureMode `json:"onFailure"`
	Retry     RetryConfig
	Steps     []Step `json:"steps"`
}

type RetryConfig struct {
	Attempts     types.ConfiguredInt `json:"attempts"`
	PauseSeconds types.ConfiguredInt `json:"pauseSeconds"`
}

type Step struct {
	Name    string              `json:"name"`
	Program string              `json:"program"`
	Args    string              `json:"args"`
	Timeout types.ConfiguredInt `json:"timeout"`
	Pause   types.ConfiguredInt `json:"pause"`
}

func (w Workflow) Validate() []error {
	var errorList []error

	vErrs := errs.ValidateWorkflowName(w.Name)
	if len(vErrs) != 0 {
		errorList = append(errorList, vErrs...)
	}

	vErrs = errs.ValidateWorkflowTime(w.Time)
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

		vErrs = errs.ValidateWorkflowStepArgs(steps.Args)
		if len(vErrs) != 0 {
			errorList = append(errorList, vErrs...)
		}
	}

	return errorList
}
