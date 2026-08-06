package workflow

import (
	"fmt"
)

type Trigger struct {
	Every   Cadence     `json:"every"`
	BeginAt MinuteOfDay `json:"beginAt"`
}

func (t *Trigger) String() string {
	return fmt.Sprintf("%s %s", t.Every.String(), t.BeginAt.String())
}
