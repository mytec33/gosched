package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mytec33/gosched/internal/helpers"
)

func TestIntegrationAbortOnFirstStepFailure(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "args":["--role","abort-first","--exit-code","5"]},
		{"name":"Second step", "program":%q, "args":["--role","abort-second","--exit-code","0"]},
		{"name":"Third step", "program":%q, "args":["--role","abort-third","--exit-code","0"]}
	]`, testprog, testprog, testprog)
	out, _ := runAbortScenario(t, steps, 3)
	assertIntegrationEvents(t, out,
		integrationEvent{`START | role=abort-first |`, "the first step to start once"},
		integrationEvent{`EXIT  | role=abort-first | 5`, "the first step to fail with exit code 5"},
		integrationEvent{`stepIndex=0 status=failed exitCode=5`, "the first step's failure to be reported"},
		integrationEvent{`workflow="Abort scenario" status=failed duration=`, "the aborted workflow to report its final failed status and duration"})
	assertAbortStepsNotStarted(t, out, "abort-second", "abort-third")
}

func TestIntegrationAbortOnMiddleStepFailureHonorsEarlierPause(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "pause":"1s",
		 "args":["--role","abort-first","--exit-code","0"]},
		{"name":"Second step", "program":%q, "args":["--role","abort-second","--exit-code","5"]},
		{"name":"Third step", "program":%q, "args":["--role","abort-third","--exit-code","0"]}
	]`, testprog, testprog, testprog)
	out, elapsed := runAbortScenario(t, steps, 3)
	assertIntegrationEvents(t, out,
		integrationEvent{`START | role=abort-first |`, "the first step to start once"},
		integrationEvent{`EXIT  | role=abort-first | 0`, "the first step to finish successfully"},
		integrationEvent{`status=paused duration=1s`, "the successful first step's 1-second pause before the second step starts"},
		integrationEvent{`START | role=abort-second |`, "the second step to start after the first step's pause"},
		integrationEvent{`EXIT  | role=abort-second | 5`, "the second step to fail with exit code 5"},
		integrationEvent{`stepIndex=1 status=failed exitCode=5`, "the second step's failure to be reported"},
		integrationEvent{`workflow="Abort scenario" status=failed duration=`, "the aborted workflow to report its final failed status and duration"})
	assertAbortStepsNotStarted(t, out, "abort-third")
	if elapsed < time.Second {
		t.Errorf("abort policy must honor the successful first step's 1-second pause; execution took %s", elapsed)
	}
}

func TestIntegrationAbortOnFinalStepFailure(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "args":["--role","abort-first","--exit-code","0"]},
		{"name":"Second step", "program":%q, "args":["--role","abort-second","--exit-code","0"]},
		{"name":"Third step", "program":%q, "pause":"60s",
		 "args":["--role","abort-third","--exit-code","5"]}
	]`, testprog, testprog, testprog)
	out, _ := runAbortScenario(t, steps, 3)
	assertIntegrationEvents(t, out,
		integrationEvent{`START | role=abort-first |`, "the first step to start once"},
		integrationEvent{`EXIT  | role=abort-first | 0`, "the first step to finish successfully"},
		integrationEvent{`START | role=abort-second |`, "the second step to start after the first step succeeds"},
		integrationEvent{`EXIT  | role=abort-second | 0`, "the second step to finish successfully"},
		integrationEvent{`START | role=abort-third |`, "the final step to start after the second step succeeds"},
		integrationEvent{`EXIT  | role=abort-third | 5`, "the final step to fail with exit code 5"},
		integrationEvent{`stepIndex=2 status=failed exitCode=5`, "the final step's failure to be reported"},
		integrationEvent{`workflow="Abort scenario" status=failed duration=`, "a final-step failure to leave the workflow failed, despite earlier successes"})
	if bytes.Contains(out, []byte(`status=paused`)) {
		t.Fatalf("abort on the final step must not wait for its configured 60-second pause\n%s", out)
	}
}

func TestIntegrationAbortOnStepTimeout(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "timeout":"1s",
		 "args":["--role","abort-first","--sleep","60","--exit-code","0"]},
		{"name":"Second step", "program":%q, "args":["--role","abort-second","--exit-code","0"]},
		{"name":"Third step", "program":%q, "args":["--role","abort-third","--exit-code","0"]}
	]`, testprog, testprog, testprog)
	out, elapsed := runAbortScenario(t, steps, 3)
	assertIntegrationEvents(t, out,
		integrationEvent{`START | role=abort-first |`, "the first step to start before its timeout expires"},
		integrationEvent{`stepIndex=0 status=failed exitCode=-1`, "the timed-out first step to be reported as failed with exit code -1"},
		integrationEvent{`workflow="Abort scenario" status=failed duration=`, "the timed-out workflow to report its final failed status and duration"})
	assertIntegrationOutput(t, out, `context deadline exceeded`, "the step failure must identify the expired timeout")
	assertAbortStepsNotStarted(t, out, "abort-second", "abort-third")
	if bytes.Contains(out, []byte("DONE  | role=abort-first |")) || bytes.Contains(out, []byte("EXIT  | role=abort-first |")) {
		t.Fatalf("timed-out child must be terminated before finishing its 60-second sleep\n%s", out)
	}
	if elapsed < time.Second {
		t.Errorf("the first step must receive its configured 1-second timeout; execution took %s", elapsed)
	}
}

func TestIntegrationAbortOnProgramCannotStart(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	missingProgram := filepath.Join(t.TempDir(), "does-not-exist")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "args":["--role","abort-first"]},
		{"name":"Second step", "program":%q, "args":["--role","abort-second","--exit-code","0"]},
		{"name":"Third step", "program":%q, "args":["--role","abort-third","--exit-code","0"]}
	]`, missingProgram, testprog, testprog)
	out, _ := runAbortScenario(t, steps, 3)
	assertIntegrationEvents(t, out,
		integrationEvent{`stepIndex=0 status=failed exitCode=-1`, "the missing first program to be reported as a launch failure with exit code -1"},
		integrationEvent{`workflow="Abort scenario" status=failed duration=`, "the workflow to report its final failed status and duration after a launch failure"})
	assertIntegrationOutput(t, out, `does-not-exist`, "the launch failure diagnostics must identify the missing executable")
	assertAbortStepsNotStarted(t, out, "abort-first", "abort-second", "abort-third")
}

func TestIntegrationAbortAllStepsSucceed(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "pause":"1s",
		 "args":["--role","abort-first","--exit-code","0"]},
		{"name":"Second step", "program":%q, "pause":"2s",
		 "args":["--role","abort-second","--exit-code","0"]},
		{"name":"Third step", "program":%q, "pause":"60s",
		 "args":["--role","abort-third","--exit-code","0"]}
	]`, testprog, testprog, testprog)
	out, elapsed := runAbortScenario(t, steps, 0)
	assertIntegrationEvents(t, out,
		integrationEvent{`START | role=abort-first |`, "the first step to start once"},
		integrationEvent{`EXIT  | role=abort-first | 0`, "the first step to finish successfully"},
		integrationEvent{`status=paused duration=1s`, "the first step's 1-second pause before the second step starts"},
		integrationEvent{`START | role=abort-second |`, "the second step to start after the first step's pause"},
		integrationEvent{`EXIT  | role=abort-second | 0`, "the second step to finish successfully"},
		integrationEvent{`status=paused duration=2s`, "the second step's 2-second pause before the final step starts"},
		integrationEvent{`START | role=abort-third |`, "the final step to start after the second step's pause"},
		integrationEvent{`EXIT  | role=abort-third | 0`, "the final step to finish successfully"},
		integrationEvent{`status="skipped pause" reason="last step"`, "the final step's 60-second pause to be skipped"},
		integrationEvent{`workflow="Abort scenario" status=completed duration=`, "the successful workflow to report its final completed status and duration"})
	if elapsed < 3*time.Second {
		t.Errorf("successful steps must honor their 1-second and 2-second pauses; execution took %s", elapsed)
	}
	if bytes.Contains(out, []byte(`status=failed`)) || bytes.Contains(out, []byte(`status=abort`)) {
		t.Fatalf("abort policy must not report failure or abort when all steps succeed\n%s", out)
	}
}

func TestIntegrationAbortSkipsFailedStepPause(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "pause":"60s",
		 "args":["--role","abort-first","--exit-code","5"]},
		{"name":"Second step", "program":%q, "args":["--role","abort-second","--exit-code","0"]},
		{"name":"Third step", "program":%q, "args":["--role","abort-third","--exit-code","0"]}
	]`, testprog, testprog, testprog)
	// The command's 15-second deadline also rules out taking the 60-second pause.
	out, _ := runAbortScenario(t, steps, 3)
	assertIntegrationEvents(t, out,
		integrationEvent{`START | role=abort-first |`, "the first step to start once"},
		integrationEvent{`EXIT  | role=abort-first | 5`, "the first step to fail with exit code 5"},
		integrationEvent{`workflow="Abort scenario" status=failed duration=`, "the workflow to finish as failed without waiting for the failed step's pause"})
	assertAbortStepsNotStarted(t, out, "abort-second", "abort-third")
	if bytes.Contains(out, []byte(`status=paused`)) {
		t.Fatalf("abort must skip the failed step's 60-second pause because no next step will run\n%s", out)
	}
}

func runAbortScenario(t *testing.T, steps string, wantExitCode int) ([]byte, time.Duration) {
	t.Helper()
	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	config := fmt.Sprintf(`[{"name":"Abort scenario", "enabled":true,
		"trigger":{"every":"1d","beginAt":"00:00"}, "onFailure":"abort", "steps":%s}]`, steps)
	manifestFile := writeRunOnceManifest(t, config)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Abort scenario")
	cmd.WaitDelay = time.Second
	started := time.Now()
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(started)
	if ctx.Err() != nil {
		t.Fatalf("abort scenario must finish within 15 seconds; a timeout or skipped pause may not have been handled: %v\n%s", ctx.Err(), out)
	}
	if wantExitCode == 0 {
		if err != nil {
			t.Fatalf("all steps succeeded, so Gosched must exit successfully; got %v\n%s", err, out)
		}
	} else {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("a failed abort workflow must return Gosched exit code %d, not success or a launch error; got %v\n%s", wantExitCode, err, out)
		}
		if exitErr.ExitCode() != wantExitCode {
			t.Fatalf("a failed abort workflow must return Gosched exit code %d; got %d\n%s", wantExitCode, exitErr.ExitCode(), out)
		}
		assertIntegrationOutput(t, out, `run command aborts on error`, "the CLI must preserve the workflow abort error")
	}
	assertIntegrationOutput(t, out, `fileCount=1`, "the scenario must load one workflow file")
	assertIntegrationOutput(t, out, `workflowCount=1`, "the scenario must construct one workflow")
	assertIntegrationOutput(t, out, `stepCount=3`, "the scenario must construct all three steps")
	if bytes.Contains(out, []byte(`retryAttempt=`)) || bytes.Contains(out, []byte(`status=exhausted`)) {
		t.Fatalf("abort policy must not retry failed steps\n%s", out)
	}

	status := "failed"
	if wantExitCode == 0 {
		status = "completed"
	}
	assertNoUnexpectedOutcome(t, out)
	assertWorkflowSummary(t, out, "Abort scenario", status)
	return out, elapsed
}

func assertAbortStepsNotStarted(t *testing.T, out []byte, roles ...string) {
	t.Helper()
	for _, role := range roles {
		if bytes.Contains(out, []byte("START | role="+role+" |")) {
			t.Fatalf("abort must prevent step %q from starting after the failure\n%s", role, out)
		}
	}
}
