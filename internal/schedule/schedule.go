package schedule

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"git.sr.ht/~mytec/gosched/internal/types"
)

type ScheduleSliceFlag []string

type Schedule struct {
	workflows []Workflow
	byMinute  map[types.MinuteOfDay][]Workflow
}

func (s *ScheduleSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *ScheduleSliceFlag) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s Schedule) PrintScheduleConfig(w io.Writer) {
	workflows := s.Workflows()
	wfWidth := len(strconv.Itoa(len(workflows)))

	for i, v := range workflows {
		fmt.Fprintf(w, "%*d: %s  %s (%s)\n", wfWidth, i+1, v.Time, v.Name, v.OnFailure)

		numSteps := len(strconv.Itoa(len(v.Steps)))
		for j, step := range v.Steps {
			printStep(numSteps, j, step, w)
		}
	}
}

func (s Schedule) PrintScheduleOperational(w io.Writer) {
	minutes := make([]types.MinuteOfDay, 0, len(s.byMinute))
	for m := range s.byMinute {
		minutes = append(minutes, m)
	}

	slices.Sort(minutes)

	for mi, m := range minutes {
		if mi > 0 {
			fmt.Fprintln(w)
		}

		workflows := s.byMinute[m]
		wfWidth := len(strconv.Itoa(len(workflows)))

		for i, v := range workflows {
			fmt.Fprintf(w, "%*d: %s  %s (%s)\n", wfWidth, i+1, m, v.Name, v.OnFailure)

			numSteps := len(strconv.Itoa(len(v.Steps)))
			for j, step := range v.Steps {
				printStep(numSteps, j, step, w)
			}
		}
	}
}

func printStep(numSteps int, stepIndex int, step Step, w io.Writer) {
	var details []string

	details = append(details, fmt.Sprintf("timeout %s", step.Timeout.Duration()))

	if step.Pause.Int() > 0 {
		details = append(details, fmt.Sprintf("pause %s", step.Pause.Duration()))
	}

	if len(details) > 0 {
		fmt.Fprintf(w, "\t\t%*d: %s (%s)\n", numSteps, stepIndex+1, step.Name, strings.Join(details, ", "))
	} else {
		fmt.Fprintf(w, "\t\t%*d: %s\n", numSteps, stepIndex+1, step.Name)
	}
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
