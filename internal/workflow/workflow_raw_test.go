package workflow

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func assertOnlyValidationError(t *testing.T, validationErrors []error, want error) {
	t.Helper()

	if len(validationErrors) == 0 {
		t.Fatalf("expected validation error %v, got none", want)
	}

	found := false
	for _, err := range validationErrors {
		if !errors.Is(err, want) {
			t.Fatalf("unexpected validation error: %v (expected only %v). Full list: %v",
				err, want, validationErrors)
		}

		found = true
	}

	if !found {
		t.Fatalf("missing expected error %v. Full list: %v", want, validationErrors)
	}
}

func validWorkflowRaw() WorkflowRaw {
	onFailure := Continue
	enabled := true
	every, _ := ParseCadence("1d")
	beginAt, _ := ParseMinuteOfDay("10:35")

	return WorkflowRaw{
		Name:    "Workflow",
		Enabled: &enabled,
		Trigger: &TriggerRaw{
			Every:   &every,
			BeginAt: &beginAt,
		},
		Retry: &RetryPolicy{
			NumberRetries: 0,
			PauseSeconds:  0,
		},
		OnFailure: &onFailure,
		Steps: []Step{
			{Name: "Step", Program: "program", Args: []string{"arg"}},
		},
	}
}

func TestWorkflowValidate(t *testing.T) {
	// only select whitespace tests are included as the validateStringValue function
	// is tested heavily on its own. Leave a few here to prove higher level wiring works.
	tests := []struct {
		name      string
		mutate    func(*WorkflowRaw)
		wantError error
	}{
		{
			name: "workflow name empty",
			mutate: func(wf *WorkflowRaw) {
				wf.Name = ""
			},
			wantError: ErrFieldEmpty,
		},
		{
			name: "workflow steps missing",
			mutate: func(wf *WorkflowRaw) {
				wf.Steps = nil
			},
			wantError: ErrStepsRequired,
		},
		{
			name: "workflow too many steps",
			mutate: func(wf *WorkflowRaw) {
				wf.Steps = repeatedSteps(MaxStepsCount + 1)
			},
			wantError: ErrStepCountExceeded,
		},
		{
			name: "enabled missing",
			mutate: func(wf *WorkflowRaw) {
				wf.Enabled = nil
			},
			wantError: ErrEnabledRequired,
		},
		{
			name: "disabled without reason",
			mutate: func(wf *WorkflowRaw) {
				enabled := false
				wf.Enabled = &enabled
			},
			wantError: ErrDisabledReasonRequired,
		},
		{
			name: "disabled reason on enabled workflow",
			mutate: func(wf *WorkflowRaw) {
				reason := "should not be allowed"
				wf.DisabledReason = &reason
			},
			wantError: ErrDisabledReasonNotAllowed,
		},
		{
			name: "disabled reason whitespace only",
			mutate: func(wf *WorkflowRaw) {
				enabled := false
				reason := "\t\n "
				wf.Enabled = &enabled
				wf.DisabledReason = &reason
			},
			wantError: ErrFieldWhitespaceOnly,
		},
		{
			name: "onFailure missing",
			mutate: func(wf *WorkflowRaw) {
				wf.OnFailure = nil
			},
			wantError: ErrOnFailureRequired,
		},
		{
			name: "retry policy requires retry config",
			mutate: func(wf *WorkflowRaw) {
				onFailure := Retry
				wf.OnFailure = &onFailure
				wf.Retry = nil
			},
			wantError: ErrRetryRequired,
		},
		{
			name: "workflow retry number retries negative",
			mutate: func(wf *WorkflowRaw) {
				wf.Retry.NumberRetries = -1
			},
			wantError: ErrRetryCountNegative,
		},
		{
			name: "workflow retry number retries too large",
			mutate: func(wf *WorkflowRaw) {
				wf.Retry.NumberRetries = MaxRetryCount + 1
			},
			wantError: ErrRetryCountTooLarge,
		},
		{
			name: "workflow retry pause seconds negative",
			mutate: func(wf *WorkflowRaw) {
				wf.Retry.PauseSeconds = -1
			},
			wantError: ErrRetryPauseNegative,
		},
		{
			name: "workflow retry pause seconds too large",
			mutate: func(wf *WorkflowRaw) {
				wf.Retry.PauseSeconds = MaxRetryPauseLimit + 1
			},
			wantError: ErrRetryPauseTooLarge,
		},
		{
			name: "step name whitespace",
			mutate: func(wf *WorkflowRaw) {
				wf.Steps[0].Name = "\t\n "
			},
			wantError: ErrFieldWhitespaceOnly,
		},
		{
			name: "program whitespace leading",
			mutate: func(wf *WorkflowRaw) {
				wf.Steps[0].Program = " program"
			},
			wantError: ErrFieldWhitespacePadded,
		},
		{
			name: "step arg invalid",
			mutate: func(wf *WorkflowRaw) {
				wf.Steps[0].Args = []string{"arg "}
			},
			wantError: ErrFieldWhitespacePadded,
		},
		{
			name: "step name duplicate",
			mutate: func(wf *WorkflowRaw) {
				wf.Steps = append(wf.Steps, Step{Name: "Step", Program: "program"})
			},
			wantError: ErrStepDuplicateName,
		},
		{
			name: "trigger block missing",
			mutate: func(wf *WorkflowRaw) {
				wf.Trigger = nil
			},
			wantError: ErrTriggerRequired,
		},
		{
			name: "trigger every missing",
			mutate: func(wf *WorkflowRaw) {
				wf.Trigger.Every = nil
			},
			wantError: ErrTriggerEveryRequired,
		},
		{
			name: "trigger beginAt missing",
			mutate: func(wf *WorkflowRaw) {
				wf.Trigger.BeginAt = nil
			},
			wantError: ErrTriggerBeginAtRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf := validWorkflowRaw()
			tt.mutate(&wf)

			sourceFileName := "not used"
			_, validationErrors := wf.Validate(sourceFileName)

			assertOnlyValidationError(t, validationErrors, tt.wantError)
		})
	}
}

func TestWorkflowValidate_DisabledValid(t *testing.T) {
	wf := validWorkflowRaw()
	enabled := false
	reason := "maintenance window"
	wf.Enabled = &enabled
	wf.DisabledReason = &reason

	sourceFile := "not used"
	trusted, validationErrors := wf.Validate(sourceFile)
	if len(validationErrors) != 0 {
		t.Fatalf("expected no validation errors, got %v", validationErrors)
	}

	if trusted.Enabled {
		t.Fatal("expected trusted workflow to be disabled")
	}

	if trusted.DisabledReason != reason {
		t.Fatalf("DisabledReason = %q, want %q", trusted.DisabledReason, reason)
	}
}

func TestValidateStringValue(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		maxLength int
		wantError error
	}{
		{"valid", "workflow", 20, nil},
		{"empty", "", 20, ErrFieldEmpty},
		{"whitespace only", "\t\n ", 20, ErrFieldWhitespaceOnly},
		{"leading whitespace", " workflow", 20, ErrFieldWhitespacePadded},
		{"trailing whitespace", "workflow ", 20, ErrFieldWhitespacePadded},
		{"too long", "workflow", 3, ErrFieldTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateStringValue("field", tt.value, tt.maxLength)

			if tt.wantError == nil {
				if len(errs) != 0 {
					t.Fatalf("validateStringValue() errors = %v, want none", errs)
				}
				return
			}

			assertOnlyValidationError(t, errs, tt.wantError)
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
		{"arg empty", []string{""}, ErrFieldEmpty},
		{"arg whitespace only", []string{"\t\n "}, ErrFieldWhitespaceOnly},
		{"arg leading whitespace", []string{" arg"}, ErrFieldWhitespacePadded},
		{"arg trailing whitespace", []string{"arg "}, ErrFieldWhitespacePadded},
		{"arg too long", []string{strings.Repeat("a", MaxStepArgLength+1)}, ErrFieldTooLong},
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

			assertOnlyValidationError(t, errs, tt.wantError)
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

func repeatedSteps(count int) []Step {
	steps := make([]Step, count)
	for i := range steps {
		steps[i] = Step{Name: fmt.Sprintf("Step %d", i+1), Program: "program"}
	}

	return steps
}
