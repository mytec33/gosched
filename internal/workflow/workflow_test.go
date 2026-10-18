package workflow

import (
	"testing"
)

func TestWorkflowStepCount(t *testing.T) {
	wf := Workflow{
		Steps: []Step{{}, {}},
	}

	if got := wf.StepCount(); got != 2 {
		t.Fatalf("StepCount() = %d, want 2", got)
	}
}
