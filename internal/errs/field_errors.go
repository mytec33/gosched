// Package errs provides validation / JSON field level errors and constraints.
package errs

import (
	"errors"
	"fmt"
)

var (
	ErrFieldEmpty                  = errors.New("cannot be empty")
	ErrFieldTooLong                = errors.New("too long")
	ErrFieldWhitespaceOnly         = errors.New("cannot be all whitespace")
	ErrFieldWhitespacePadded       = errors.New("leading or trailing whitespace")
	ErrNumberNegative              = errors.New("number cannot be negative, must be zero (indefinite) or greater")
	ErrTimeFormatInvalid           = errors.New("invalid time format")
	ErrOnFailureInvalid            = errors.New("invalid workflow on failure mode")
	ErrOnFailureRequired           = errors.New("onFailure field required")
	ErrRetryRequired               = errors.New("retry config required when onFailure is set to retry")
	ErrRetryCountNegative          = errors.New("retry count must be 0 or greater")
	ErrRetryPauseNegative          = errors.New("retry pause seconds must be 0 or greater")
	ErrStepArgsCountExceeded       = fmt.Errorf("too many args provided: max is %d", MaxStepArgsCount)
	ErrStepArgsTotalLengthExceeded = fmt.Errorf("total length of all args exceeds limit: max is %d", MaxStepArgsTotalLength)
	ErrStepCountExceeded           = fmt.Errorf("too many steps in workflow: max is %d", MaxStepsCount)
	ErrStepsRequired               = fmt.Errorf("steps are required")
	ErrTriggerRequired             = fmt.Errorf("trigger field is required")
	ErrTriggerBeginAtRequired      = errors.New("trigger beginAt field is required")
	ErrTriggerEveryRequired        = errors.New("trigger every field is required")
	ErrWorkflowCountExceeded       = fmt.Errorf("too many workflows in schedule: max is %d", MaxWorkflowCount)
)

var (
	ErrDuplicateStepName     = errors.New("step name is a duplicate")
	ErrDuplicateWorkflowName = errors.New("duplicate workflow name")
)
