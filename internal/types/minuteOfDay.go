// Package types provides types specific to this program
package types

import (
	"encoding/json"
	"fmt"
	"time"

	"git.sr.ht/~mytec/gosched/internal/errs"
)

const MinutesInDay int = 24 * 60

type MinuteOfDay int

func (m *MinuteOfDay) MinutesSince(previous MinuteOfDay) int {
	return (int(*m) - int(previous) + MinutesInDay) % MinutesInDay
}

func ParseMinuteOfDay(s string) (MinuteOfDay, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", errs.ErrTimeFormatInvalid, err)
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
