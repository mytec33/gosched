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
	Retry     RetryConfig         `json:"retry"`
	Steps     []Step              `json:"steps"`
}

type RetryConfig struct {
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

	if w.Retry.NumberRetries < 0 {
		errorList = append(errorList, errs.ErrNegativeNumber)
	}

	if w.Retry.PauseSeconds < 0 {
		errorList = append(errorList, errs.ErrNegativeNumber)
	}

	for _, steps := range w.Steps {
		vErrs := errs.ValidateWorkflowStepName(steps.Name)
		if len(vErrs) != 0 {
			errorList = append(errorList, vErrs...)
		}

		if steps.Timeout < 0 {
			errorList = append(errorList, errs.ErrNegativeNumber)
		}

		if steps.Pause < 0 {
			errorList = append(errorList, errs.ErrNegativeNumber)
		}

		trimmedProgram := strings.TrimSpace(steps.Program)

		if steps.Program == "" {
			errorList = append(errorList, errs.ErrEmpty)
		} else if trimmedProgram == "" {
			errorList = append(errorList, errs.ErrWhitespaceAll)
		} else if trimmedProgram != steps.Program {
			errorList = append(errorList, errs.ErrWhitespaceLeadingOrTrailing)
		}

		for _, arg := range steps.Args {
			trimmed := strings.TrimSpace(arg)

			switch {
			case arg == "":
				errorList = append(errorList, errs.ErrEmpty)
			case trimmed == "":
				errorList = append(errorList, errs.ErrWhitespaceAll)
			case trimmed != arg:
				errorList = append(errorList, errs.ErrWhitespaceLeadingOrTrailing)
			case len(arg) > errs.MaxProgramArgsLengths:
				errorList = append(errorList, errs.ErrArgsTooLong)
			}
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

func WorkflowAbortsOnFailure(wf Workflow) bool {
	return wf.OnFailure != nil && *wf.OnFailure == policy.Abort
}

func WorkflowContinuesOnFailure(wf Workflow) bool {
	return wf.OnFailure != nil && *wf.OnFailure == policy.Continue
}

func WorkflowRetriesOnFailure(wf Workflow) bool {
	return wf.OnFailure != nil && *wf.OnFailure == policy.Retry
}
