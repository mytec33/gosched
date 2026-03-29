package schedule

import "git.sr.ht/~mytec/gosched/internal/types"

type Schedule struct {
	wf map[types.MinuteOfDay][]Workflow
}

func (s Schedule) WorkflowCount() int {
	count := 0
	for _, wfs := range s.wf {
		count += len(wfs)
	}
	return count
}

func (s Schedule) WorkflowsAtMinute(k types.MinuteOfDay) []Workflow {
	return s.wf[k]
}
