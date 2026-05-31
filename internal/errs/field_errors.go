// Package errs provides validation / JSON field level errors and constraints.
package errs

import (
	"errors"
)

var (
	ErrEmpty                       = errors.New("cannot be empty")
	ErrNegativeNumber              = errors.New("number cannot be negative, must be zero (indefinite) or greater")
	ErrOnFailureInvalidMode        = errors.New("invalid workflow on failure mode")
	ErrOnFailureRequired           = errors.New("onFailure field required")
	ErrInvalidTimeFormat           = errors.New("invalid time format")
	ErrArgsTooLong                 = errors.New("args too long")
	ErrTooLong                     = errors.New("too long")
	ErrRetryCountNegative          = errors.New("retry count must be 0 or greater")
	ErrRetryPauseNegative          = errors.New("retry pause seconds must be 0 or greater")
	ErrWhitespaceAll               = errors.New("cannot be all whitespace")
	ErrWhitespaceLeadingOrTrailing = errors.New("leading or trailing whitespace")
)

var (
	ErrDuplicateStepName     = errors.New("step name is a duplicate")
	ErrDuplicateWorkflowName = errors.New("duplicate workflow name")
)
