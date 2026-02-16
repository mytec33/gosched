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
