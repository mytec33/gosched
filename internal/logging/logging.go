// Package logging provides base and injected logging
package logging

import (
	"crypto/rand"
	"log/slog"
	"os"
	"time"

	"github.com/oklog/ulid/v2"
)

var StdOut = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

type WorkflowLogger struct {
	WfRunID string
	Out     *slog.Logger
}

func NewWorkflowLogger(workflowName string) WorkflowLogger {
	now := time.Now()

	wfID := ulid.MustNew(ulid.Timestamp(now), rand.Reader).String()

	return WorkflowLogger{
		WfRunID: wfID,
		Out: StdOut.With(
			slog.String("wfRunID", wfID),
		),
	}
}
