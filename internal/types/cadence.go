package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrCadenceEmpty          = errors.New("cadence string cannot be empty")
	ErrCadenceTooShort       = errors.New("cadence string must include both a number and a unit")
	ErrCadenceNumInvalid     = errors.New("cadence numeric value is invalid")
	ErrCadenceBoundsInvalid  = errors.New("cadence value must be between 1 and 60")
	ErrCadenceUnitInvalid    = errors.New("cadence unit must be d, h, or m")
	ErrCadenceDayExceeded    = errors.New("daily cadence cannot be greater than 1d")
	ErrCadenceHourExceeded   = errors.New("hourly cadence cannot be greater than 23h")
	ErrCadenceMinuteExceeded = errors.New("minute cadence cannot be greater than 59m")
)

type Cadence struct {
	Repetition int
	Measure    string
}

func (c *Cadence) validateCadence() error {
	if c.Repetition > 60 || c.Repetition < 1 {
		return ErrCadenceBoundsInvalid
	}

	switch c.Measure {
	case "d":
		if c.Repetition > 1 {
			return ErrCadenceDayExceeded
		}
	case "h":
		if c.Repetition > 23 {
			return ErrCadenceHourExceeded
		}
	case "m":
		if c.Repetition > 59 {
			return ErrCadenceMinuteExceeded
		}
	default:
		return ErrCadenceUnitInvalid
	}

	return nil
}

func ParseCadence(s string) (Cadence, error) {
	var cadence Cadence

	if s == "" {
		return Cadence{}, ErrCadenceEmpty
	}

	if len(s) < 2 {
		return Cadence{}, ErrCadenceTooShort
	}

	v, err := strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return Cadence{}, ErrCadenceNumInvalid
	}
	cadence.Repetition = v
	cadence.Measure = strings.ToLower(string(s[len(s)-1:]))

	err = cadence.validateCadence()
	if err != nil {
		return Cadence{}, err
	}

	return cadence, nil
}

func (c *Cadence) String() string {
	return fmt.Sprintf("%d%s", c.Repetition, c.Measure)
}

func (c *Cadence) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	v, err := ParseCadence(s)
	if err != nil {
		return err
	}

	*c = v
	return nil
}
