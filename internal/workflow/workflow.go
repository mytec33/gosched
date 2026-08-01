// Package workflow defines trusted workflow values and the validation rules
// used to build them from raw configuration.
package workflow

type Workflow struct {
	Name           string
	Enabled        bool
	DisabledReason string
	Trigger        Trigger
	OnFailure      FailureMode
	Retry          RetryPolicy
	Steps          []Step
}

type Step struct {
	Name    string         `json:"name"`
	Program string         `json:"program"`
	Args    []string       `json:"args"`
	Timeout ConfigDuration `json:"timeout"`
	Pause   ConfigDuration `json:"pause"`
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
