// Package schedule provides schedule layout and locking
package schedule

import (
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

	vErrs := errs.ValidateWorkflowName(w.Name)
	if len(vErrs) != 0 {
		errorList = append(errorList, vErrs...)
	}

	vErrs = errs.ValidateWorkflowTime(w.Time)
	if len(vErrs) != 0 {
		errorList = append(errorList, vErrs...)
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

		vErrs = errs.ValidateWorkflowStepTimeout(steps.Timeout)
		if len(vErrs) != 0 {
			errorList = append(errorList, vErrs...)
		}

		vErrs = errs.ValidateWorkflowStepPause(steps.Pause)
		if len(vErrs) != 0 {
			errorList = append(errorList, vErrs...)
		}
	}

	return errorList
}
