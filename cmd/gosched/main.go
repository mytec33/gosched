package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"git.sr.ht/~mytec/gosched/internal/decode"
	"git.sr.ht/~mytec/gosched/internal/helpers"
	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/runner"
	"git.sr.ht/~mytec/gosched/internal/schedule"
)

const (
	ExitSuccess                int = 0
	ExitInvalidArgs            int = 1
	ExitWorkflowNotFoundByName int = 2
	ExitExecuteWorkflow        int = 3
	ExitScheduleError          int = 4
	ExitScheduleFailed         int = 5
	ExitRunOnceUnexpected      int = 7
	ExitInterruptSignal        int = 8
)

func main() {
	os.Exit(run())
}

func run() int {
	manifestFlag := ""
	newConfig := false
	runThisOnce := ""
	printSchedule := ""

	flag.StringVar(&manifestFlag, "manifest", "", "path to a manifest file containing schedule configuration files to load")
	flag.BoolVar(&newConfig, "new-config", false, "create an new configuration to begin with")
	flag.StringVar(&runThisOnce, "run-once", "", "run a workflow by name bypassing its scheduled time")
	flag.StringVar(&printSchedule, "print-schedule", "", "configuration summary: config, operational")
	flag.Parse()

	if newConfig {
		generateNewConfig()
		return ExitSuccess
	}

	if manifestFlag == "" {
		logging.StdOut.Error("startup", "reason", "missing required -manifest argument")
		flag.Usage()
		return ExitInvalidArgs
	}

	sched, errs := schedule.New(manifestFlag)
	if len(errs) > 0 {
		displayCfgErrors(errs)
		return ExitScheduleError
	}

	if printSchedule != "" {
		err := schedule.PrintConfiguration(printSchedule, sched)
		if err != nil {
			logging.StdOut.Error("startup", "reason", "unable to print", "error", err, "arg", printSchedule)
			return ExitInvalidArgs
		}
		return ExitSuccess
	}

	// This goes after newConfig or any other option that prints to STDOUT so only the output we
	// wish to print is shown and not logging messages. Those don't play well with JSON. :-)
	//
	// Also, a warn state to let whomever is responsible for this scheduling that the process has
	// started up. This is important from an admin point (uncontrolled shutdown, etc.)
	schedule.DisplayScheduleStats(sched)
	logging.StdOut.Warn("startup", "reason", "scheduler service started")

	if runThisOnce != "" {
		logging.StdOut.Info("startup", "reason", "run once started", "workflow", runThisOnce)

		var exitCode int

		err := runner.RunScheduleOnce(sched, runThisOnce)
		switch {
		case err == nil:
			exitCode = ExitSuccess
		case errors.Is(err, runner.ErrWorkflowNotFoundByName):
			exitCode = ExitWorkflowNotFoundByName
		case errors.Is(err, runner.ErrExecutingWorkflow):
			exitCode = ExitExecuteWorkflow
		default:
			exitCode = ExitRunOnceUnexpected
		}

		if err != nil {
			logging.StdOut.Info("run once stopped", "reason", err)
		}

		return exitCode
	} else {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		err := runner.RunSchedule(ctx, sched)
		switch {
		case err == nil:
			return ExitSuccess
		case errors.Is(err, runner.ErrSignalInterrupt):
			logging.StdOut.Error("scheduler interrupted", "reason", err)
			return ExitInterruptSignal
		default:
			logging.StdOut.Error("scheduler stopped", "reason", err)
			return ExitScheduleFailed
		}
	}
}

func displayCfgErrors(fileErrors []error) {
	logging.StdOut.Error("startup", "reason", "configuration invalid")

	for _, err := range fileErrors {
		fileErr, ok := errors.AsType[decode.FileValidationError](err)
		if ok {
			logging.StdOut.Error("startup", "file", fileErr.File,
				"reason", fileErr.Err)
		} else {
			logging.StdOut.Error("startup", "reason", err)
		}
	}
}

func generateNewConfig() {
	helpers.GenerateExampleConfig()
}
