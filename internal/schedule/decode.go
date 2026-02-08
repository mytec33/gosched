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

func ReadScheduleFile(filename string) (Schedule, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Schedule{}, fmt.Errorf("open workflows file %q: %w", filename, err)
	}
	defer f.Close()

	s, err := DecodeSchedule(f)
	if err != nil {
		return Schedule{}, fmt.Errorf("decode workflows file %q: %w", filename, err)
	}
	return s, nil
}

func DecodeSchedule(r io.Reader) (Schedule, error) {
	s := Schedule{wf: make(map[MinuteKey][]Workflow)}
	var wf []Workflow

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	err := dec.Decode(&wf)
	if err != nil {
		return s, fmt.Errorf("%w: %w", ErrDecodeSchedule, err)
	}

	// Enforce exactly one top-level JSON value; allow only trailing whitespace.
	err = dec.Decode(&struct{}{})
	if err == nil {
		return s, fmt.Errorf("%w: trailing data", ErrDecodeSchedule)
	} else if !errors.Is(err, io.EOF) {
		return s, fmt.Errorf("%w: trailing data: %w", ErrDecodeSchedule, err)
	}

	schedule := s
	for _, wf := range wf {
		k, err := normalizeTime(wf.Time)
		if err != nil {
			return s, fmt.Errorf("error normalizing workflow %q: %w", wf.Name, err)
		}

		wf.Time = string(k)
		schedule.wf[k] = append(schedule.wf[k], wf)
	}

	return schedule, nil
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
