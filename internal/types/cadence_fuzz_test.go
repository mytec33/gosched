package types

import "testing"

func FuzzParseCadence(f *testing.F) {
	// LLM provided
	// go test ./internal/types -fuzz=FuzzParseCadence -fuzztime=5m

	for _, seed := range []string{"1d", "1h", "15m", "", "0m", "60m", "abc", "999999999999m"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		c, err := ParseCadence(input)
		if err != nil {
			return
		}

		if c.Repetition < 1 || c.Repetition > 60 {
			t.Fatalf("repetition out of bounds: %d", c.Repetition)
		}

		switch c.Measure {
		case CadenceDay, CadenceHour, CadenceMinute:
		default:
			t.Fatalf("invalid measure: %v", c.Measure)
		}

		if _, err := ParseCadence(c.String()); err != nil {
			t.Fatalf("valid cadence did not round trip: %v", err)
		}
	})
}
