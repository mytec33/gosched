package errs

import (
	"fmt"
	"strings"
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
	if max < 0 {
		panic("LimitLength: max must be >= 0")
	}
	return func(s string) error {
		if len(s) > max {
			return ErrTooLong
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

func RequireNoWhitespace(s string) error {
	if s == "" {
		return nil
	}

	if strings.TrimSpace(s) == "" {
		return ErrWhitespaceAll
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

var ValidateWorkflowStepName = createRuleset(
	"workflow.step.name",
	RequireNonEmpty,
	RequireNoWhitespace,
	RequireNoLeadingTrailingWhitespace,
	LimitLength(MaxWorkflowNameLength),
)
