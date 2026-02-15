// Package errs provides validation / JSON field level errors and constraints.
package errs

import (
	"errors"
	"fmt"
	"time"
)

const (
	MaxWorkflowNameLength  int = 256
	MaxWorkflowTimeLength  int = 5     // hh:mm
	maxWorkflowStepPause   int = 3600  // 1 hour
	maxWorkflowStepTimeout int = 43200 // 12 hours
)

var (
	ErrEmpty                       = errors.New("cannot be empty")
	ErrNonNegativeNumber           = errors.New("number cannot be negative")
	ErrNotANumber                  = errors.New("value must be a number")
	ErrStepPauseInvalid            = errors.New("invalid pause duration, must be zero (no pause) or greater")
	ErrStepPauseTooLong            = fmt.Errorf("invalid pause duration, must be less than equal %d or %s", maxWorkflowStepPause, time.Duration(maxWorkflowStepPause)*time.Second)
	ErrStepTimeoutInvalid          = errors.New("invalid timeout duration, must be zero (no timeout) or greater")
	ErrStepTimeoutTooLong          = fmt.Errorf("invalid timeout duration, must be less than equal %d or %s", maxWorkflowStepTimeout, time.Duration(maxWorkflowStepTimeout)*time.Second)
	ErrInvalidTime                 = errors.New("invalid time value")
	ErrTooLong                     = errors.New("too long")
	ErrWhitespaceAll               = errors.New("cannot be all whitespace")
	ErrWhitespaceLeadingOrTrailing = errors.New("leading or trailing whitespace")
)
