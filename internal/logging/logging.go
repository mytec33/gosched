// Package logging provides base and injected logging
package logging

import (
	"log/slog"
	"os"
	"uuid"
)

var StdOut = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

type WorkflowLogger struct {
	WfRunID string
	Out     *slog.Logger
}

func NewWorkflowLogger(workflowName string) WorkflowLogger {
	wfID := uuid.NewV7().String()

	return WorkflowLogger{
		WfRunID: wfID,
		Out: StdOut.With(
			slog.String("wfRunID", wfID),
			slog.String("workflow", workflowName),
		),
	}
}
