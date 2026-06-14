package schedule

import (
	"testing"

	"git.sr.ht/~mytec/gosched/internal/types"
)

func TestWorkflowFailureModeChecks(t *testing.T) {
	tests := []struct {
		name         string
		onFailure    types.FailureMode
		wantAbort    bool
		wantContinue bool
		wantRetry    bool
	}{
		{
			name:      "abort",
			onFailure: types.Abort,
			wantAbort: true,
		},
		{
			name:         "continue",
			onFailure:    types.Continue,
			wantContinue: true,
		},
		{
			name:      "retry",
			onFailure: types.Retry,
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
