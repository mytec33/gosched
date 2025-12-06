package logging

import (
	"log/slog"
	"os"
)

var StdoutLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// Logger for stderr - typically for errors and warnings
var StderrLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelWarn,
}))
