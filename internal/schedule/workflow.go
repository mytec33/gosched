// Package schedule provides schedule layout and locking
package schedule

import (
	"strconv"

	"git.sr.ht/~mytec/gosched/internal/errs"
)

type Workflow struct {
	Name  string `json:"name"`
	Time  string `json:"time"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Name    string `json:"name"`
	Program string `json:"program"`
	Args    string `json:"args"`
	Timeout int    `json:"timeout"`
	Pause   int    `json:"pause"`
}

func (w Workflow) Validate() []error {
	var errorList []error

	validationErrs := errs.ValidateWorkflowName(w.Name)
	if len(validationErrs) != 0 {
		errorList = append(errorList, validationErrs...)
	}

	validationErrs = errs.ValidateWorkflowTime(w.Time)
	if len(validationErrs) != 0 {
		errorList = append(errorList, validationErrs...)
	}

	if len(w.Steps) == 0 {
		errorList = append(errorList, errs.ValidationError{Field: "workflow.steps", Err: errs.ErrEmpty})
	}

	for _, steps := range w.Steps {
		validationErrs := errs.ValidateWorkflowStepName(steps.Name)
		if len(validationErrs) != 0 {
			errorList = append(errorList, validationErrs...)
		}

		validationErrs = errs.ValidateWorkflowStepProgram(steps.Program)
		if len(validationErrs) != 0 {
			errorList = append(errorList, validationErrs...)
		}

		validationErrs = errs.ValidateWorkflowStepArgs(steps.Args)
		if len(validationErrs) != 0 {
			errorList = append(errorList, validationErrs...)
		}

		validationErrs = errs.ValidateWorkflowStepTimeout(strconv.Itoa(steps.Timeout))
		if len(validationErrs) != 0 {
			errorList = append(errorList, validationErrs...)
		}

		validationErrs = errs.ValidateWorkflowStepPause(strconv.Itoa(steps.Pause))
		if len(validationErrs) != 0 {
			errorList = append(errorList, validationErrs...)
		}
	}

	return errorList
}
