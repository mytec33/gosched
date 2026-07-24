package workflow

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestFailureMode_StringMethod(t *testing.T) {
	tests := []struct {
		name string
		mode FailureMode
		want string
	}{
		{"abort", Abort, "abort"},
		{"continue", Continue, "continue"},
		{"retry", Retry, "retry"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Fatalf("%s.String() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestFailureMode_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  FailureMode
	}{
		{"abort", `"abort"`, Abort},
		{"continue", `"continue"`, Continue},
		{"retry", `"retry"`, Retry},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got FailureMode
			err := json.Unmarshal([]byte(tt.input), &got)
			if err != nil {
				t.Fatalf("json.Unmarshal(%s) unexpected error = %v", tt.input, err)
			}

			if got != tt.want {
				t.Fatalf("json.Unmarshal(%s) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFailureMode_UnmarshalJSON_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"empty string", `""`, ErrOnFailureInvalid},
		{"unknown failure mode", `"stop"`, ErrOnFailureInvalid},
		{"uppercase failure mode", `"ABORT"`, ErrOnFailureInvalid},
		{"non-string json value", `123`, new(json.UnmarshalTypeError)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got FailureMode
			err := json.Unmarshal([]byte(tt.input), &got)

			if !matchesFailureModeUnmarshalError(err, tt.wantErr) {
				t.Fatalf("json.Unmarshal(%s) error = %v, wantErr %T", tt.input, err, tt.wantErr)
			}
		})
	}
}

func matchesFailureModeUnmarshalError(got error, want error) bool {
	var typeErr *json.UnmarshalTypeError
	if errors.As(want, &typeErr) {
		return errors.As(got, &typeErr)
	}

	return errors.Is(got, want)
}
