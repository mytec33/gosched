package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const MinutesInDay int = 24 * 60

var (
	ErrTimeFormatInvalid = errors.New("invalid time format")
)

type MinuteOfDay int

func MinuteOfDayFromTime(t time.Time) MinuteOfDay {
	return MinuteOfDay(t.Hour()*60 + t.Minute())
}

func (m *MinuteOfDay) MinutesSince(previous MinuteOfDay) int {
	currentMinute := int(*m)
	previousMinute := int(previous)

	if currentMinute < 0 || currentMinute >= MinutesInDay {
		panic(fmt.Sprintf(
			"workflow.MinuteOfDay error: invalid current minute: %d",
			currentMinute,
		))
	}

	if previousMinute < 0 || previousMinute >= MinutesInDay {
		panic(fmt.Sprintf(
			"workflow.MinuteOfDay error: invalid previous minute: %d",
			previousMinute,
		))
	}

	return (int(*m) - int(previous) + MinutesInDay) % MinutesInDay
}

func ParseMinuteOfDay(s string) (MinuteOfDay, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrTimeFormatInvalid, err)
	}

	return MinuteOfDay(t.Hour()*60 + t.Minute()), nil
}

func (m *MinuteOfDay) String() string {
	minute := int(*m)
	if minute < 0 || minute >= MinutesInDay {
		panic(fmt.Sprintf("workflow.MinuteOfDay error: invalid minute: %d", minute))
	}

	h := int(*m) / 60
	min := int(*m) % 60

	return fmt.Sprintf("%02d:%02d", h, min)
}

func (m *MinuteOfDay) UnmarshalJSON(b []byte) error {
	var s string

	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	v, err := ParseMinuteOfDay(s)
	if err != nil {
		return err
	}

	*m = v
	return nil
}
