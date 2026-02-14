package schedule

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
