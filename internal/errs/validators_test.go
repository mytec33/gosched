package errs

import (
	"testing"
)

func TestLimitLength(t *testing.T) {
	tests := []struct {
		name      string
		maxLength int
		value     string
		wantError error
	}{
		{name: "valid 1", maxLength: 5, value: "12345", wantError: nil},
		{name: "zero max and empty still valid", maxLength: 0, value: "", wantError: nil},
		{name: "invalid 1", maxLength: 0, value: "1", wantError: ErrTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := LimitLength(tt.maxLength)

			err := f(tt.value)
			if err != tt.wantError {
				t.Fatalf("%s: got %q, want %q", tt.name, err, tt.wantError)
			}
		})
	}
}

func TestRequireNoLeadingTrailingWhitespace(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError error
	}{
		{name: "empty string ok", value: "", wantError: nil},
		{name: "only whitespace ok", value: " \t", wantError: nil},
		{name: "leading whitespace", value: " abcd", wantError: ErrWhitespaceLeadingOrTrailing},
		{name: "trailing whitespace", value: "efgh ", wantError: ErrWhitespaceLeadingOrTrailing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireNoLeadingTrailingWhitespace(tt.value)
			if err != tt.wantError {
				t.Fatalf("%s: got %q, want %q", tt.name, err, tt.wantError)
			}
		})
	}
}

func TestRequireNonEmpty(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError error
	}{
		{name: "empty string ok", value: "", wantError: ErrEmpty},
		{name: "not empty with whitespace", value: " \t", wantError: nil},
		{name: "not empt with chars", value: "abcd", wantError: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireNonEmpty(tt.value)
			if err != tt.wantError {
				t.Fatalf("%s: got %q, want %q", tt.name, err, tt.wantError)
			}
		})
	}
}

func TestRequireNoWhitespace(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError error
	}{
		{name: "empty string ok", value: "", wantError: nil},
		{name: "string with no whitespace ok", value: "abcd", wantError: nil},
		{name: "string with whitespace", value: " \t", wantError: ErrWhitespaceAll},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireNoWhitespace(tt.value)
			if err != tt.wantError {
				t.Fatalf("%s: got %q, want %q", tt.name, err, tt.wantError)
			}
		})
	}
}
