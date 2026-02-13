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

// Logger for stderr - typically for errors and warnings
var StdErr = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelWarn,
}))

type WorkFlowLogger struct {
	WfRunID string
	Out     *slog.Logger
	Err     *slog.Logger
}

func NewWorkFlowLogger(workflowName string) WorkFlowLogger {
	now := time.Now()

	wfID := ulid.MustNew(ulid.Timestamp(now), rand.Reader).String()

	return WorkFlowLogger{
		WfRunID: wfID,
		Out: StdOut.With(
			slog.String("wfRunID", wfID),
			slog.String("workflow", workflowName),
		),
		Err: StdErr.With(
			slog.String("wfRunID", wfID),
			slog.String("workflow", workflowName),
		),
	}
}
