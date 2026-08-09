// Package decode reads workflow configuration input and converts it into
// trusted workflow values or validation errors.
package decode

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"git.sr.ht/~mytec/gosched/internal/manifest"
	"git.sr.ht/~mytec/gosched/internal/schedule"
	"git.sr.ht/~mytec/gosched/internal/workflow"
)

type FileValidationError struct {
	File string
	Err  error
}

func (e FileValidationError) Error() string {
	return fmt.Sprintf("%s: %v", e.File, e.Err)
}

func (e FileValidationError) Unwrap() error {
	return e.Err
}

var (
	ErrDecodeWorkflow = errors.New("decode workflow")
	ErrFileIOError    = errors.New("error opening file")
)

// LoadSchedule opens the schedule file and delegates decoding and validation.
// Validation errors are returned in the slice. The returned error is reserved for
// I/O or decoding failures.
func LoadSchedule(filename manifest.WorkflowFiles) (schedule.Schedule, []error, error) {
	var allValidationErrors []error
	var allWorkflows []workflow.Workflow

	for _, file := range filename {
		wfs, validationErrors, err := decodeWorkflowFile(file)
		if err != nil {
			return schedule.Schedule{}, validationErrors, err
		}

		if len(validationErrors) > 0 {
			for _, vErr := range validationErrors {
				allValidationErrors = append(allValidationErrors,
					FileValidationError{
						File: file,
						Err:  vErr,
					})
			}
			continue
		}

		allWorkflows = append(allWorkflows, wfs...)
	}
	if len(allValidationErrors) > 0 {
		return schedule.Schedule{}, allValidationErrors, nil
	}

	sched, errorList := schedule.New(allWorkflows)

	return sched, errorList, nil
}

func decodeWorkflowFile(file string) ([]workflow.Workflow, []error, error) {
	f, err := os.Open(file)
	if err != nil {
		return []workflow.Workflow{}, nil, fmt.Errorf("%w: %q: %w", ErrFileIOError, file, err)
	}
	defer f.Close()

	wfs, validationErrors, err := decodeWorkflows(f, file)
	if err != nil {
		return []workflow.Workflow{}, validationErrors, fmt.Errorf("decode workflows file %q: %w", file, err)
	}

	return wfs, validationErrors, nil
}

// decodeWorkflows reads JSON and performs validation.
// The returned slice contains validation errors found in the input.
// The returned error is reserved for I/O or decoding failures.
func decodeWorkflows(r io.Reader, sourceFile string) ([]workflow.Workflow, []error, error) {
	var workflows []workflow.Workflow
	var workflowsRaw []workflow.WorkflowRaw

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	err := dec.Decode(&workflowsRaw)
	if err != nil {
		return []workflow.Workflow{}, nil, fmt.Errorf("%w: %w", ErrDecodeWorkflow, err)
	}

	// Enforce exactly one top-level JSON value; allow only trailing whitespace.
	err = dec.Decode(&struct{}{})
	if err == nil {
		return []workflow.Workflow{}, nil, fmt.Errorf("%w: trailing data", ErrDecodeWorkflow)
	} else if !errors.Is(err, io.EOF) {
		return []workflow.Workflow{}, nil, fmt.Errorf("%w: trailing data: %w", ErrDecodeWorkflow, err)
	}

	// Loop through workflows to validate and bail if anything found
	var valErrs []error
	for _, wfRaw := range workflowsRaw {
		wf, valErrors := wfRaw.Validate(sourceFile)
		if len(valErrors) > 0 {
			valErrs = append(valErrs, valErrors...)
			continue
		}

		// Keep workflows as the trusted set; invalid raw workflows never cross
		// the decode boundary.
		workflows = append(workflows, wf)
	}

	if len(valErrs) > 0 {
		return []workflow.Workflow{}, valErrs, nil
	}

	return workflows, nil, nil
}
