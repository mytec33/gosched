// Package errs provides validation / JSON field level errors and constraints.
package errs

import (
	"errors"
)

const (
	MaxWorkflowNameLength  int = 256
	MaxWorkflowTimeLength  int = 5     // hh:mm
	MaxWorkflowStepPause   int = 3600  // 1 hour
	MaxWorkflowStepTimeout int = 43200 // 12 hours
)

var (
	ErrEmpty                       = errors.New("cannot be empty")
	ErrExceedsMaxLimit             = errors.New("number exceeds maximum value")
	ErrNonNegativeNumber           = errors.New("number cannot be negative, must be zero (indefinite) or greater")
	ErrNotANumber                  = errors.New("value must be a number")
	ErrInvalidTime                 = errors.New("invalid time value")
	ErrTooLong                     = errors.New("too long")
	ErrWhitespaceAll               = errors.New("cannot be all whitespace")
	ErrWhitespaceLeadingOrTrailing = errors.New("leading or trailing whitespace")
)
