// Package schedule provides schedule layout and locking
package schedule

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"git.sr.ht/~mytec/gosched/internal/platform"
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

type ValidationError struct {
	Field string
	Err   error
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %v", e.Field, e.Err)
}

func (e ValidationError) Unwrap() error {
	return e.Err
}

const (
	maxWorkflowNameLength int = 256
)

var (
	ErrEmpty                       = errors.New("cannot be empty")
	ErrInvalidDuration             = errors.New("invalid number, must be zero or greater")
	ErrInvalidTime                 = errors.New("invalid time value")
	ErrTooLong                     = errors.New("too long")
	ErrWhitespaceAll               = errors.New("cannot be all whitespace")
	ErrWhitespaceLeadingOrTrailing = errors.New("leading or trailing whitespace")
)

func (w Workflow) Validate() []error {
	var errorList []error

	err := validateRequired("workflow.name", w.Name, validateIdentifier)
	if err != nil {
		errorList = append(errorList, err)
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
		stepPrefix := fmt.Sprintf("workflow.steps[%d]", i)

		err = validateRequired(stepPrefix+".name", steps.Name, validateIdentifier)
		if err != nil {
			errorList = append(errorList, err)
		}

		err = validateRequired(stepPrefix+".program", steps.Program, platform.ValidateProgramPath)
		if err != nil {
			errorList = append(errorList, err)
		}

		if steps.Timeout < 0 {
			errorList = append(errorList, invalid(stepPrefix+".timeout", ErrInvalidDuration))
		}

		if steps.Pause < 0 {
			errorList = append(errorList, invalid(stepPrefix+".pause", ErrInvalidDuration))
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
	return err == nil
}

func validateIdentifier(s string) error {
	if s == "" {
		return ErrEmpty
	}

	if strings.TrimSpace(s) == "" {
		return ErrWhitespaceAll
	}

	if strings.TrimSpace(s) != s {
		return ErrWhitespaceLeadingOrTrailing
	}

	if len(s) > maxWorkflowNameLength {
		return ErrTooLong
	}
	return nil
}

func validateRequired(field, value string, check func(string) error) error {
	if value == "" {
		return invalid(field, ErrEmpty)
	}

	if err := check(value); err != nil {
		return invalid(field, err)
	}
	return nil
}
