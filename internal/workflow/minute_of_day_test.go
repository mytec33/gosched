package workflow

import (
	"errors"
	"testing"
	"time"
)

func TestMinuteOfDayFromTime(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want string
	}{
		{
			name: "midnight",
			in:   time.Date(2026, time.June, 14, 0, 0, 0, 0, time.Local),
			want: "00:00",
		},
		{
			name: "middle of day ignores seconds and nanos",
			in:   time.Date(2026, time.June, 14, 11, 45, 59, 123, time.Local),
			want: "11:45",
		},
		{
			name: "last minute of day",
			in:   time.Date(2026, time.June, 14, 23, 59, 0, 0, time.Local),
			want: "23:59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MinuteOfDayFromTime(tt.in)
			if got.String() != tt.want {
				t.Fatalf("MinuteOfDayFromTime(%v).String() = %q, want %q", tt.in, got.String(), tt.want)
			}
		})
	}
}

func TestParseMinuteOfDay_Valid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "midnight", input: "00:00", want: "00:00"},
		{name: "first minute", input: "00:01", want: "00:01"},
		{name: "zero-padded hour", input: "08:00", want: "08:00"},
		{name: "non-padded hour normalizes", input: "8:00", want: "08:00"},
		{name: "middle of day", input: "11:45", want: "11:45"},
		{name: "last minute of day", input: "23:59", want: "23:59"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMinuteOfDay(tt.input)
			if err != nil {
				t.Fatalf("ParseMinuteOfDay(%q) unexpected error = %v", tt.input, err)
			}

			if got.String() != tt.want {
				t.Fatalf("ParseMinuteOfDay(%q).String() = %q, want %q", tt.input, got.String(), tt.want)
			}
		})
	}
}

func TestParseMinuteOfDay_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: ErrTimeFormatInvalid,
		},
		{
			name:    "leading whitespace",
			input:   " 10:00",
			wantErr: ErrTimeFormatInvalid,
		},
		{
			name:    "trailing whitespace",
			input:   "10:00 ",
			wantErr: ErrTimeFormatInvalid,
		},
		{
			name:    "invalid hour",
			input:   "24:00",
			wantErr: ErrTimeFormatInvalid,
		},
		{
			name:    "invalid minute",
			input:   "12:60",
			wantErr: ErrTimeFormatInvalid,
		},
		{
			name:    "missing colon",
			input:   "105",
			wantErr: ErrTimeFormatInvalid,
		},
		{
			name:    "missing second minute digit",
			input:   "10:5",
			wantErr: ErrTimeFormatInvalid,
		},
		{
			name:    "invalid input",
			input:   "abc",
			wantErr: ErrTimeFormatInvalid,
		},
		{
			name:    "invalid milliseconds",
			input:   "6:45.000",
			wantErr: ErrTimeFormatInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseMinuteOfDay(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseMinuteOfDay(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestMinuteOfDayMinutesSince(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		previous string
		want     int
	}{
		{
			name:     "same minute",
			current:  "10:00",
			previous: "10:00",
			want:     0,
		},
		{
			name:     "next minute",
			current:  "10:01",
			previous: "10:00",
			want:     1,
		},
		{
			name:     "skipped minutes",
			current:  "10:05",
			previous: "10:00",
			want:     5,
		},
		{
			name:     "midnight wrap",
			current:  "00:01",
			previous: "23:59",
			want:     2,
		},
		{
			name:     "previous one minute ahead wraps to max",
			current:  "10:00",
			previous: "10:01",
			want:     1439,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, err := ParseMinuteOfDay(tt.current)
			if err != nil {
				t.Fatalf("ParseMinuteOfDay(%q) unexpected error = %v", tt.current, err)
			}

			previous, err := ParseMinuteOfDay(tt.previous)
			if err != nil {
				t.Fatalf("ParseMinuteOfDay(%q) unexpected error = %v", tt.previous, err)
			}

			got := current.MinutesSince(previous)
			if got != tt.want {
				t.Fatalf("%v.MinutesSince(%v) = %d, want %d", current, previous, got, tt.want)
			}
		})
	}
}
