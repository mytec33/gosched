package types

import (
	"errors"
	"testing"
)

func TestParseCadence(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantRep     int
		wantMeasure CadenceMeasure
	}{
		{"valid minute", "45m", 45, CadenceMinute},
		{"valid hour", "12h", 12, CadenceHour},
		{"valid day", "1d", 1, CadenceDay},
		{"uppercase unit normalization", "15H", 15, CadenceHour},          // accepted, converted to 15h with same meaning
		{"uppercase leading zero normalization", "05m", 5, CadenceMinute}, // accepted, normalizes to 5m
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCadence(tt.input)
			if err != nil {
				t.Fatalf("ParseCadence(%q) unexpected error = %v", tt.input, err)
			}

			if got.Repetition() != tt.wantRep {
				t.Errorf("ParseCadence(%q) Repetition = %d, want %d", tt.input, got.Repetition(), tt.wantRep)
			}
			if got.Measure() != tt.wantMeasure {
				t.Errorf("ParseCadence(%q) Measure = %q, want %q", tt.input, got.Measure(), tt.wantMeasure)
			}
		})
	}
}

func TestParseCadence_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"empty string", "", ErrCadenceEmpty},
		{"too short missing unit", "5", ErrCadenceTooShort},
		{"invalid integer payload", "abcde", ErrCadenceNumInvalid},
		{"non-numeric prefix", "1a2h", ErrCadenceNumInvalid},
		{"invalid measure unit", "12s", ErrCadenceUnitInvalid},
		{"repetition max breach", "99m", ErrCadenceMinuteExceeded},
		{"repetition min breach", "0m", ErrCadenceBoundsInvalid},
		{"day boundary breach", "2d", ErrCadenceDayExceeded},
		{"hour boundary breach", "24h", ErrCadenceHourExceeded},
		{"minute boundary breach", "60m", ErrCadenceMinuteExceeded},
		{"leading minus", "-1d", ErrCadenceNumInvalid},
		{"leading plus", "+1d", ErrCadenceNumInvalid},
		{"space between repetition and unit", "5 m", ErrCadenceNumInvalid},
		{"leading spaces", " 3m", ErrCadenceNumInvalid},
		{"leading tab", "\t3m", ErrCadenceNumInvalid},
		{"unicode digit", "１m", ErrCadenceNumNonASCII},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCadence(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseCadence(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
