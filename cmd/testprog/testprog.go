// Package testprog provides a program to do scheduler testing
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	var (
		exitCode       int
		failOnceMarker string
		role           string
		sleepSeconds   int
	)

	flag.IntVar(&sleepSeconds, "sleep", 0, "number of seconds to sleep. zero means no sleep but still run")
	flag.StringVar(&role, "role", "", "identifier describing how this program is used in the schedule (required)")
	flag.IntVar(&exitCode, "exit-code", -1, "exit code the program should return")
	flag.StringVar(&failOnceMarker, "fail-once-marker", "", "marker file used to fail the first run and succeed thereafter")
	flag.Parse()

	if sleepSeconds < 0 {
		fail("invalid --sleep value (must be >= 0)")
	}

	if role == "" {
		fail("missing --role value")
	}

	if failOnceMarker != "" {
		if exitCode <= 0 {
			fail("--fail-once-marker requires --exit-code greater than zero")
		}

		firstRun, err := claimFirstRun(failOnceMarker)
		if err != nil {
			fail(fmt.Sprintf("unable to create fail-once marker: %v", err))
		}

		if !firstRun {
			exitCode = 0
		}
	}

	start := time.Now().UTC()

	log("START", role, fmt.Sprintf("sleep=%ds", sleepSeconds))
	time.Sleep(time.Duration(sleepSeconds) * time.Second)
	log("DONE", role, fmt.Sprintf("elapsed=%s", time.Since(start).Round(time.Millisecond)))

	if exitCode > -1 {
		log("EXIT", role, fmt.Sprintf("%v", exitCode))
		os.Exit(exitCode)
	}

	os.Exit(0)
}

func claimFirstRun(marker string) (bool, error) {
	file, err := os.OpenFile(marker, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if err := file.Close(); err != nil {
		return false, err
	}

	return true, nil
}

func log(state, role, msg string) {
	fmt.Printf(
		"%s | %-5s | role=%s | %s\n",
		time.Now().UTC().Format(time.RFC3339),
		state,
		role,
		msg,
	)
}

func fail(msg string) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
	flag.Usage()
	os.Exit(2)
}
