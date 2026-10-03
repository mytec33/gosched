package schedule

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"git.sr.ht/~mytec/gosched/internal/workflow"
)

func (s Schedule) Print(method string, w io.Writer) error {
	switch method {
	case "config":
		return s.printScheduleConfig(w)
	case "operational":
		return s.printScheduleOperational(w)
	default:
		return fmt.Errorf("unknown print config method: %s", method)
	}
}

func PrintConfiguration(method string, s Schedule) error {
	err := s.Print(method, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}

func (s Schedule) printScheduleConfig(w io.Writer) error {
	workflows := s.Workflows()
	wfWidth := len(strconv.Itoa(len(workflows)))

	currentSourceFile := ""
	for i, v := range workflows {
		if currentSourceFile != v.SourceFile {
			currentSourceFile = v.SourceFile
			if _, err := fmt.Fprintf(w, "\n%s:\n", currentSourceFile); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "%*d: %s  %s (%s)%s\n",
			wfWidth, i+1, v.Trigger.String(), v.Name, v.OnFailure, disabledSuffix(v)); err != nil {
			return err
		}

		numSteps := len(strconv.Itoa(len(v.Steps)))
		for j, step := range v.Steps {
			if err := printStep(numSteps, j, step, w); err != nil {
				return err
			}
		}
	}

	_, err := fmt.Fprintf(w, "\n")
	return err
}

func displayReason(s string) string {
	const max = 40
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func disabledSuffix(wf workflow.Workflow) string {
	if wf.Enabled {
		return ""
	}
	return fmt.Sprintf(" <--- disabled: %s", displayReason(wf.DisabledReason))
}

func (s Schedule) printScheduleOperational(w io.Writer) error {
	minutes := make([]workflow.MinuteOfDay, 0, len(s.byMinute))
	for m := range s.byMinute {
		minutes = append(minutes, m)
	}

	slices.Sort(minutes)

	for mi, m := range minutes {
		if mi > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}

		workflows := s.byMinute[m]
		wfWidth := len(strconv.Itoa(len(workflows)))

		for i, v := range workflows {
			if _, err := fmt.Fprintf(w, "%*d: %s  %s (%s)%s\n",
				wfWidth, i+1, m.String(), v.Name, v.OnFailure, disabledSuffix(v)); err != nil {
				return err
			}

			numSteps := len(strconv.Itoa(len(v.Steps)))
			for j, step := range v.Steps {
				if err := printStep(numSteps, j, step, w); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func printStep(numSteps int, stepIndex int, step workflow.Step, w io.Writer) error {
	var details []string

	if step.Timeout.Duration() > 0 {
		details = append(details, fmt.Sprintf("timeout %s", step.Timeout))
	} else {
		details = append(details, "no timeout")
	}

	if step.Pause.Duration() > 0 {
		details = append(details, fmt.Sprintf("pause %s", step.Pause))
	}

	if len(details) > 0 {
		_, err := fmt.Fprintf(w, "\t\t%*d: %s (%s)\n", numSteps, stepIndex+1, step.Name, strings.Join(details, ", "))
		return err
	} else {
		_, err := fmt.Fprintf(w, "\t\t%*d: %s\n", numSteps, stepIndex+1, step.Name)
		return err
	}
}
