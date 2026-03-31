package schedule

import "git.sr.ht/~mytec/gosched/internal/types"

type Schedule struct {
	workflows []Workflow
	byMinute  map[types.MinuteOfDay][]Workflow
}

func (s Schedule) WorkflowCount() int {
	count := 0
	for _, wfs := range s.byMinute {
		count += len(wfs)
	}
	return count
}

func (s Schedule) Workflows() []Workflow {
	out := make([]Workflow, len(s.workflows))
	copy(out, s.workflows)

	return out
}

func (s Schedule) WorkflowsAtMinute(k types.MinuteOfDay) []Workflow {
	return s.byMinute[k]
}
