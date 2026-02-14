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

func ReadScheduleFile(filename string) (Schedule, []error, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Schedule{}, nil, fmt.Errorf("open workflows file %q: %w", filename, err)
	}
	defer f.Close()

	s, errors, err := DecodeSchedule(f)
	if err != nil {
		return Schedule{}, errors, fmt.Errorf("decode workflows file %q: %w", filename, err)
	}

	if len(errors) > 0 {
		return Schedule{}, errors, nil
	}
	return s, nil, nil
}

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
