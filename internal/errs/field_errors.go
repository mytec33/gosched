// Package errs provides validation / JSON field level errors and constraints.
package errs

import (
	"errors"
)

const (
	MaxWorkflowNameLength int = 256
	MaxWorkflowTimeLength int = 5 // hh:mm
)

var (
	ErrEmpty                       = errors.New("cannot be empty")
	ErrNegativeNumber              = errors.New("number cannot be negative, must be zero (indefinite) or greater")
	ErrOnFailureInvalidMode        = errors.New("invalid workflow on failure mode")
	ErrOnFailureRequired           = errors.New("onFailure field required")
	ErrInvalidTime                 = errors.New("invalid time value")
	ErrTooLong                     = errors.New("too long")
	ErrWhitespaceAll               = errors.New("cannot be all whitespace")
	ErrWhitespaceLeadingOrTrailing = errors.New("leading or trailing whitespace")
)
