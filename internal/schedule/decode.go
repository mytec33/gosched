package schedule

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func ReadScheduleFile(filename string) (Schedule, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open workflows file %q: %w", filename, err)
	}
	defer f.Close()

	s, err := DecodeSchedule(f)
	if err != nil {
		return nil, fmt.Errorf("decode workflows file %q: %w", filename, err)
	}
	return s, nil
}

func DecodeSchedule(r io.Reader) (Schedule, error) {
	var wf []Workflow

	dec := json.NewDecoder(r)
	err := dec.Decode(&wf)
	if err != nil {
		return nil, fmt.Errorf("error parsing scheduler configuration: %w", err)
	}

	schedule := make(Schedule)
	for _, wf := range wf {
		k, err := normalizeTime(wf.Time)
		if err != nil {
			return nil, fmt.Errorf("error decoding workflow %q: %w", wf.Name, err)
		}

		wf.Time = string(k)
		schedule[k] = append(schedule[k], wf)
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
