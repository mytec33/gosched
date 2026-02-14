// Package schedule provides schedule layout and locking
package schedule

import (
	"errors"
	"fmt"
	"time"
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
	Timeout int    `json:"timeOut"`
}

type ValidationError struct {
	Field string
	Err   error
}

const (
	maxWorkflowLength int = 256
)

var (
	ErrEmpty           = errors.New("cannot be empty")
	ErrInvalidDuration = errors.New("invalid number, must be zero or greater")
	ErrInvalidTime     = errors.New("invalid time value")
	ErrTooLong         = errors.New("too long")
)

func (w Workflow) Validate() []error {
	var errorList []error

	if w.Name == "" {
		errorList = append(errorList, invalid("workflow.name", ErrEmpty))
	} else if len(w.Name) > maxWorkflowLength {
		errorList = append(errorList, invalid("workflow.name", ErrTooLong))
	}

	if w.Time == "" {
		errorList = append(errorList, invalid("workflow.time", ErrEmpty))
	} else if !parseTime(w.Time) {
		errorList = append(errorList, invalid("workflow.time", ErrInvalidTime))
	}

	if len(w.Steps) == 0 {
		errorList = append(errorList, invalid("workflow.steps", ErrEmpty))
	}

	for i, steps := range w.Steps {
		if steps.Name == "" {
			errorList = append(errorList, invalid(fmt.Sprintf("workflow.steps[%d].name", i), ErrEmpty))
		} else if len(steps.Name) > maxWorkflowLength {
			errorList = append(errorList, invalid(fmt.Sprintf("workflow.steps[%d].name", i), ErrTooLong))
		}

		if steps.Timeout < 0 {
			errorList = append(errorList, invalid(fmt.Sprintf("workflow.steps[%d].timeout", i), ErrInvalidDuration))
		}
	}

	return errorList
}

func invalid(field string, reason error) error {
	return ValidationError{Field: field, Err: reason}
}

func parseTime(timeVal string) bool {
	layout := "15:04"

	_, err := time.Parse(layout, timeVal)
	if err != nil {
		return false
	}

	return true
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %v", e.Field, e.Err)
}
