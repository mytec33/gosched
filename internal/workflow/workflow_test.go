package workflow

import (
	"testing"
)

func TestWorkflowFailureModeChecks(t *testing.T) {
	tests := []struct {
		name         string
		onFailure    FailureMode
		wantAbort    bool
		wantContinue bool
		wantRetry    bool
	}{
		{
			name:      "abort",
			onFailure: Abort,
			wantAbort: true,
		},
		{
			name:         "continue",
			onFailure:    Continue,
			wantContinue: true,
		},
		{
			name:      "retry",
			onFailure: Retry,
			wantRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf := Workflow{OnFailure: tt.onFailure}

			if got := WorkflowAbortsOnFailure(wf); got != tt.wantAbort {
				t.Fatalf("WorkflowAbortsOnFailure() = %v, want %v", got, tt.wantAbort)
			}

			if got := WorkflowContinuesOnFailure(wf); got != tt.wantContinue {
				t.Fatalf("WorkflowContinuesOnFailure() = %v, want %v", got, tt.wantContinue)
			}

			if got := WorkflowRetriesOnFailure(wf); got != tt.wantRetry {
				t.Fatalf("WorkflowRetriesOnFailure() = %v, want %v", got, tt.wantRetry)
			}
		})
	}
}

func TestWorkflowStepCount(t *testing.T) {
	wf := Workflow{
		Steps: []Step{{}, {}},
	}

	if got := wf.StepCount(); got != 2 {
		t.Fatalf("StepCount() = %d, want 2", got)
	}
}
