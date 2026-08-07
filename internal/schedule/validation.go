package schedule

import (
	"fmt"
	"strings"
)

func (s Schedule) validate() []error {
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
