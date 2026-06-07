package schedule

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"git.sr.ht/~mytec/gosched/internal/types"
)

type FileValidationError struct {
	File string
	Err  error
}

func (e FileValidationError) Error() string {
	return fmt.Sprintf("%s: %v", e.File, e.Err)
}

func (e FileValidationError) Unwrap() error {
	return e.Err
}

var (
	ErrDecodeSchedule = errors.New("decode schedule")
	ErrFileIOError    = errors.New("error opening file")
)

// ReadScheduleFiles opens the schedule file and delegates decoding and validation.
// Validation errors are returned in the slice. The returned error is reserved for
// I/O or decoding failures.
func ReadScheduleFiles(filename ScheduleSliceFlag) (Schedule, []error, error) {
	var allValidationErrors []error
	var schedule Schedule
	schedule.byMinute = make(map[types.MinuteOfDay][]Workflow)

	for _, file := range filename {
		s, validationErrors, err := decodeScheduleFile(file)
		if err != nil {
			return Schedule{}, validationErrors, err
		}

		if len(validationErrors) > 0 {
			for _, vErr := range validationErrors {
				allValidationErrors = append(allValidationErrors,
					FileValidationError{
						File: file,
						Err:  vErr,
					})
			}
			continue
		}

		schedule.workflows = append(schedule.workflows, s.workflows...)
	}
	if len(allValidationErrors) > 0 {
		return Schedule{}, allValidationErrors, nil
	}

	return schedule, nil, nil
}

func decodeScheduleFile(file string) (Schedule, []error, error) {
	f, err := os.Open(file)
	if err != nil {
		return Schedule{}, nil, fmt.Errorf("%w: %q: %w", ErrFileIOError, file, err)
	}
	defer f.Close()

	s, validationErrors, err := DecodeWorkflows(f)
	if err != nil {
		return Schedule{}, validationErrors, fmt.Errorf("decode workflows file %q: %w", file, err)
	}

	return s, validationErrors, nil
}

// DecodeWorkflows reads JSON and performs validation.
// The returned slice contains validation errors found in the input.
// The returned error is reserved for I/O or decoding failures.
func DecodeWorkflows(r io.Reader) (Schedule, []error, error) {
	s := Schedule{byMinute: make(map[types.MinuteOfDay][]Workflow)}
	var workflows []Workflow
	var workflowsRaw []WorkflowRaw

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	err := dec.Decode(&workflowsRaw)
	if err != nil {
		return s, nil, fmt.Errorf("%w: %w", ErrDecodeSchedule, err)
	}

	// Enforce exactly one top-level JSON value; allow only trailing whitespace.
	err = dec.Decode(&struct{}{})
	if err == nil {
		return s, nil, fmt.Errorf("%w: trailing data", ErrDecodeSchedule)
	} else if !errors.Is(err, io.EOF) {
		return s, nil, fmt.Errorf("%w: trailing data: %w", ErrDecodeSchedule, err)
	}

	// Loop through workflows to validate and bail if anything found
	var valErrs []error
	for _, wfRaw := range workflowsRaw {
		wf, valErrors := wfRaw.Validate()
		if len(valErrors) > 0 {
			valErrs = append(valErrs, valErrors...)
			continue
		}

		// Keep workflows as the trusted set; invalid raw workflows never cross
		// the decode boundary.
		workflows = append(workflows, wf)
	}

	if len(valErrs) > 0 {
		return Schedule{}, valErrs, nil
	}

	s.workflows = workflows

	return s, nil, nil
}
