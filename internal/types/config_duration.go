package types

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrConfigDurationInvalid     = errors.New("invalid duration")
	ErrConfigDurationSignInvalid = errors.New("+ and - symbols not allowed")
)

// ConfigDuration is the trusted config representation of a runtime duration.
// It wraps time.Duration while keeping config parsing rules separate from both
// time.ParseDuration and Cadence recurrence rules.
type ConfigDuration struct {
	duration time.Duration
}

func (c ConfigDuration) Duration() time.Duration {
	return c.duration
}

func ParseConfigDuration(s string) (ConfigDuration, error) {
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		return ConfigDuration{}, ErrConfigDurationSignInvalid
	}

	// normalize to be flexible with H vs h, M vs m, etc.
	s = strings.ToLower(s)

	duration, err := time.ParseDuration(s)
	if err != nil {
		return ConfigDuration{}, ErrConfigDurationInvalid
	}

	return ConfigDuration{duration: duration}, nil
}

func (c ConfigDuration) String() string {
	return c.duration.String()
}

func (c *ConfigDuration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	duration, err := ParseConfigDuration(s)
	if err != nil {
		return err
	}

	*c = duration
	return nil
}
