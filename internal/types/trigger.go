package types

import (
	"encoding/json"
	"fmt"
)

type Trigger struct {
	Every   *Cadence     `json:"every"`
	BeginAt *MinuteOfDay `json:"beginAt"`
}

func (t *Trigger) String() string {
	return fmt.Sprintf("%s %s", t.Every.String(), t.BeginAt.String())
}

func (t *Trigger) UnmarshalJSON(b []byte) error {
	var raw struct {
		Every   *Cadence     `json:"every"`
		BeginAt *MinuteOfDay `json:"beginAt"`
	}

	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	t.Every = raw.Every
	t.BeginAt = raw.BeginAt

	return nil
}
