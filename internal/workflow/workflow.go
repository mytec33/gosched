// Package workflow defines trusted workflow values and the validation rules
// used to build them from raw configuration.
package workflow

type Workflow struct {
	SourceFile     string
	Name           string
	Enabled        bool
	DisabledReason string
	Trigger        Trigger
	OnFailure      FailureMode
	Retry          RetryPolicy
	Steps          []Step
}

func WorkflowAbortsOnFailure(wf Workflow) bool {
	return wf.OnFailure == Abort
}

func WorkflowContinuesOnFailure(wf Workflow) bool {
	return wf.OnFailure == Continue
}

func WorkflowRetriesOnFailure(wf Workflow) bool {
	return wf.OnFailure == Retry
}
