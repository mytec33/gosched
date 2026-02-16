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

func TestLimitLength_PanicOnNegative(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for negative max")
		}
	}()

	_ = LimitLength(-1)
}

func TestLimitValue(t *testing.T) {
	tests := []struct {
		name      string
		maxValue  int
		value     int
		wantError error
	}{
		{name: "equal ok", maxValue: 5, value: 5, wantError: nil},
		{name: "below ok", maxValue: 5, value: 4, wantError: nil},
		{name: "zero ok", maxValue: 5, value: 0, wantError: nil},
		{name: "just over", maxValue: 5, value: 6, wantError: ErrExceedsMaxLimit},
		{name: "max zero value one", maxValue: 0, value: 1, wantError: ErrExceedsMaxLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := LimitValue(tt.maxValue)

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

func TestRequireNonNegative(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		wantError error
	}{
		{name: "zero ok", value: 0, wantError: nil},
		{name: "greater than zero ok", value: 110, wantError: nil},
		{name: "less than zero", value: -1, wantError: ErrNegativeNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireNoNegative(tt.value)
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

func TestRequireValidTime(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError error
	}{
		{name: "valid time", value: "10:30", wantError: nil},
		{name: "empty time", value: "", wantError: ErrInvalidTime},
		{name: "missing hours", value: ":10", wantError: ErrInvalidTime},
		{name: "missing minutes", value: "9:", wantError: ErrInvalidTime},
		{name: "missing colon", value: "1020", wantError: ErrInvalidTime},
		{name: "missing colon with space", value: "10 20", wantError: ErrInvalidTime},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireValidTime(tt.value)
			if err != tt.wantError {
				t.Fatalf("%s: got %q, want %q", tt.name, err, tt.wantError)
			}
		})
	}
}
