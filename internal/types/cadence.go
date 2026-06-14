package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrCadenceEmpty          = errors.New("cadence string cannot be empty")
	ErrCadenceTooShort       = errors.New("cadence string must include both a number and a unit")
	ErrCadenceTooLong        = errors.New("cadence string is too long")
	ErrCadenceNumNonASCII    = errors.New("cadence numeric value must use ASCII digits 0-9")
	ErrCadenceNumInvalid     = errors.New("cadence numeric value is invalid")
	ErrCadenceBoundsInvalid  = errors.New("cadence value must be between 1 and 60")
	ErrCadenceUnitInvalid    = errors.New("cadence unit must be d, h, or m")
	ErrCadenceDayExceeded    = errors.New("daily cadence cannot be greater than 1d")
	ErrCadenceHourExceeded   = errors.New("hourly cadence cannot be greater than 23h")
	ErrCadenceMinuteExceeded = errors.New("minute cadence cannot be greater than 59m")
)

// Unexported fields due to these values changing or being overwritten could have a huge impact
// on the program. A repetition of 0 could be an endless loop. It was worth protecting this
// further than the typical opaque pattern like MinuteOfDay.

type Cadence struct {
	repetition int
	measure    CadenceMeasure
}

func (c Cadence) Repetition() int {
	return c.repetition
}

func (c Cadence) Measure() CadenceMeasure {
	return c.measure
}

func (c Cadence) validateCadence() error {
	switch c.measure {
	case CadenceDay:
		if c.repetition > 1 {
			return ErrCadenceDayExceeded
		}
	case CadenceHour:
		if c.repetition > 23 {
			return ErrCadenceHourExceeded
		}
	case CadenceMinute:
		if c.repetition > 59 {
			return ErrCadenceMinuteExceeded
		}
	default:
		return ErrCadenceUnitInvalid
	}

	return nil
}

func validateCadenceNumberPayload(num string) error {
	for _, r := range num {
		switch {
		case r > unicode.MaxASCII:
			return ErrCadenceNumNonASCII
		case r < '0' || r > '9':
			return ErrCadenceNumInvalid
		}
	}

	return nil
}

func parseMeasure(unit string) (CadenceMeasure, error) {
	switch unit {
	case "d":
		return CadenceDay, nil
	case "h":
		return CadenceHour, nil
	case "m":
		return CadenceMinute, nil
	default:
		return CadenceMeasure{}, ErrCadenceUnitInvalid
	}
}

func ParseCadence(s string) (Cadence, error) {
	if s == "" {
		return Cadence{}, ErrCadenceEmpty
	}

	if len(s) < 2 {
		return Cadence{}, ErrCadenceTooShort
	}

	num := s[:len(s)-1]
	if err := validateCadenceNumberPayload(num); err != nil {
		return Cadence{}, err
	}

	// Done after we look for unicode values to report them as unicode rather than "too long"
	if len(s) > 3 {
		return Cadence{}, ErrCadenceTooLong
	}

	v, err := strconv.Atoi(num)
	if err != nil {
		return Cadence{}, ErrCadenceNumInvalid
	}

	if v > 60 || v < 1 {
		return Cadence{}, ErrCadenceBoundsInvalid
	}

	measure, err := parseMeasure(strings.ToLower(string(s[len(s)-1:])))
	if err != nil {
		return Cadence{}, err
	}

	cadence := Cadence{
		repetition: v,
		measure:    measure,
	}
	if err := cadence.validateCadence(); err != nil {
		return Cadence{}, err
	}

	return cadence, nil
}

func (c Cadence) String() string {
	return fmt.Sprintf("%d%s", c.repetition, c.measure)
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
