package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.sr.ht/~mytec/gosched/internal/decode"
	"git.sr.ht/~mytec/gosched/internal/helpers"
	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/manifest"
	"git.sr.ht/~mytec/gosched/internal/runner"
	"git.sr.ht/~mytec/gosched/internal/schedule"
)

const (
	ExitSuccess                int = 0
	ExitNoConfig               int = 1
	ExitValidation             int = 3
	ExitInvalidArgs            int = 4
	ExitManifestError          int = 8
	ExitWorkflowNotFoundByName int = 9
	ExitExecuteWorkflow        int = 10
	ExitScheduleValidation     int = 11
	ExitScheduleExpansion      int = 12
	ExitScheduleFailed         int = 13
	ExitRunOnceUnexpected      int = 14
	ExitInterruptSignal        int = 15
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

	workflowFiles, err := manifest.ParseManifest(manifestFlag)
	if err != nil {
		logging.StdOut.Error("startup", "reason", "no files found in manifest", "error", err)
		return ExitManifestError
	}

	sched, decodeErrors, err := decode.LoadSchedule(workflowFiles)
	if err != nil {
		logging.StdOut.Error("startup", "reason", "failed to load schedule", "error", err)
		return ExitNoConfig
	}

	// Show totals of parts used to assemble the schedule
	logging.StdOut.Info("startup", "fileCount", len(workflowFiles), "workflowCount",
		sched.WorkflowCount(), "stepCount", sched.StepCount())

	for _, v := range workflowFiles {
		logging.StdOut.Info("startup", "workflowFile", v)
	}

	if len(decodeErrors) > 0 {
		displayCfgErrors(decodeErrors)
		return ExitValidation
	}

	if printSchedule != "" {
		return printConfiguration(printSchedule, sched)
	}

	// This goes after newConfig or any other option that prints to STDOUT so only the output we
	// wish to print is shown and not logging messages. Those don't play well with JSON. :-)
	//
	// Also, a warn state to let whomever is responsible for this scheduling that the process has
	// started up. This is important from an admin point (uncontrolled shutdown, etc.)
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

func printConfiguration(method string, s schedule.Schedule) int {
	err := s.Print(method, os.Stdout)
	if err != nil {
		fmt.Println(err)
		return ExitInvalidArgs
	}

	return ExitSuccess
}

func displayCfgErrors(fileErrors []error) {
	logging.StdOut.Error("startup", "reason", "configuration invalid")

	for _, err := range fileErrors {
		var fileErr decode.FileValidationError

		if errors.As(err, &fileErr) {
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
