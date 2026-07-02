// Package schedule arranges trusted workflows into an operational schedule
// that can be validated, expanded, queried, and printed.
package schedule

import (
	"errors"
	"fmt"
	"strings"

	"git.sr.ht/~mytec/gosched/internal/workflow"
)

const (
	MinutesPerDay int = 24 * 60
)

const (
	MaxWorkflowCount int = 64
)

var (
	ErrDuplicateWorkflowName  = errors.New("duplicate workflow name")
	ErrWorkflowCountExceeded  = fmt.Errorf("too many workflows in schedule: max is %d", MaxWorkflowCount)
	ErrTriggerIntervalInvalid = errors.New("trigger interval must be greater than zero")
	ErrWorkflowNameNotFound   = errors.New("workflow not found by name")
)

type ScheduleSliceFlag []string

type Schedule struct {
	workflows []workflow.Workflow
	byMinute  map[workflow.MinuteOfDay][]workflow.Workflow
}

func NewSchedule() Schedule {
	return Schedule{byMinute: make(map[workflow.MinuteOfDay][]workflow.Workflow)}
}

func FromWorkflows(w []workflow.Workflow) Schedule {
	s := NewSchedule()
	s.workflows = append(s.workflows, w...)

	return s
}

func (s *ScheduleSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *ScheduleSliceFlag) String() string {
	return fmt.Sprintf("%v", *s)
}

func expandCadence(trigger workflow.Trigger) ([]workflow.MinuteOfDay, error) {
	var interval int
	var minutes []workflow.MinuteOfDay

	// Minutes are the key focus of this scheduler. Minute is also the smallest
	// unit of time so we can calc against minutes in day as our upper boundary
	switch trigger.Every.Measure() {
	case workflow.CadenceDay:
		interval = trigger.Every.Repetition() * MinutesPerDay
	case workflow.CadenceHour:
		interval = trigger.Every.Repetition() * 60
	case workflow.CadenceMinute:
		interval = trigger.Every.Repetition()
	}

	if interval <= 0 {
		return minutes, ErrTriggerIntervalInvalid
	}

	for minute := int(*trigger.BeginAt); minute < MinutesPerDay; minute += interval {
		minutes = append(minutes, workflow.MinuteOfDay(minute))
	}

	return minutes, nil
}

func (s *Schedule) ExpandSchedule() error {
	var newByMinute = make(map[workflow.MinuteOfDay][]workflow.Workflow)

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

func (s Schedule) GetWorkflowByName(n string) (workflow.Workflow, error) {
	name := strings.ToLower(n)

	for _, v := range s.Workflows() {
		if strings.ToLower(v.Name) == name {
			return v, nil
		}
	}

	// Return workflow name as it appears in the config vs lower case version
	return workflow.Workflow{}, fmt.Errorf("%w: %s", ErrWorkflowNameNotFound, n)
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
	if count > MaxWorkflowCount {
		errorList = append(errorList, fmt.Errorf("%w: got %d", ErrWorkflowCountExceeded, count))
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
				ErrDuplicateWorkflowName, wf.Name, key))
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

func (s Schedule) Workflows() []workflow.Workflow {
	out := make([]workflow.Workflow, len(s.workflows))
	copy(out, s.workflows)

	return out
}

func (s Schedule) WorkflowsAtMinute(k workflow.MinuteOfDay) []workflow.Workflow {
	return s.byMinute[k]
}
