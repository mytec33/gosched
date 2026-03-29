package schedule

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"git.sr.ht/~mytec/gosched/internal/errs"
	"git.sr.ht/~mytec/gosched/internal/types"
)

var ErrDecodeSchedule = errors.New("decode schedule")

// ReadScheduleFile opens the schedule file and delegates decoding and validation.
// Validation errors are returned in the slice. The returned error is reserved for
// I/O or decoding failures.
func ReadScheduleFile(filename string) (Schedule, []error, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Schedule{}, nil, fmt.Errorf("open workflows file %q: %w", filename, err)
	}
	defer f.Close()

	s, validationErrors, err := DecodeSchedule(f)
	if err != nil {
		return Schedule{}, validationErrors, fmt.Errorf("decode workflows file %q: %w", filename, err)
	}

	if len(validationErrors) > 0 {
		return Schedule{}, validationErrors, nil
	}
	return s, nil, nil
}

// DecodeSchedule reads JSON and performs validation.
// The returned slice contains validation errors found in the input.
// The returned error is reserved for I/O or decoding failures.
func DecodeSchedule(r io.Reader) (Schedule, []error, error) {
	s := Schedule{wf: make(map[types.MinuteOfDay][]Workflow)}
	var workflows []Workflow

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	err := dec.Decode(&workflows)
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
	for _, wf := range workflows {
		valErrors := wf.Validate()
		if len(valErrors) > 0 {
			valErrs = append(valErrs, valErrors...)
		}
	}

	valErrs = append(valErrs, validateUniqueWorkflowNames(workflows)...)

	if len(valErrs) > 0 {
		return Schedule{}, valErrs, nil
	}

	// Loop once again to do normalization
	schedule := s
	for _, wf := range workflows {
		schedule.wf[wf.Time] = append(schedule.wf[wf.Time], wf)
	}

	return schedule, nil, nil
}

func validateUniqueWorkflowNames(wfs []Workflow) []error {
	wfNames := make(map[string]struct{})
	var errors []error

	for _, wf := range wfs {
		_, exists := wfNames[wf.Name]
		if exists {
			errors = append(errors, fmt.Errorf("%w: %s", errs.ErrDuplicateWorkflowName, wf.Name))
		} else {
			wfNames[wf.Name] = struct{}{}
		}
	}

	return errors
}
