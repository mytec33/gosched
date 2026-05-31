// Package errs provides validation / JSON field level errors and constraints.
package errs

import (
	"errors"
	"fmt"
)

var (
	ErrEmpty                       = errors.New("cannot be empty")
	ErrNegativeNumber              = errors.New("number cannot be negative, must be zero (indefinite) or greater")
	ErrOnFailureInvalidMode        = errors.New("invalid workflow on failure mode")
	ErrOnFailureRequired           = errors.New("onFailure field required")
	ErrInvalidTimeFormat           = errors.New("invalid time format")
	ErrRetryRequired               = errors.New("retry config required when onFailure is set to retry")
	ErrStepArgsTooMany             = fmt.Errorf("too many args provided: max is %d", MaxStepArgsCount)
	ErrStepArgsTotalLength         = fmt.Errorf("total length of all args exceeds limit: max is %d", MaxStepArgsTotalLength)
	ErrStepsCount                  = fmt.Errorf("too many steps in workflow: max is %d", MaxStepsCount)
	ErrStepsMissing                = fmt.Errorf("steps are required")
	ErrTimeFieldNotPresent         = fmt.Errorf("time field is required")
	ErrTooLong                     = errors.New("too long")
	ErrRetryCountNegative          = errors.New("retry count must be 0 or greater")
	ErrRetryPauseNegative          = errors.New("retry pause seconds must be 0 or greater")
	ErrWhitespaceAll               = errors.New("cannot be all whitespace")
	ErrWhitespaceLeadingOrTrailing = errors.New("leading or trailing whitespace")
	ErrWorkflowCount               = fmt.Errorf("too many workflows in schedule: max is %d", MaxWorkflowCount)
)

var (
	ErrDuplicateStepName     = errors.New("step name is a duplicate")
	ErrDuplicateWorkflowName = errors.New("duplicate workflow name")
)
