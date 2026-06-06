package types

import (
	"errors"
	"testing"
)

func TestParseCadence(t *testing.T) {
	// 1. Define the test case layout structure
	tests := []struct {
		name        string
		input       string
		wantRep     int
		wantMeasure string
		wantErr     error
	}{
		// Success cases
		{"Valid minute", "45m", 45, "m", nil},
		{"Valid hour", "12h", 12, "h", nil},
		{"Valid day", "1d", 1, "d", nil},
		{"Uppercase normalization", "15H", 15, "h", nil},

		// Structural / Parsing failures
		{"Empty string", "", 0, "", ErrCadenceEmpty},
		{"Too short missing unit", "5", 0, "", ErrCadenceTooShort},
		{"Invalid integer payload", "abcde", 0, "", ErrCadenceNumInvalid},
		{"Non-numeric prefix", "1a2h", 0, "", ErrCadenceNumInvalid},

		// Business / Boundary failures
		{"Invalid measure unit", "12s", 0, "", ErrCadenceUnitInvalid},
		{"Global repetition max breach", "99m", 0, "", ErrCadenceBoundsInvalid},
		{"Global repetition min breach", "0m", 0, "", ErrCadenceBoundsInvalid},
		{"Day boundary breach", "2d", 0, "", ErrCadenceDayExceeded},
		{"Hour boundary breach", "24h", 0, "", ErrCadenceHourExceeded},
		{"Minute boundary breach", "60m", 0, "", ErrCadenceMinuteExceeded},
	}

	// 2. Iterate through each test case sequentially
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCadence(tt.input)

			// Check for expected error matching
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseCadence(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}

			// If we didn't expect an error, verify the struct values are correct
			if tt.wantErr == nil {
				if got.Repetition != tt.wantRep {
					t.Errorf("ParseCadence(%q) Repetition = %d, want %d", tt.input, got.Repetition, tt.wantRep)
				}
				if got.Measure != tt.wantMeasure {
					t.Errorf("ParseCadence(%q) Measure = %q, want %q", tt.input, got.Measure, tt.wantMeasure)
				}
			}
		})
	}
}
