package schedule

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"git.sr.ht/~mytec/gosched/internal/types"
)

var ErrDecodeSchedule = errors.New("decode schedule")

// ReadScheduleFiles opens the schedule file and delegates decoding and validation.
// Validation errors are returned in the slice. The returned error is reserved for
// I/O or decoding failures.
func ReadScheduleFiles(filename ScheduleSliceFlag) (Schedule, []error, error) {
	var schedule Schedule
	schedule.byMinute = make(map[types.MinuteOfDay][]Workflow)

	for _, file := range filename {
		f, err := os.Open(file)
		if err != nil {
			return Schedule{}, nil, fmt.Errorf("open workflows file %q: %w", file, err)
		}
		defer f.Close() // Keep in mind with many files this could be an issue but not yet

		s, validationErrors, err := DecodeSchedule(f)
		if err != nil {
			return Schedule{}, validationErrors, fmt.Errorf("decode workflows file %q: %w", file, err)
		}

		if len(validationErrors) > 0 {
			return Schedule{}, validationErrors, nil
		}

		for _, wf := range s.workflows {
			schedule.workflows = append(schedule.workflows, wf)
			schedule.byMinute[*wf.Trigger.BeginAt] = append(schedule.byMinute[*wf.Trigger.BeginAt], wf)
		}
	}
	return schedule, nil, nil
}

// DecodeSchedule reads JSON and performs validation.
// The returned slice contains validation errors found in the input.
// The returned error is reserved for I/O or decoding failures.
func DecodeSchedule(r io.Reader) (Schedule, []error, error) {
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

	// Loop once again to do normalization
	schedule := s
	for _, wf := range workflows {
		schedule.byMinute[*wf.Trigger.BeginAt] = append(schedule.byMinute[*wf.Trigger.BeginAt], wf)
	}
	schedule.workflows = workflows

	return schedule, nil, nil
}
