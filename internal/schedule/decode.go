package schedule

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"git.sr.ht/~mytec/gosched/internal/errs"
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
			schedule.byMinute[wf.Time] = append(schedule.byMinute[wf.Time], wf)
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
	valErrs = append(valErrs, validateWorkflowCount(workflows)...)

	if len(valErrs) > 0 {
		return Schedule{}, valErrs, nil
	}

	// Loop once again to do normalization
	schedule := s
	for _, wf := range workflows {
		schedule.byMinute[wf.Time] = append(schedule.byMinute[wf.Time], wf)
	}
	schedule.workflows = workflows

	return schedule, nil, nil
}

func validateUniqueWorkflowNames(wfs []Workflow) []error {
	wfNames := make(map[string]string)
	var errorList []error

	for _, wf := range wfs {
		key := strings.ToLower(strings.TrimSpace(wf.Name))

		v, exists := wfNames[key]
		if exists {
			errorList = append(errorList, fmt.Errorf("%w: '%v' duplicates '%v'", errs.ErrDuplicateWorkflowName, wf.Name, v))
		} else {
			wfNames[key] = wf.Name
		}
	}

	return errorList
}

func validateWorkflowCount(wfs []Workflow) []error {
	if len(wfs) > errs.MaxWorkflowCount {
		return []error{errs.ErrWorkflowCount}
	}

	return nil
}
