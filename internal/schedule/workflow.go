// Package schedule provides schedule layout and locking
package schedule

import (
	"git.sr.ht/~mytec/gosched/internal/policy"
	"git.sr.ht/~mytec/gosched/internal/types"
)

type Workflow struct {
	Name      string             `json:"name"`
	Trigger   types.Trigger      `json:"trigger"`
	OnFailure policy.FailureMode `json:"onFailure"`
	Retry     *RetryPolicy       `json:"retry"`
	Steps     []Step             `json:"steps"`
}

type RetryPolicy struct {
	NumberRetries int `json:"numberRetries"`
	PauseSeconds  int `json:"pauseSeconds"`
}

type Step struct {
	Name    string   `json:"name"`
	Program string   `json:"program"`
	Args    []string `json:"args"`
	Timeout int      `json:"timeout"`
	Pause   int      `json:"pause"`
}

func WorkflowAbortsOnFailure(wf Workflow) bool {
	return wf.OnFailure == policy.Abort
}

func WorkflowContinuesOnFailure(wf Workflow) bool {
	return wf.OnFailure == policy.Continue
}

func WorkflowRetriesOnFailure(wf Workflow) bool {
	return wf.OnFailure == policy.Retry
}
