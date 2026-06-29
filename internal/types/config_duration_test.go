package types

import (
	"errors"
	"testing"
	"time"
)

/*
These testse are not intended to test time.Duration but to show how type ConfigDuration differs
from the type it mimics, time.Duration.
*/
func TestParseConfigDuration(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantDuration time.Duration
	}{
		{"uppercase unit normalization - hour", "15H", 15 * time.Hour},     // accepted, converted to 15h with same meaning
		{"uppercase unit normalization - minute", "10M", 10 * time.Minute}, // accepted, converted to 10m with same meaning
		{"uppercase unit normalization - second", "30S", 30 * time.Second}, // accepted, converted to 30s with same meaning

		// These are allowed in duration whereas Cadence doesn't allow
		{"valid hours > 24", "48h", 48 * time.Hour},
		{"valid minutes 0", "0m", 0 * time.Minute},
		{"valid minutes 99", "99m", 99 * time.Minute},
		{"valid seconds", "7200s", 2 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfigDuration(tt.input)
			if err != nil {
				t.Fatalf("ParseConfigDuration(%q) unexpected error = %v", tt.input, err)
			}

			if got.Duration() != tt.wantDuration {
				t.Errorf("ParseConfigDuration(%q) want %q", tt.input, tt.wantDuration)
			}
		})
	}
}

func TestParseConfigDuration_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"empty string", "", ErrConfigDurationInvalid},
		{"invalid minus symbol", "-1h", ErrConfigDurationSignInvalid},
		{"invalid plus symbol", "+2s", ErrConfigDurationSignInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseConfigDuration(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseConfigDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
