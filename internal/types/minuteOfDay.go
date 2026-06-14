// Package types provides types specific to this program
package types

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

func minuteOfDayFromTime(t time.Time) int {
	return t.Hour()*60 + t.Minute()
}

func MinuteOfDayFromTime(t time.Time) MinuteOfDay {
	return MinuteOfDay(minuteOfDayFromTime(t))
}

func (m *MinuteOfDay) MinutesSince(previous MinuteOfDay) int {
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
	h := int(*m) / 60
	min := int(*m) % 60

	return fmt.Sprintf("%02d:%02d", h, min)
}

func (m *MinuteOfDay) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	v, err := ParseMinuteOfDay(s)
	if err != nil {
		return err
	}

	*m = v
	return nil
}
