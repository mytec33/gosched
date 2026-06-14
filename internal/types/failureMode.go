package types

import (
	"encoding/json"
	"fmt"

	"git.sr.ht/~mytec/gosched/internal/errs"
)

type FailureMode struct {
	v string
}

var (
	Abort    = FailureMode{"abort"}
	Continue = FailureMode{"continue"}
	Retry    = FailureMode{"retry"}
)

func (f FailureMode) String() string {
	return f.v
}

func (f FailureMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.v)
}

func (f *FailureMode) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "abort":
		*f = Abort
	case "continue":
		*f = Continue
	case "retry":
		*f = Retry
	default:
		return fmt.Errorf("%w: %q", errs.ErrOnFailureInvalid, s)
	}

	return nil
}
