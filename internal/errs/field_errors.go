// Package errs provides validation / JSON field level errors and constraints.
package errs

import (
	"errors"
)

const (
	MaxWorkflowNameLength      int = 256
	MaxWorkflowTimeLength      int = 5     // hh:mm
	MaxWorkflowOnFailureLength int = 8     // max for continue
	MaxWorkflowRetryAttempts   int = 5     // how many times to retry on failure
	MaxWorkflowRetryPause      int = 300   // 5 minutes max retry
	MaxWorkflowStepPause       int = 3600  // 1 hour
	MaxWorkflowStepTimeout     int = 43200 // 12 hours
)

var (
	ErrEmpty                       = errors.New("cannot be empty")
	ErrExceedsMaxLimit             = errors.New("number exceeds maximum value")
	ErrNegativeNumber              = errors.New("number cannot be negative, must be zero (indefinite) or greater")
	ErrNotANumber                  = errors.New("value must be a number")
	ErrOnFailureInvalidMode        = errors.New("invalid workflow on failure mode")
	ErrOnFailureRequired           = errors.New("onFailure field required")
	ErrInvalidRetryValues          = errors.New("values must one of: abort, continue, or retry")
	ErrInvalidTime                 = errors.New("invalid time value")
	ErrTooLong                     = errors.New("too long")
	ErrWhitespaceAll               = errors.New("cannot be all whitespace")
	ErrWhitespaceLeadingOrTrailing = errors.New("leading or trailing whitespace")
)
