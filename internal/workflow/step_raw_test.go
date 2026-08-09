package workflow

import (
	"reflect"
	"strings"
	"testing"
)

func TestStepValidate_Valid(t *testing.T) {
	timeoutDuration, err := ParseConfigDuration("0h1m")
	if err != nil {
		t.Fatalf("expected no error parsing configduration, got %v", err)
	}

	pauseDuration, err := ParseConfigDuration("1h")
	if err != nil {
		t.Fatalf("expected no error parsing configduration, got %v", err)
	}

	tests := []struct {
		name string
		data StepRaw
		want Step
	}{
		{
			name: "step valid all fields provided",
			data: StepRaw{
				Name: "step name", Program: "program name", Args: []string{"one", "two"},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			want: Step{Name: "step name", Program: "program name", Args: []string{"one", "two"},
				Timeout: timeoutDuration, Pause: pauseDuration,
			},
		},
		{
			name: "step valid with timeout omitted",
			data: StepRaw{
				Name: "step name", Program: "program name", Args: []string{"one", "two"},
				Timeout: nil, Pause: pauseDuration,
			},
			want: Step{Name: "step name", Program: "program name", Args: []string{"one", "two"},
				Timeout: defaultTimeout, Pause: pauseDuration,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errorList := tt.data.Validate()
			if len(errorList) > 0 {
				t.Fatalf("got %v, expected no errors", errorList)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, expected %v", got, tt.want)
			}
		})
	}
}

func TestStepValidate_Invalid(t *testing.T) {
	timeoutDuration, err := ParseConfigDuration("0h1m")
	if err != nil {
		t.Fatalf("expected no error parsing configduration, got %v", err)
	}

	pauseDuration, err := ParseConfigDuration("1h")
	if err != nil {
		t.Fatalf("expected no error parsing configduration, got %v", err)
	}

	pauseDurationInvalid, err := ParseConfigDuration("1h1m")
	if err != nil {
		t.Fatalf("expected no error parsing configduration, got %v", err)
	}

	tooManyArgs := make([]string, MaxStepArgsCount+1)
	for i := range tooManyArgs {
		tooManyArgs[i] = "arg"
	}

	tests := []struct {
		name      string
		data      StepRaw
		wantError error
	}{
		{
			name: "step name empty",
			data: StepRaw{
				Name: "", Program: "program name", Args: []string{"one", "two"},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepFieldEmpty,
		},
		{
			name: "step name too long",
			data: StepRaw{
				Name: strings.Repeat("a", MaxStepNameLength+1), Program: "program name", Args: []string{"one", "two"},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepFieldTooLong,
		},
		{
			name: "step name whitespace",
			data: StepRaw{
				Name: "\t\n ", Program: "program name", Args: []string{"one", "two"},
				Timeout: nil, Pause: pauseDuration,
			},
			wantError: ErrStepFieldWhitespaceOnly,
		},
		{
			name: "step name leading whitespace",
			data: StepRaw{
				Name: "   step name", Program: "program name", Args: []string{"one", "two"},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepFieldWhitespacePadded,
		},
		{
			name: "step name trailing whitespace",
			data: StepRaw{
				Name: "step name\n", Program: "program name", Args: []string{"one", "two"},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepFieldWhitespacePadded,
		},
		{
			name: "program name empty",
			data: StepRaw{
				Name: "step name", Program: "", Args: []string{"one", "two"},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepFieldEmpty,
		},
		{
			name: "program name too long",
			data: StepRaw{
				Name: "step name", Program: strings.Repeat("a", MaxStepProgramLength+1), Args: []string{"one", "two"},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepFieldTooLong,
		},
		{
			name: "program name whitespace",
			data: StepRaw{
				Name: "step name", Program: "\t\n ", Args: []string{"one"},
				Timeout: nil, Pause: pauseDuration,
			},
			wantError: ErrStepFieldWhitespaceOnly,
		},
		{
			name: "program name leading whitespace",
			data: StepRaw{
				Name: "step name", Program: " program name", Args: []string{"one", "two", "three"},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepFieldWhitespacePadded,
		},
		{
			name: "program name trailing whitespace",
			data: StepRaw{
				Name: "step name", Program: "program name\t", Args: []string{"one", "two"},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepFieldWhitespacePadded,
		},
		{
			name: "arg empty",
			data: StepRaw{
				Name: "step name", Program: "program name", Args: []string{"one", "", "three", "four"},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepFieldEmpty,
		},
		{
			name: "arg too many",
			data: StepRaw{
				Name: "step name", Program: "program name", Args: tooManyArgs,
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepArgsCountExceeded,
		},
		{
			name: "arg leading whitespace",
			data: StepRaw{
				Name: "step name", Program: "program name", Args: []string{"one", " two"},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepFieldWhitespacePadded,
		},
		{
			name: "arg trailing whitespace",
			data: StepRaw{
				Name: "step name", Program: "program name", Args: []string{"one", "two "},
				Timeout: &timeoutDuration, Pause: pauseDuration,
			},
			wantError: ErrStepFieldWhitespacePadded,
		},

		{
			name: "pause duration exceeded",
			data: StepRaw{
				Name: "step name", Program: "program name", Args: []string{"one", "two"},
				Timeout: &timeoutDuration, Pause: pauseDurationInvalid,
			},
			wantError: ErrStepPauseDurationTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errorList := tt.data.Validate()
			if len(errorList) == 0 {
				t.Fatalf("got no errors, expected errors")
			}

			assertHasValidationError(t, errorList, tt.wantError)
		})
	}
}

func TestValidateStepArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError error
	}{
		{"nil args allowed", nil, nil},
		{"empty args allowed", []string{}, nil},
		{"valid args", []string{"one", "two"}, nil},
		{"arg empty", []string{""}, ErrStepFieldEmpty},
		{"arg whitespace only", []string{"\t\n "}, ErrStepFieldWhitespaceOnly},
		{"arg leading whitespace", []string{" arg"}, ErrStepFieldWhitespacePadded},
		{"arg trailing whitespace", []string{"arg "}, ErrStepFieldWhitespacePadded},
		{"arg too long", []string{strings.Repeat("a", MaxStepArgLength+1)}, ErrStepFieldTooLong},
		{"too many args", repeatedArgs(MaxStepArgsCount+1, "arg"), ErrStepArgsCountExceeded},
		{
			name:      "total args too long",
			args:      repeatedArgs(17, strings.Repeat("a", 241)),
			wantError: ErrStepArgsTotalLengthExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateStepArgs("field", tt.args)

			if tt.wantError == nil {
				if len(errs) != 0 {
					t.Fatalf("validateStepArgs() errors = %v, want none", errs)
				}
				return
			}

			assertHasValidationError(t, errs, tt.wantError)
		})
	}
}

func repeatedArgs(count int, value string) []string {
	args := make([]string, count)
	for i := range args {
		args[i] = value
	}

	return args
}

func TestValidateStepStringsValue(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		maxLength int
		wantError error
	}{
		{"valid", "step", 20, nil},
		{"empty", "", 20, ErrStepFieldEmpty},
		{"whitespace only", "\t\n ", 20, ErrStepFieldWhitespaceOnly},
		{"leading whitespace", " workflow", 20, ErrStepFieldWhitespacePadded},
		{"trailing whitespace", "workflow ", 20, ErrStepFieldWhitespacePadded},
		{"too long", "program", 3, ErrStepFieldTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateStepStrings("field", tt.value, tt.maxLength)

			if tt.wantError == nil {
				if len(errs) != 0 {
					t.Fatalf("validateStepStrings() errors = %v, want none", errs)
				}
				return
			}

			assertHasValidationError(t, errs, tt.wantError)
		})
	}
}
