// Package testprog provides a program to do scheduler testing
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	var (
		exitCode     int
		role         string
		sleepSeconds int
	)

	flag.IntVar(&sleepSeconds, "sleep", 0, "number of seconds to sleep. zero means no sleep but still run")
	flag.StringVar(&role, "role", "", "identifier describing how this program is used in the schedule (required)")
	flag.IntVar(&exitCode, "exitCode", -1, "exit code the program should return")
	flag.Parse()

	if sleepSeconds < 0 {
		fail("invalid --sleep value (must be >= 0)")
	}

	if role == "" {
		fail("missing --role value")
	}

	start := time.Now().UTC()

	log("START", role, fmt.Sprintf("sleep=%ds", sleepSeconds))
	time.Sleep(time.Duration(sleepSeconds) * time.Second)
	log("DONE", role, fmt.Sprintf("elapsed=%s", time.Since(start).Round(time.Millisecond)))

	if exitCode > -1 {
		os.Exit(exitCode)
	}

	os.Exit(0)
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
