package errs

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

func LimitValue(max int) rule {
	return func(s string) error {
		num, err := strconv.Atoi(s)
		if err != nil {
			return ErrNotANumber
		}

		if num > max {
			return ErrExceedsMaxLimit
		}

		return nil
	}
}

func RequireNoLeadingTrailingWhitespace(s string) error {
	if s == "" {
		return nil
	}

	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil // let ErrWhitespaceAll handle all-whitespace
	}

	if trimmed != s {
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
	if s == "" {
		return nil
	}

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

var ValidateWorkflowName = createRuleset(
	"workflow.name",
	RequireNonEmpty,
	RequireNoWhitespace,
	RequireNoLeadingTrailingWhitespace,
	LimitLength(MaxWorkflowNameLength),
)

var ValidateWorkflowTime = createRuleset(
	"workflow.time",
	RequireNonEmpty,
	RequireNoWhitespace,
	RequireNoLeadingTrailingWhitespace,
	RequireValidTime,
	LimitLength(MaxWorkflowTimeLength),
)

var ValidateWorkflowStepArgs = createRuleset(
	"workflow.step.arg",
	RequireNoWhitespace,
	RequireNoLeadingTrailingWhitespace,
	LimitLength(MaxWorkflowNameLength),
)

var ValidateWorkflowStepName = createRuleset(
	"workflow.step.name",
	RequireNonEmpty,
	RequireNoWhitespace,
	RequireNoLeadingTrailingWhitespace,
	LimitLength(MaxWorkflowNameLength),
)

var ValidateWorkflowStepPause = createRuleset(
	"workflow.step.pause",
	LimitValue(MaxWorkflowStepPause),
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
	LimitValue(MaxWorkflowStepTimeout),
	NonNegative(1),
)
