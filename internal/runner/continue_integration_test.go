package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mytec33/gosched/internal/helpers"
)

func TestIntegrationContinueAfterFirstStepFails(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "args":["--role","continue-first","--exit-code","5"]},
		{"name":"Second step", "program":%q, "args":["--role","continue-second","--exit-code","0"]},
		{"name":"Third step", "program":%q, "args":["--role","continue-third","--exit-code","0"]}
	]`, testprog, testprog, testprog)
	out, _ := runContinueScenario(t, steps)
	assertIntegrationEvents(t, out,
		integrationEvent{`START | role=continue-first |`, "the first step to start"},
		integrationEvent{`EXIT  | role=continue-first | 5`, "the first step to exit with the configured failure code 5"},
		integrationEvent{`START | role=continue-second |`, "the second step to start after the first step finishes"},
		integrationEvent{`EXIT  | role=continue-second | 0`, "the second step to finish successfully"},
		integrationEvent{`START | role=continue-third |`, "the third step to start after the second step finishes"},
		integrationEvent{`EXIT  | role=continue-third | 0`, "the final step to finish successfully"})
	assertIntegrationOutput(t, out, `stepIndex=0 status=failed exitCode=5`, "the first step's failure must be reported with exit code 5")
	assertWorkflowSummary(t, out, "Continue scenario", "partial")
}

func TestIntegrationContinueAfterMultipleFailuresIncludingLastStep(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "args":["--role","continue-first","--exit-code","5"]},
		{"name":"Second step", "program":%q, "args":["--role","continue-second","--exit-code","7"]},
		{"name":"Third step", "program":%q, "args":["--role","continue-third","--exit-code","9"]}
	]`, testprog, testprog, testprog)
	out, _ := runContinueScenario(t, steps)
	assertIntegrationEvents(t, out,
		integrationEvent{`START | role=continue-first |`, "the first step to start"},
		integrationEvent{`EXIT  | role=continue-first | 5`, "the first step to exit with the configured failure code 5"},
		integrationEvent{`START | role=continue-second |`, "the second step to start after the first step finishes"},
		integrationEvent{`EXIT  | role=continue-second | 7`, "the second step to exit with the configured failure code 7"},
		integrationEvent{`START | role=continue-third |`, "the third step to start after the second step finishes"},
		integrationEvent{`EXIT  | role=continue-third | 9`, "the final step to exit with the configured failure code 9"})
	assertIntegrationOutput(t, out, `stepIndex=0 status=failed exitCode=5`, "the first step's failure must be reported with exit code 5")
	assertIntegrationOutput(t, out, `stepIndex=1 status=failed exitCode=7`, "the second step's failure must be reported with exit code 7")
	assertIntegrationOutput(t, out, `stepIndex=2 status=failed exitCode=9`, "the final step's failure must be reported with exit code 9")
	assertWorkflowSummary(t, out, "Continue scenario", "partial")
}

func TestIntegrationContinueAllStepsSucceed(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "args":["--role","continue-first","--exit-code","0"]},
		{"name":"Second step", "program":%q, "args":["--role","continue-second","--exit-code","0"]},
		{"name":"Third step", "program":%q, "args":["--role","continue-third","--exit-code","0"]}
	]`, testprog, testprog, testprog)
	out, _ := runContinueScenario(t, steps)
	assertIntegrationEvents(t, out,
		integrationEvent{`START | role=continue-first |`, "the first step to start"},
		integrationEvent{`EXIT  | role=continue-first | 0`, "the first step to finish successfully"},
		integrationEvent{`START | role=continue-second |`, "the second step to start after the first step finishes"},
		integrationEvent{`EXIT  | role=continue-second | 0`, "the second step to finish successfully"},
		integrationEvent{`START | role=continue-third |`, "the third step to start after the second step finishes"},
		integrationEvent{`EXIT  | role=continue-third | 0`, "the final step to finish successfully"})
	if bytes.Contains(out, []byte(`status=failed`)) {
		t.Fatalf("successful steps must not report failure\n%s", out)
	}
	assertWorkflowSummary(t, out, "Continue scenario", "completed")
}

func TestIntegrationContinueAfterStepTimeout(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "timeout":"1s",
		 "args":["--role","continue-first","--sleep","60","--exit-code","0"]},
		{"name":"Second step", "program":%q, "args":["--role","continue-second","--exit-code","0"]},
		{"name":"Third step", "program":%q, "args":["--role","continue-third","--exit-code","0"]}
	]`, testprog, testprog, testprog)
	out, elapsed := runContinueScenario(t, steps)
	assertIntegrationEvents(t, out,
		integrationEvent{`START | role=continue-first |`, "the first step to start"},
		integrationEvent{`stepIndex=0 status=failed exitCode=-1`, "the first step to be reported as failed with exit code -1 before continuing"},
		integrationEvent{`START | role=continue-second |`, "the second step to start after the first step finishes"},
		integrationEvent{`EXIT  | role=continue-second | 0`, "the second step to finish successfully"},
		integrationEvent{`START | role=continue-third |`, "the third step to start after the second step finishes"},
		integrationEvent{`EXIT  | role=continue-third | 0`, "the final step to finish successfully"})
	assertIntegrationOutput(t, out, `context deadline exceeded`, "the first step's failure must identify its expired timeout")
	if bytes.Contains(out, []byte("DONE  | role=continue-first |")) || bytes.Contains(out, []byte("EXIT  | role=continue-first |")) {
		t.Fatalf("timed-out child must not finish its 60-second sleep\n%s", out)
	}
	if elapsed < time.Second {
		t.Errorf("step timed out before its configured 1-second deadline: %s", elapsed)
	}
	assertWorkflowSummary(t, out, "Continue scenario", "partial")
}

func TestIntegrationContinueAfterProgramCannotStart(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	missingProgram := filepath.Join(t.TempDir(), "does-not-exist")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "args":["--role","continue-first"]},
		{"name":"Second step", "program":%q, "args":["--role","continue-second","--exit-code","0"]},
		{"name":"Third step", "program":%q, "args":["--role","continue-third","--exit-code","0"]}
	]`, missingProgram, testprog, testprog)
	out, _ := runContinueScenario(t, steps)
	assertIntegrationEvents(t, out,
		integrationEvent{`stepIndex=0 status=failed exitCode=-1`, "the first step to be reported as failed with exit code -1 before continuing"},
		integrationEvent{`START | role=continue-second |`, "the second step to start after the first step finishes"},
		integrationEvent{`EXIT  | role=continue-second | 0`, "the second step to finish successfully"},
		integrationEvent{`START | role=continue-third |`, "the third step to start after the second step finishes"},
		integrationEvent{`EXIT  | role=continue-third | 0`, "the final step to finish successfully"})
	if bytes.Contains(out, []byte("START | role=continue-first |")) {
		t.Fatalf("missing program must not start\n%s", out)
	}
	assertIntegrationOutput(t, out, `reason=`, "the program launch failure must include a diagnostic reason")
	assertIntegrationOutput(t, out, `does-not-exist`, "the launch failure diagnostics must identify the missing executable")
	assertWorkflowSummary(t, out, "Continue scenario", "partial")
}

func TestIntegrationContinueHonorsPausesExceptAfterLastStep(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	steps := fmt.Sprintf(`[
		{"name":"First step", "program":%q, "pause":"2s",
		 "args":["--role","continue-first","--exit-code","5"]},
		{"name":"Second step", "program":%q, "pause":"1s",
		 "args":["--role","continue-second","--exit-code","0"]},
		{"name":"Third step", "program":%q, "pause":"60s",
		 "args":["--role","continue-third","--exit-code","9"]}
	]`, testprog, testprog, testprog)
	out, elapsed := runContinueScenario(t, steps)
	assertIntegrationEvents(t, out,
		integrationEvent{`START | role=continue-first |`, "the first step to start"},
		integrationEvent{`EXIT  | role=continue-first | 5`, "the first step to exit with the configured failure code 5"},
		integrationEvent{`status=paused duration=2s`, "a 2-second pause after the first step fails, before the second step starts"},
		integrationEvent{`START | role=continue-second |`, "the second step to start after the first step finishes"},
		integrationEvent{`EXIT  | role=continue-second | 0`, "the second step to finish successfully"},
		integrationEvent{`status=paused duration=1s`, "a 1-second pause after the second step succeeds, before the final step starts"},
		integrationEvent{`START | role=continue-third |`, "the third step to start after the second step finishes"},
		integrationEvent{`EXIT  | role=continue-third | 9`, "the final step to exit with the configured failure code 9"},
		integrationEvent{`status="skipped pause" reason="last step"`, "the final step's 60-second pause to be skipped because no steps remain"})
	pauses := bytes.Count(out, []byte(`status=paused`))
	if pauses != 2 {
		t.Fatalf("expected only the two between-step pauses, got %d\n%s", pauses, out)
	}
	if elapsed < 3*time.Second {
		t.Errorf("expected 2-second and 1-second pauses, execution took %s", elapsed)
	}
	assertWorkflowSummary(t, out, "Continue scenario", "partial")
}

func runContinueScenario(t *testing.T, steps string) ([]byte, time.Duration) {
	t.Helper()
	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	config := fmt.Sprintf(`[{"name":"Continue scenario", "enabled":true,
		"trigger":{"every":"1d","beginAt":"00:00"}, "onFailure":"continue", "steps":%s}]`, steps)
	manifestFile := writeRunOnceManifest(t, config)
	// Long sleeps and the final 60-second pause must not hold up completion.
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Continue scenario")
	cmd.WaitDelay = time.Second
	started := time.Now()
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("continue workflow must finish within 15 seconds and return success even when steps fail; command returned %v (context: %v)\n%s", err, ctx.Err(), out)
	}
	assertIntegrationOutput(t, out, `fileCount=1`, "the scenario must load exactly one workflow file")
	assertIntegrationOutput(t, out, `workflowCount=1`, "the scenario must construct exactly one workflow")
	assertIntegrationOutput(t, out, `stepCount=3`, "the scenario must construct all three configured steps")
	if bytes.Contains(out, []byte(`retryAttempt=`)) || bytes.Contains(out, []byte(`status=exhausted`)) {
		t.Fatalf("continue must not retry failed steps\n%s", out)
	}
	assertNoUnexpectedOutcome(t, out)
	return out, elapsed
}

func assertIntegrationEvents(t *testing.T, out []byte, events ...integrationEvent) {
	t.Helper()
	remaining := out
	for _, event := range events {
		count := bytes.Count(out, []byte(event.output))
		if count == 0 {
			t.Fatalf("expected %s, but it was not reported (log detail: %q)\n%s", event.description, event.output, out)
		}
		if count != 1 {
			t.Fatalf("expected %s exactly once; found %d matching events (log detail: %q)\n%s", event.description, count, event.output, out)
		}
		position := bytes.Index(remaining, []byte(event.output))
		if position < 0 {
			t.Fatalf("expected %s at this point in the workflow, but it was reported earlier and out of order (log detail: %q)\n%s", event.description, event.output, out)
		}
		remaining = remaining[position+len(event.output):]
	}
}
