package schedule

import "errors"

type MinuteKey string

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

func (s Schedule) Validate() []error {
	if len(s.wf) == 0 {
		return []error{errors.New("no schedule found")}
	}

	return []error{}
}
