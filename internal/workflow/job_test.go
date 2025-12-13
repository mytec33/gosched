package workflow

import (
	"encoding/json"
	"testing"
)

type ErrorCode string

const (
	ErrCodeMissingField    ErrorCode = "missing_field"
	ErrCodeDuplicateName   ErrorCode = "duplicate_name"
	ErrCodeUnknownJob      ErrorCode = "unknown_job"
	ErrCodeDependencyCycle ErrorCode = "dependency_cycle"
	ErrCodeInvalidSchedule ErrorCode = "invalid_schedule"
	// etc
)

func TestValidation_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantCode ErrorCode
	}{
		{
			name: "no workflows found",
			json: ``,
		},
		{
			name: "no workflows found 2",
			json: `[{}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var schedule []Workflow

			err := json.Unmarshal([]byte(tt.json), &schedule)
			if err == nil {
				t.Fatalf("%v: expected error, got nil", tt.name)
			} else {
				t.Fatalf("%v: got error, len of schedule %v", tt.name, len(schedule))
			}
		})
	}
}
