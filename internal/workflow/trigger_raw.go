package workflow

import (
	"encoding/json"
	"fmt"
)

type TriggerRaw struct {
	Every   *Cadence     `json:"every"`
	BeginAt *MinuteOfDay `json:"beginAt"`
}

func (t *TriggerRaw) String() string {
	return fmt.Sprintf("%s %s", t.Every.String(), t.BeginAt.String())
}

func (t *TriggerRaw) UnmarshalJSON(b []byte) error {
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
