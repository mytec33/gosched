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
		s.printScheduleConfig(w)
	case "operational":
		s.printScheduleOperational(w)
	default:
		return fmt.Errorf("unknown print config method: %s", method)
	}

	return nil
}

func PrintConfiguration(method string, s Schedule) error {
	err := s.Print(method, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}

func (s Schedule) printScheduleConfig(w io.Writer) {
	workflows := s.Workflows()
	wfWidth := len(strconv.Itoa(len(workflows)))

	currentSourceFile := ""
	for i, v := range workflows {
		if currentSourceFile != v.SourceFile {
			currentSourceFile = v.SourceFile
			fmt.Fprintf(w, "\n%s:\n", currentSourceFile)
		}
		fmt.Fprintf(w, "%*d: %s  %s (%s)%s\n",
			wfWidth, i+1, v.Trigger.String(), v.Name, v.OnFailure, disabledSuffix(v))

		numSteps := len(strconv.Itoa(len(v.Steps)))
		for j, step := range v.Steps {
			printStep(numSteps, j, step, w)
		}
	}

	fmt.Fprintf(w, "\n")
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

func (s Schedule) printScheduleOperational(w io.Writer) {
	minutes := make([]workflow.MinuteOfDay, 0, len(s.byMinute))
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
			fmt.Fprintf(w, "%*d: %s  %s (%s)%s\n",
				wfWidth, i+1, m.String(), v.Name, v.OnFailure, disabledSuffix(v))

			numSteps := len(strconv.Itoa(len(v.Steps)))
			for j, step := range v.Steps {
				printStep(numSteps, j, step, w)
			}
		}
	}
}

func printStep(numSteps int, stepIndex int, step workflow.Step, w io.Writer) {
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
		fmt.Fprintf(w, "\t\t%*d: %s (%s)\n", numSteps, stepIndex+1, step.Name, strings.Join(details, ", "))
	} else {
		fmt.Fprintf(w, "\t\t%*d: %s\n", numSteps, stepIndex+1, step.Name)
	}
}
