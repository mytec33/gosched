package schedule

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"git.sr.ht/~mytec/gosched/internal/errs"
	"git.sr.ht/~mytec/gosched/internal/helpers"
	"git.sr.ht/~mytec/gosched/internal/types"
)

const (
	MinutesPerDay int = 24 * 60
)

var (
	ErrTriggerIntervalInvalid = errors.New("trigger interval must be greater than zero")
	ErrWorkflowNameNotFound   = errors.New("workflow not found by name")
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

func expandCadence(trigger types.Trigger) ([]types.MinuteOfDay, error) {
	var interval int
	var minutes []types.MinuteOfDay

	// Minutes are the key focus of this scheduler. Minute is also the smallest
	// unit of time so we can calc against minutes in day as our upper boundary
	switch trigger.Every.Measure() {
	case types.CadenceDay:
		interval = trigger.Every.Repetition() * MinutesPerDay
	case types.CadenceHour:
		interval = trigger.Every.Repetition() * 60
	case types.CadenceMinute:
		interval = trigger.Every.Repetition()
	}

	if interval <= 0 {
		return minutes, ErrTriggerIntervalInvalid
	}

	for minute := int(*trigger.BeginAt); minute < MinutesPerDay; minute += interval {
		minutes = append(minutes, types.MinuteOfDay(minute))
	}

	return minutes, nil
}

func (s *Schedule) ExpandSchedule() error {
	var newByMinute = make(map[types.MinuteOfDay][]Workflow)

	for _, wf := range s.workflows {
		minutes, err := expandCadence(wf.Trigger)
		if err != nil {
			return fmt.Errorf("%s interval %s: %w", wf.Name, wf.Trigger.Every, err)
		}

		for _, m := range minutes {
			newByMinute[m] = append(newByMinute[m], wf)
		}
	}

	// Replace old schedule with new schedule to avoid comparing
	// new entries against previous entries. It is simpler to create
	// a new schedule than to patch an existing schedule.
	s.byMinute = newByMinute

	return nil
}

func (s Schedule) GetWorkflowByName(n string) (Workflow, error) {
	name := strings.ToLower(n)

	for _, v := range s.Workflows() {
		if strings.ToLower(v.Name) == name {
			return v, nil
		}
	}

	// Return workflow name as it appears in the config vs lower case version
	return Workflow{}, fmt.Errorf("%w: %s", ErrWorkflowNameNotFound, n)
}

func (s Schedule) PrintScheduleConfig(w io.Writer) {
	workflows := s.Workflows()
	wfWidth := len(strconv.Itoa(len(workflows)))

	for i, v := range workflows {
		fmt.Fprintf(w, "%*d: %s  %s (%s)\n", wfWidth, i+1, v.Trigger.String(), v.Name, v.OnFailure)

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
			fmt.Fprintf(w, "%*d: %s  %s (%s)\n", wfWidth, i+1, m.String(), v.Name, v.OnFailure)

			numSteps := len(strconv.Itoa(len(v.Steps)))
			for j, step := range v.Steps {
				printStep(numSteps, j, step, w)
			}
		}
	}
}

func printStep(numSteps int, stepIndex int, step Step, w io.Writer) {
	var details []string

	details = append(details, fmt.Sprintf("timeout %s", helpers.SecondsDuration(step.Timeout)))

	if step.Pause > 0 {
		details = append(details, fmt.Sprintf("pause %s", helpers.SecondsDuration(step.Pause)))
	}

	if len(details) > 0 {
		fmt.Fprintf(w, "\t\t%*d: %s (%s)\n", numSteps, stepIndex+1, step.Name, strings.Join(details, ", "))
	} else {
		fmt.Fprintf(w, "\t\t%*d: %s\n", numSteps, stepIndex+1, step.Name)
	}
}

func (s Schedule) Validate() []error {
	var errorList []error

	errorList = append(errorList, s.validateWorkflowCount()...)
	errorList = append(errorList, s.validateWorkflowNameDuplicates()...)

	return errorList
}

func (s Schedule) validateWorkflowCount() []error {
	var errorList []error

	count := s.WorkflowCount()
	if count > errs.MaxWorkflowCount {
		errorList = append(errorList, fmt.Errorf("%w: got %d", errs.ErrWorkflowCountExceeded, count))
	}

	return errorList
}

func (s Schedule) validateWorkflowNameDuplicates() []error {
	var errorList []error
	seen := make(map[string]string)

	for _, wf := range s.Workflows() {
		wfKey := strings.TrimSpace(strings.ToLower(wf.Name))

		key, ok := seen[wfKey]
		if ok {
			errorList = append(errorList, fmt.Errorf("%w: duplicate workflow name: '%v' duplicates '%v'",
				errs.ErrDuplicateWorkflowName, wf.Name, key))
		} else {
			seen[wfKey] = wf.Name
		}
	}

	return errorList
}

func (s Schedule) WorkflowCount() int {
	// workflows is the canonical list; byMinute is only an index.
	return len(s.workflows)
}

func (s Schedule) Workflows() []Workflow {
	out := make([]Workflow, len(s.workflows))
	copy(out, s.workflows)

	return out
}

func (s Schedule) WorkflowsAtMinute(k types.MinuteOfDay) []Workflow {
	return s.byMinute[k]
}
