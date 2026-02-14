package schedule

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
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
	s := Schedule{wf: make(map[MinuteKey][]Workflow)}
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

	if len(valErrs) > 0 {
		return Schedule{}, valErrs, nil
	}

	// Loop once again to do normalization
	schedule := s
	for _, wf := range workflows {
		k, err := normalizeTime(wf.Time)
		if err != nil {
			return s, nil, fmt.Errorf("error normalizing workflow %q: %w", wf.Name, err)
		}

		wf.Time = string(k)
		schedule.wf[k] = append(schedule.wf[k], wf)
	}

	return schedule, nil, nil
}

func normalizeTime(timeKey string) (MinuteKey, error) {
	var h, m int
	if _, err := fmt.Sscanf(strings.TrimSpace(timeKey), "%d:%d", &h, &m); err != nil {
		return "", fmt.Errorf("invalid time %q", timeKey)
	}

	nt := fmt.Sprintf("%02d:%02d", h, m)
	_, err := time.Parse("15:04", nt)
	if err != nil {
		return "", fmt.Errorf("invalid normalized time %q: %w", nt, err)
	}
	return MinuteKey(nt), nil
}
