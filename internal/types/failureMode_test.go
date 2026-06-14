package types

import (
	"encoding/json"
	"errors"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/errs"
)

func TestFailureModeString(t *testing.T) {
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

func TestFailureModeUnmarshalJSON(t *testing.T) {
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

func TestFailureModeUnmarshalJSON_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"empty string", `""`, errs.ErrOnFailureInvalid},
		{"unknown failure mode", `"stop"`, errs.ErrOnFailureInvalid},
		{"uppercase failure mode", `"ABORT"`, errs.ErrOnFailureInvalid},
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

func TestFailureModeMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		mode FailureMode
		want string
	}{
		{"abort", Abort, `"abort"`},
		{"continue", Continue, `"continue"`},
		{"retry", Retry, `"retry"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.mode)
			if err != nil {
				t.Fatalf("json.Marshal(%s) unexpected error = %v", tt.name, err)
			}

			if string(got) != tt.want {
				t.Fatalf("json.Marshal(%s) = %s, want %s", tt.name, got, tt.want)
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
