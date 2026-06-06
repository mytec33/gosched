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
		wantMeasure CadenceMeasure
		wantErr     error
	}{
		// Success cases
		{"valid minute", "45m", 45, CadenceMinute, nil},
		{"valid hour", "12h", 12, CadenceHour, nil},
		{"valid day", "1d", 1, CadenceDay, nil},
		{"uppercase normalization", "15H", 15, CadenceHour, nil},

		// Structural / Parsing failures
		{"empty string", "", 0, CadenceMeasure{}, ErrCadenceEmpty},
		{"too short missing unit", "5", 0, CadenceMeasure{}, ErrCadenceTooShort},
		{"invalid integer payload", "abcde", 0, CadenceMeasure{}, ErrCadenceNumInvalid},
		{"non-numeric prefix", "1a2h", 0, CadenceMeasure{}, ErrCadenceNumInvalid},

		// Business / Boundary failures
		{"Invalid measure unit", "12s", 0, CadenceMeasure{}, ErrCadenceUnitInvalid},
		{"repetition max breach", "99m", 0, CadenceMeasure{}, ErrCadenceBoundsInvalid},
		{"repetition min breach", "0m", 0, CadenceMeasure{}, ErrCadenceBoundsInvalid},
		{"day boundary breach", "2d", 0, CadenceMeasure{}, ErrCadenceDayExceeded},
		{"hour boundary breach", "24h", 0, CadenceMeasure{}, ErrCadenceHourExceeded},
		{"minute boundary breach", "60m", 0, CadenceMeasure{}, ErrCadenceMinuteExceeded},
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
