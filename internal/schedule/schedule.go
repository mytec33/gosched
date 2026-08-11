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
	MinWorkflowCount int = 1
	MaxWorkflowCount int = 64
)

var (
	ErrDuplicateWorkflowName  = errors.New("duplicate workflow name")
	ErrWorkflowCountExceeded  = fmt.Errorf("too many workflows in schedule: max is %d", MaxWorkflowCount)
	ErrTriggerIntervalInvalid = errors.New("trigger interval must be greater than zero")
	ErrWorkflowNameNotFound   = errors.New("workflow not found by name")
	ErrWorkflowsEmpty         = errors.New("a schedule requires at least 1 workflow")
)

type Schedule struct {
	workflows []workflow.Workflow
	byMinute  map[workflow.MinuteOfDay][]workflow.Workflow
}

func New(wfs []workflow.Workflow) (Schedule, []error) {
	var errorList []error

	s := Schedule{
		workflows: append([]workflow.Workflow(nil), wfs...),
	}

	errorList = s.validate()
	if len(errorList) > 0 {
		return Schedule{}, errorList
	}

	err := s.expandSchedule()
	if err != nil {
		return Schedule{}, []error{err}
	}

	return s, nil
}

func expandCadence(trigger workflow.Trigger) ([]workflow.MinuteOfDay, error) {
	var interval int
	var minutes []workflow.MinuteOfDay

	// Minutes are the key focus of this scheduler. Minute is also the smallest
	// unit of time so we can calc against minutes in day as our upper boundary
	switch trigger.Every.Measure() {
	case workflow.CadenceDay:
		interval = trigger.Every.Repetition() * workflow.MinutesInDay
	case workflow.CadenceHour:
		interval = trigger.Every.Repetition() * 60
	case workflow.CadenceMinute:
		interval = trigger.Every.Repetition()
	}

	if interval <= 0 {
		return minutes, ErrTriggerIntervalInvalid
	}

	for minute := int(trigger.BeginAt); minute < MinutesPerDay; minute += interval {
		minutes = append(minutes, workflow.MinuteOfDay(minute))
	}

	return minutes, nil
}

func (s *Schedule) expandSchedule() error {
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

func (s Schedule) EveryWorkflowHasName() bool {
	if len(s.workflows) == 0 {
		return false
	}

	for _, wf := range s.workflows {
		if wf.Name == "" {
			return false
		}
	}

	return true
}

func (s Schedule) EveryWorkflowHasSteps() bool {
	if len(s.workflows) == 0 {
		return false
	}

	for _, v := range s.workflows {
		if v.StepCount() == 0 {
			return false
		}
	}

	return true
}

func (s Schedule) StepCount() int {
	var count int

	for _, v := range s.workflows {
		count += v.StepCount()
	}

	return count
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
