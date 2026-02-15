package errs

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~mytec/gosched/internal/errs"
	"git.sr.ht/~mytec/gosched/internal/platform"
)

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

type rule func(string) error
type ruleSet func(string) []error

func createRuleset(field string, r ...rule) ruleSet {
	return func(value string) []error {
		var out []error
		for _, rule := range r {
			err := rule(value)
			if err != nil {
				out = append(out, ValidationError{Field: field, Err: err})
			}
		}
		return out
	}
}

func LimitLength(max int) rule {
	return func(s string) error {
		if len(s) > max {
			return ErrTooLong
		}

		return nil
	}
}

func RequireNoLeadingTrailingWhitespace(s string) error {
	if strings.TrimSpace(s) != s {
		return ErrWhitespaceLeadingOrTrailing
	}

	return nil
}

func RequireNonEmpty(s string) error {
	if s == "" {
		return ErrEmpty
	}

	return nil
}

func RequireNonNegative(x int) rule {
	return func(s string) error {
		num, err := strconv.Atoi(s)
		if err != nil {
			return ErrNotANumber
		}

		if num < 0 {
			return ErrNonNegativeNumber
		}

		return nil
	}
}

func NonNegative(x int) rule {
	return func(s string) error {
		if s == "" {
			return nil
		}

		num, err := strconv.Atoi(s)
		if err != nil {
			return ErrNotANumber
		}

		if num < 0 {
			return ErrNonNegativeNumber
		}

		return nil
	}
}

func RequireNoWhitespace(s string) error {
	if strings.TrimSpace(s) == "" {
		return ErrWhitespaceAll
	}

	return nil
}

func RequireValidTime(timeVal string) error {
	layout := "15:04"

	_, err := time.Parse(layout, timeVal)
	if err != nil {
		return ErrInvalidTime
	}

	return nil
}

func ValidateProgramPath(s string) error {
	if s == "" {
		return errs.ErrEmpty
	}

	if strings.TrimSpace(s) == "" {
		return errs.ErrWhitespaceAll
	}

	if strings.TrimSpace(s) != s {
		return errs.ErrWhitespaceLeadingOrTrailing
	}

	if len(s) > errs.MaxPathLength() {
		return errs.ErrTooLong
	}

	return nil
}

var ValidateWorkflowName = createRuleset(
	"workflow.name",
	RequireNonEmpty,
	RequireNoWhitespace,
	RequireNoLeadingTrailingWhitespace,
	LimitLength(errs.MaxWorkflowNameLength),
)

var ValidateWorkflowTime = createRuleset(
	"workflow.time",
	RequireNonEmpty,
	RequireNoWhitespace,
	RequireNoLeadingTrailingWhitespace,
	RequireValidTime,
	LimitLength(errs.MaxWorkflowTimeLength),
)

var ValidateWorkflowStepArgs = createRuleset(
	"workflow.step.arg",
	RequireNoWhitespace,
	RequireNoLeadingTrailingWhitespace,
	LimitLength(errs.MaxWorkflowNameLength),
)

var ValidateWorkflowStepName = createRuleset(
	"workflow.step.name",
	RequireNonEmpty,
	RequireNoWhitespace,
	RequireNoLeadingTrailingWhitespace,
	LimitLength(errs.MaxWorkflowNameLength),
)

var ValidateWorkflowStepPause = createRuleset(
	"workflow.step.pause",
	NonNegative(1),
)

var ValidateWorkflowStepProgram = createRuleset(
	"workflow.step.program",
	RequireNonEmpty,
	RequireNoWhitespace,
	RequireNoLeadingTrailingWhitespace,
	LimitLength(platform.MaxPathLength()),
)

var ValidateWorkflowStepTimeout = createRuleset(
	"workflow.step.timeout",
	NonNegative(1),
)
