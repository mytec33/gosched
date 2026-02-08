// Package schedule provides schedule layout and locking
package schedule

import "errors"

type Schedule struct {
	wf map[MinuteKey][]Workflow
}

func (s Schedule) WorkflowCount() int {
	count := 0
	for _, wfs := range s.wf {
		count += len(wfs)
	}
	return count
}

func (s Schedule) WorkflowsAtMinute(k MinuteKey) []Workflow {
	return s.wf[k]
}

type MinuteKey string

type Workflow struct {
	Name  string `json:"name"`
	Time  string `json:"time"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Name    string `json:"name"`
	Program string `json:"program"`
	Args    string `json:"args"`
}

func (w Workflow) Validate() []error {

	return []error{errors.New("foo")}
}

func ValidateSchedule(s Schedule) []error {
	if len(s.wf) == 0 {
		return []error{errors.New("no schedule found")}
	}

	return []error{}
}
