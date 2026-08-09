package workflow

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func assertHasValidationError(t *testing.T, validationErrors []error, want error) {
	t.Helper()

	for _, err := range validationErrors {
		if errors.Is(err, want) {
			return
		}
	}

	t.Fatalf("missing expected error %v in %v", want, validationErrors)
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
		Steps: []StepRaw{
			{Name: "Step", Program: "program", Args: []string{"arg"}},
		},
	}
}

func TestWorkflowValidate(t *testing.T) {
	// only select whitespace tests are included as the validateWorkflowStringValue function
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
			name: "workflow invalid step rejected",
			mutate: func(wf *WorkflowRaw) {
				wf.Steps[0].Name = ""
			},
			wantError: ErrStepFieldEmpty,
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
			wantError: ErrStepsCountExceeded,
		},
		{
			name: "step name duplicate",
			mutate: func(wf *WorkflowRaw) {
				wf.Steps = append(wf.Steps, StepRaw{Name: "Step", Program: "program"})
			},
			wantError: ErrStepsDuplicateName,
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
			got, validationErrors := wf.Validate(sourceFileName)

			assertHasValidationError(t, validationErrors, tt.wantError)

			if !reflect.DeepEqual(got, Workflow{}) {
				t.Fatalf("got %v, expected no trusted workflow", got)
			}
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

func TestValidateWorkflowStringValue(t *testing.T) {
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
			errs := validateWorkflowStringValue("field", tt.value, tt.maxLength)

			if tt.wantError == nil {
				if len(errs) != 0 {
					t.Fatalf("validateWorkflowStringValue() errors = %v, want none", errs)
				}
				return
			}

			assertHasValidationError(t, errs, tt.wantError)
		})
	}
}

func repeatedSteps(count int) []StepRaw {
	steps := make([]StepRaw, count)
	for i := range steps {
		steps[i] = StepRaw{Name: fmt.Sprintf("Step %d", i+1), Program: "program"}
	}

	return steps
}
