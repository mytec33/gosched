package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrCadenceEmpty          = errors.New("cadence string cannot be empty")
	ErrCadenceTooShort       = errors.New("cadence string must include both a number and a unit")
	ErrCadenceNumNonASCII    = errors.New("cadence numeric value must use ASCII digits 0-9")
	ErrCadenceNumInvalid     = errors.New("cadence numeric value is invalid")
	ErrCadenceBoundsInvalid  = errors.New("cadence value of zero found must be between 1 and 60")
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

func (c Cadence) IntervalMinutes() int {
	repetition := c.Repetition()
	measure := c.Measure()

	switch measure {
	case CadenceDay:
		return repetition * MinutesInDay
	case CadenceHour:
		return repetition * 60
	case CadenceMinute:
		return repetition
	default:
		panic("workflow.Cadence error: invalid unit")
	}
}

func (c Cadence) Repetition() int {
	if c.repetition < 1 {
		panic(fmt.Sprintf("workflow.Cadence error: invalid repetition: %v", c.repetition))
	}
	return c.repetition
}

func (c Cadence) Measure() CadenceMeasure {
	if c.measure == CadenceUnknown {
		panic("workflow.Cadence error: invalid measure")
	}
	return c.measure
}

func (c Cadence) validateCadence() error {
	if c.repetition < 1 {
		return ErrCadenceBoundsInvalid
	}

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

func parseRepetition(num string) (int, error) {
	for _, r := range num {
		switch {
		case r > unicode.MaxASCII:
			return 0, ErrCadenceNumNonASCII
		case r < '0' || r > '9':
			return 0, ErrCadenceNumInvalid
		}
	}

	repetition, err := strconv.Atoi(num)
	if err != nil {
		return 0, ErrCadenceNumInvalid
	}

	return repetition, nil
}

func ParseCadence(s string) (Cadence, error) {
	if s == "" {
		return Cadence{}, ErrCadenceEmpty
	}

	if utf8.RuneCountInString(s) < 2 {
		return Cadence{}, ErrCadenceTooShort
	}

	_, lastCharSize := utf8.DecodeLastRuneInString(s)

	rawNum := s[:len(s)-lastCharSize]
	repetition, err := parseRepetition(rawNum)
	if err != nil {
		return Cadence{}, err
	}

	last := strings.ToLower(s[len(s)-lastCharSize:])
	measure, err := parseMeasure(last)
	if err != nil {
		return Cadence{}, err
	}

	cadence := Cadence{
		repetition: repetition,
		measure:    measure,
	}
	err = cadence.validateCadence()
	if err != nil {
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
