package workflow

import (
	"fmt"
	"strings"
	"time"
)

const (
	MaxStepArgLength       int           = 256
	MaxStepArgsCount       int           = 64
	MaxStepArgsTotalLength int           = 4096
	MaxStepPauseDuration   time.Duration = time.Hour * 1
	MaxStepProgramLength   int           = 256
	MaxStepNameLength      int           = 256
)

var (
	ErrStepArgsCountExceeded       = fmt.Errorf("too many args provided: max is %d", MaxStepArgsCount)
	ErrStepArgsTotalLengthExceeded = fmt.Errorf("total length of all args exceeds limit: max is %d", MaxStepArgsTotalLength)
	ErrStepFieldEmpty              = fmt.Errorf("field cannot be empty")
	ErrStepFieldTooLong            = fmt.Errorf("field longer than allowed limit")
	ErrStepFieldWhitespacePadded   = fmt.Errorf("field cannot contain leading or trailing whitespace")
	ErrStepFieldWhitespaceOnly     = fmt.Errorf("field cannot contain whitespace only")
	ErrStepPauseDurationTooLarge   = fmt.Errorf("pause duration must be 1 hour or less")
)

var defaultTimeout = ConfigDuration{duration: 30 * time.Second}

type StepRaw struct {
	Name    string          `json:"name"`
	Program string          `json:"program"`
	Args    []string        `json:"args"`
	Timeout *ConfigDuration `json:"timeout"`
	Pause   ConfigDuration  `json:"pause"`
}

func (s StepRaw) Validate() (Step, []error) {
	var errorList []error
	var step Step

	errorList = append(errorList, validateStepStrings("step", s.Name,
		MaxStepNameLength)...)
	step.Name = s.Name

	errorList = append(errorList, validateStepStrings("program", s.Program,
		MaxStepProgramLength)...)
	step.Program = s.Program

	errorList = append(errorList, validateStepArgs("args", s.Args)...)
	step.Args = append([]string(nil), s.Args...)

	// This is where validation isn't quite validation. We cannot allow a nil
	// value to move forward, so in the absence of one, a reasonable value is
	// provided for the user. If it cuts them short, they can adjust and provide
	// a more suitable value.
	if s.Timeout == nil {
		step.Timeout = defaultTimeout
	} else {
		step.Timeout = *s.Timeout
	}

	if s.Pause.Duration() > MaxStepPauseDuration {
		errorList = append(errorList, fmt.Errorf("%s: %w", "pause", ErrStepPauseDurationTooLarge))
	}
	step.Pause.duration = s.Pause.duration

	if len(errorList) > 0 {
		return Step{}, errorList
	}

	return step, nil
}

func validateStepArgs(field string, args []string) []error {
	var errorList []error

	if len(args) > MaxStepArgsCount {
		errorList = append(errorList, ErrStepArgsCountExceeded)
	}

	totalArgsLength := 0
	for j, arg := range args {
		totalArgsLength += len(arg)

		argField := fmt.Sprintf("%s[%d]", field, j+1)
		errorList = append(errorList, validateStepStrings(argField, arg, MaxStepArgLength)...)
	}

	if totalArgsLength > MaxStepArgsTotalLength {
		errorList = append(errorList, ErrStepArgsTotalLengthExceeded)
	}

	return errorList
}

func validateStepStrings(field string, s string, maxLength int) []error {
	var errorList []error
	trimmed := strings.TrimSpace(s)

	if s == "" {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, ErrStepFieldEmpty))
	} else if trimmed == "" {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, ErrStepFieldWhitespaceOnly))
	} else if trimmed != s {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, ErrStepFieldWhitespacePadded))
	} else if len(trimmed) > maxLength {
		errorList = append(errorList, fmt.Errorf("%s: %w", field, ErrStepFieldTooLong))
	}

	return errorList
}
