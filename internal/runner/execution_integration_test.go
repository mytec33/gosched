package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mytec33/gosched/internal/helpers"
)

func TestIntegrationRetryExhaustionRunsNextStep(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}

	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	config := fmt.Sprintf(`[
		{
			"name": "Retry exhaustion",
			"enabled": true,
			"trigger": {"every": "1d", "beginAt": "00:00"},
			"onFailure": "retry",
			"retry": {"numberRetries": 2, "pauseSeconds": 1},
			"steps": [
				{"name": "Always fails", "program": %q,
				 "args": ["--role", "retry-failure", "--exit-code", "5"]},
				{"name": "Runs after exhaustion", "program": %q,
				 "args": ["--role", "after-retries", "--exit-code", "0"]}
			]
		}
	]`, testprog, testprog)
	manifestFile := writeRunOnceManifest(t, config)

	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Retry exhaustion")
	cmd.WaitDelay = time.Second
	started := time.Now()
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("retry workflow must finish within 15 seconds and return success; command returned %v (context: %v)\n%s", err, ctx.Err(), out)
	}

	assertIntegrationOutput(t, out, `fileCount=1`, "the scenario must load exactly one workflow file")
	assertIntegrationOutput(t, out, `workflowCount=1`, "the scenario must construct exactly one workflow")
	assertIntegrationOutput(t, out, `stepCount=2`, "the scenario must construct both configured steps")
	assertIntegrationOutput(t, out, `status=exhausted`, "the failed step must report that its allowed retries were exhausted")
	assertRetryRecoveryReports(t, out, 0)
	assertNoUnexpectedOutcome(t, out)
	assertWorkflowSummary(t, out, "Retry exhaustion", "partial")

	attempts := bytes.Count(out, []byte("START | role=retry-failure |"))
	if attempts != 3 {
		t.Fatalf("the failing step must run exactly 3 times (initial attempt plus 2 retries); observed %d starts\n%s", attempts, out)
	}

	nextStepRuns := bytes.Count(out, []byte("START | role=after-retries |"))
	if nextStepRuns != 1 {
		t.Fatalf("expected next step to run once, got %d runs\n%s", nextStepRuns, out)
	}

	nextStepSuccesses := bytes.Count(out, []byte("EXIT  | role=after-retries | 0"))
	if nextStepSuccesses != 1 {
		t.Fatalf("expected next step to succeed once, got %d successes\n%s", nextStepSuccesses, out)
	}

	lastFailure := bytes.LastIndex(out, []byte("EXIT  | role=retry-failure | 5"))
	nextStep := bytes.Index(out, []byte("START | role=after-retries |"))
	if lastFailure < 0 || nextStep <= lastFailure {
		t.Fatalf("after retry exhaustion, the following step must start after the failing step exits with code 5; that exit is missing or the following step started too early\n%s", out)
	}

	if elapsed < 2*time.Second {
		t.Errorf("expected two 1-second retry pauses, execution took %s", elapsed)
	}
}

func TestIntegrationZeroRetriesRunsNextStepWithoutRetryPause(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}

	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	config := fmt.Sprintf(`[
		{
			"name": "Zero retries",
			"enabled": true,
			"trigger": {"every": "1d", "beginAt": "00:00"},
			"onFailure": "retry",
			"retry": {"numberRetries": 0, "pauseSeconds": 60},
			"steps": [
				{"name": "Fails once without retrying", "program": %q,
				 "args": ["--role", "zero-retry-failure", "--exit-code", "5"]},
				{"name": "Runs after initial failure", "program": %q,
				 "args": ["--role", "after-zero-retries", "--exit-code", "0"]}
			]
		}
	]`, testprog, testprog)
	manifestFile := writeRunOnceManifest(t, config)

	// Completing before the configured 60-second pause proves it was bypassed.
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Zero retries")
	cmd.WaitDelay = time.Second

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("zero retries must bypass the configured 60-second retry pause and finish within 15 seconds; command returned %v (context: %v)\n%s", err, ctx.Err(), out)
	}

	attempts := bytes.Count(out, []byte("START | role=zero-retry-failure |"))
	if attempts != 1 {
		t.Fatalf("zero retries must run the failing step once with no additional attempts; observed %d starts\n%s", attempts, out)
	}

	failure := bytes.Index(out, []byte("EXIT  | role=zero-retry-failure | 5"))
	nextStep := bytes.Index(out, []byte("START | role=after-zero-retries |"))
	if failure < 0 || nextStep <= failure {
		t.Fatalf("with zero retries, the following step must start after the initial attempt exits with code 5; that exit is missing or the following step started too early\n%s", out)
	}

	nextStepRuns := bytes.Count(out, []byte("START | role=after-zero-retries |"))
	if nextStepRuns != 1 {
		t.Fatalf("expected next step to run once, got %d runs\n%s", nextStepRuns, out)
	}

	nextStepSuccesses := bytes.Count(out, []byte("EXIT  | role=after-zero-retries | 0"))
	if nextStepSuccesses != 1 {
		t.Fatalf("expected next step to succeed once, got %d successes\n%s", nextStepSuccesses, out)
	}

	if bytes.Contains(out, []byte(`retryAttempt=`)) || bytes.Contains(out, []byte(`status=paused`)) {
		t.Fatalf("zero retries must not attempt or pause for a retry\n%s", out)
	}
	assertRetryRecoveryReports(t, out, 0)
	assertNoUnexpectedOutcome(t, out)
	assertWorkflowSummary(t, out, "Zero retries", "partial")
}

func TestIntegrationRetrySuccessWithAllStepsRetrying(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}

	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	dir := t.TempDir()
	firstMarker := filepath.Join(dir, "first-step")
	secondMarker := filepath.Join(dir, "second-step")
	thirdMarker := filepath.Join(dir, "third-step")
	config := fmt.Sprintf(`[
		{
			"name": "Retry success",
			"enabled": true,
			"trigger": {"every": "1d", "beginAt": "00:00"},
			"onFailure": "retry",
			"retry": {"numberRetries": 2, "pauseSeconds": 1},
			"steps": [
				{"name": "First step", "program": %q,
				 "args": ["--role", "retry-first", "--exit-code", "5", "--fail-once-marker", %q]},
				{"name": "Second step", "program": %q,
				 "args": ["--role", "retry-second", "--exit-code", "5", "--fail-once-marker", %q]},
				{"name": "Third step", "program": %q,
				 "args": ["--role", "retry-third", "--exit-code", "5", "--fail-once-marker", %q]}
			]
		}
	]`, testprog, firstMarker, testprog, secondMarker, testprog, thirdMarker)
	manifestFile := writeRunOnceManifest(t, config)

	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Retry success")
	cmd.WaitDelay = time.Second
	started := time.Now()
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("retry workflow must finish within 15 seconds and return success; command returned %v (context: %v)\n%s", err, ctx.Err(), out)
	}

	for _, step := range []struct {
		role     string
		name     string
		attempts int
	}{
		{"retry-first", "first step", 2},
		{"retry-second", "second step", 2},
		{"retry-third", "third step", 2},
	} {
		attempts := bytes.Count(out, []byte("START | role="+step.role+" |"))
		if attempts != step.attempts {
			t.Fatalf("%s must run %d time(s), including any retry, and stop once successful; observed %d starts\n%s", step.name, step.attempts, attempts, out)
		}
	}

	// Each failure must be retried successfully before the following step starts.
	events := []integrationEvent{
		{`START | role=retry-first |`, "the first step to start its initial attempt"},
		{`EXIT  | role=retry-first | 5`, "the first step to fail initially with exit code 5"},
		{`START | role=retry-first |`, "the first step to start its retry after the initial failure"},
		{`EXIT  | role=retry-first | 0`, "the first step to recover successfully on retry before proceeding"},
		{`stepName="First step" stepIndex=0 status=completed retryAttempt=1`, "the first step to report recovery on retry 1 before execution proceeds"},
		{`START | role=retry-second |`, "the second step to start its initial attempt"},
		{`EXIT  | role=retry-second | 5`, "the second step to fail initially with exit code 5"},
		{`START | role=retry-second |`, "the second step to start its retry after the initial failure"},
		{`EXIT  | role=retry-second | 0`, "the second step to recover successfully on retry before proceeding"},
		{`stepName="Second step" stepIndex=1 status=completed retryAttempt=1`, "the second step to report recovery on retry 1 before execution proceeds"},
		{`START | role=retry-third |`, "the third step to start its initial attempt"},
		{`EXIT  | role=retry-third | 5`, "the third step to fail initially with exit code 5"},
		{`START | role=retry-third |`, "the third step to start its retry after the initial failure"},
		{`EXIT  | role=retry-third | 0`, "the third step to recover successfully on retry before proceeding"},
		{`stepName="Third step" stepIndex=2 status=completed retryAttempt=1`, "the third step to report recovery on retry 1 before execution proceeds"},
	}
	remaining := out
	for _, event := range events {
		position := bytes.Index(remaining, []byte(event.output))
		if position < 0 {
			t.Fatalf("expected %s at this point in the workflow; its event is missing or out of order (log detail: %q)\n%s", event.description, event.output, out)
		}
		remaining = remaining[position+len(event.output):]
	}
	if elapsed < 3*time.Second {
		t.Errorf("expected three 1-second retry pauses, execution took %s", elapsed)
	}
	assertRetryRecoveryReports(t, out, 3)
	assertIntegrationOutput(t, out, `stepCount=3`, "the scenario must construct all three configured steps")
	assertNoUnexpectedOutcome(t, out)
	assertWorkflowSummary(t, out, "Retry success", "completed")
	if bytes.Contains(out, []byte(`status=partial`)) || bytes.Contains(out, []byte(`status=exhausted`)) {
		t.Fatalf("successful retries must not leave the workflow partial or exhausted\n%s", out)
	}
}

func TestIntegrationRetrySuccessWithFinalStepSucceedingImmediately(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}

	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	dir := t.TempDir()
	firstMarker := filepath.Join(dir, "first-step")
	secondMarker := filepath.Join(dir, "second-step")
	config := fmt.Sprintf(`[
		{
			"name": "Retry success",
			"enabled": true,
			"trigger": {"every": "1d", "beginAt": "00:00"},
			"onFailure": "retry",
			"retry": {"numberRetries": 2, "pauseSeconds": 1},
			"steps": [
				{"name": "First step", "program": %q,
				 "args": ["--role", "retry-first", "--exit-code", "5", "--fail-once-marker", %q]},
				{"name": "Second step", "program": %q,
				 "args": ["--role", "retry-second", "--exit-code", "5", "--fail-once-marker", %q]},
				{"name": "Third step", "program": %q,
				 "args": ["--role", "retry-third", "--exit-code", "0"]}
			]
		}
	]`, testprog, firstMarker, testprog, secondMarker, testprog)
	manifestFile := writeRunOnceManifest(t, config)

	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Retry success")
	cmd.WaitDelay = time.Second
	started := time.Now()
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("retry workflow must finish within 15 seconds and return success; command returned %v (context: %v)\n%s", err, ctx.Err(), out)
	}

	for _, step := range []struct {
		role     string
		name     string
		attempts int
	}{
		{"retry-first", "first step", 2},
		{"retry-second", "second step", 2},
		{"retry-third", "third step", 1},
	} {
		attempts := bytes.Count(out, []byte("START | role="+step.role+" |"))
		if attempts != step.attempts {
			t.Fatalf("%s must run %d time(s), including any retry, and stop once successful; observed %d starts\n%s", step.name, step.attempts, attempts, out)
		}
	}

	// Each failure must be retried successfully before the following step starts.
	events := []integrationEvent{
		{`START | role=retry-first |`, "the first step to start its initial attempt"},
		{`EXIT  | role=retry-first | 5`, "the first step to fail initially with exit code 5"},
		{`START | role=retry-first |`, "the first step to start its retry after the initial failure"},
		{`EXIT  | role=retry-first | 0`, "the first step to recover successfully on retry before proceeding"},
		{`stepName="First step" stepIndex=0 status=completed retryAttempt=1`, "the first step to report recovery on retry 1 before execution proceeds"},
		{`START | role=retry-second |`, "the second step to start its initial attempt"},
		{`EXIT  | role=retry-second | 5`, "the second step to fail initially with exit code 5"},
		{`START | role=retry-second |`, "the second step to start its retry after the initial failure"},
		{`EXIT  | role=retry-second | 0`, "the second step to recover successfully on retry before proceeding"},
		{`stepName="Second step" stepIndex=1 status=completed retryAttempt=1`, "the second step to report recovery on retry 1 before execution proceeds"},
		{`START | role=retry-third |`, "the third step to start its initial attempt"},
		{`EXIT  | role=retry-third | 0`, "the third step to succeed without needing a retry"},
	}
	remaining := out
	for _, event := range events {
		position := bytes.Index(remaining, []byte(event.output))
		if position < 0 {
			t.Fatalf("expected %s at this point in the workflow; its event is missing or out of order (log detail: %q)\n%s", event.description, event.output, out)
		}
		remaining = remaining[position+len(event.output):]
	}
	if elapsed < 2*time.Second {
		t.Errorf("expected two 1-second retry pauses, execution took %s", elapsed)
	}
	assertRetryRecoveryReports(t, out, 2)
	assertIntegrationOutput(t, out, `stepCount=3`, "the scenario must construct all three configured steps")
	assertNoUnexpectedOutcome(t, out)
	assertWorkflowSummary(t, out, "Retry success", "completed")
	if bytes.Contains(out, []byte(`status=partial`)) || bytes.Contains(out, []byte(`status=exhausted`)) {
		t.Fatalf("successful retries must not leave the workflow partial or exhausted\n%s", out)
	}
}

func TestIntegrationRetryRecoveryPreservesEarlierPartialOutcome(t *testing.T) {
	if os.Getenv("GOSCHED_INTEGRATION") != "1" {
		t.Skip("set GOSCHED_INTEGRATION=1 to run execution scenarios")
	}

	scheduler := helpers.BuildBinary(t, "gosched", "cmd/gosched")
	testprog := helpers.BuildBinary(t, "testprog", "cmd/testprog")
	marker := filepath.Join(t.TempDir(), "second-step")
	config := fmt.Sprintf(`[
		{
			"name": "Retry mixed outcome",
			"enabled": true,
			"trigger": {"every": "1d", "beginAt": "00:00"},
			"onFailure": "retry",
			"retry": {"numberRetries": 2, "pauseSeconds": 1},
			"steps": [
				{"name": "First step", "program": %q,
				 "args": ["--role", "mixed-first", "--exit-code", "5"]},
				{"name": "Second step", "program": %q,
				 "args": ["--role", "mixed-second", "--exit-code", "5", "--fail-once-marker", %q]},
				{"name": "Third step", "program": %q,
				 "args": ["--role", "mixed-third", "--exit-code", "0"]}
			]
		}
	]`, testprog, testprog, marker, testprog)
	manifestFile := writeRunOnceManifest(t, config)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, scheduler, "-manifest", manifestFile, "-run-once", "Retry mixed outcome")
	cmd.WaitDelay = time.Second
	started := time.Now()
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("retry workflow must finish later steps after an exhausted step; command returned %v (context: %v)\n%s", err, ctx.Err(), out)
	}

	for _, step := range []struct {
		role     string
		name     string
		attempts int
	}{
		{"mixed-first", "first step that exhausts retries", 3},
		{"mixed-second", "second step that recovers on retry", 2},
		{"mixed-third", "third step that succeeds initially", 1},
	} {
		attempts := bytes.Count(out, []byte("START | role="+step.role+" |"))
		if attempts != step.attempts {
			t.Fatalf("%s must run exactly %d times; observed %d starts\n%s", step.name, step.attempts, attempts, out)
		}
	}
	failures := bytes.Count(out, []byte("EXIT  | role=mixed-first | 5"))
	if failures != 3 {
		t.Fatalf("the first step must fail its initial attempt and both retries; observed %d failures\n%s", failures, out)
	}
	events := []integrationEvent{
		{`EXIT  | role=mixed-first | 5`, "the first step to fail before retry exhaustion"},
		{`stepName="First step" stepIndex=0 status=exhausted retries=2`, "the first step to exhaust both retries before proceeding"},
		{`START | role=mixed-second |`, "the second step to start after the first step exhausts retries"},
		{`EXIT  | role=mixed-second | 5`, "the second step to fail on its initial attempt"},
		{`START | role=mixed-second |`, "the second step to retry after its initial failure"},
		{`EXIT  | role=mixed-second | 0`, "the second step to recover successfully on retry"},
		{`stepName="Second step" stepIndex=1 status=completed retryAttempt=1`, "the second step to report recovery on retry 1"},
		{`START | role=mixed-third |`, "the third step to start after the second step recovers"},
		{`EXIT  | role=mixed-third | 0`, "the third step to finish successfully"},
	}
	remaining := out
	for _, event := range events {
		position := bytes.Index(remaining, []byte(event.output))
		if position < 0 {
			t.Fatalf("expected %s; its event is missing or out of order (log detail: %q)\n%s", event.description, event.output, out)
		}
		remaining = remaining[position+len(event.output):]
	}
	if elapsed < 3*time.Second {
		t.Errorf("expected two retry waits for the exhausted step and one for the recovered step; execution took %s", elapsed)
	}
	assertIntegrationOutput(t, out, `stepCount=3`, "the scenario must construct all three configured steps")
	assertRetryRecoveryReports(t, out, 1)
	assertNoUnexpectedOutcome(t, out)
	assertWorkflowSummary(t, out, "Retry mixed outcome", "partial")
}

// Descriptions explain the behavior; output retains the exact log contract.
type integrationEvent struct {
	output      string
	description string
}

func assertIntegrationOutput(t *testing.T, out []byte, detail, expectation string) {
	t.Helper()
	if !bytes.Contains(out, []byte(detail)) {
		t.Fatalf("%s; expected log detail %q was missing\n%s", expectation, detail, out)
	}
}

func assertRetryRecoveryReports(t *testing.T, out []byte, want int) {
	t.Helper()
	reports := 0
	for _, line := range bytes.Split(out, []byte("\n")) {
		if !bytes.Contains(line, []byte(`msg="step retry"`)) || !bytes.Contains(line, []byte(`status=completed`)) {
			continue
		}
		reports++
		if !bytes.Contains(line, []byte(`level=INFO`)) {
			t.Fatalf("successful retry recovery must be reported at INFO level, not as an error\n%s", out)
		}
	}
	if reports != want {
		t.Fatalf("expected %d successful retry recovery reports for the steps that actually recovered; got %d\n%s", want, reports, out)
	}
}

// Valid execution scenarios must never leave their outcome unresolved or take
// an internal-error branch, even if an expected summary is also present.
func assertNoUnexpectedOutcome(t *testing.T, out []byte) {
	t.Helper()
	lower := bytes.ToLower(out)
	for _, diagnostic := range []string{"status=unknown", "contact developer", "contact the developer"} {
		if bytes.Contains(lower, []byte(diagnostic)) {
			t.Fatalf("a supported workflow scenario must establish a known outcome without developer-contact diagnostics; found %q\n%s", diagnostic, out)
		}
	}
}

func assertWorkflowSummary(t *testing.T, out []byte, name, status string) {
	t.Helper()
	identity := fmt.Sprintf("workflow=%q", name)
	reports := 0
	var summary []byte
	for _, line := range bytes.Split(out, []byte("\n")) {
		if bytes.Contains(line, []byte("msg=workflow ")) && bytes.Contains(line, []byte(identity)) && !bytes.Contains(line, []byte(" status=started ")) {
			reports++
			summary = line
		}
	}
	if reports != 1 {
		t.Fatalf("workflow %q must emit exactly one final report with status %s and total duration; found %d\n%s", name, status, reports, out)
	}
	var reportedStatus, durationText string
	for _, field := range strings.Fields(string(summary)) {
		if strings.HasPrefix(field, "status=") {
			reportedStatus = strings.TrimPrefix(field, "status=")
		}
		if strings.HasPrefix(field, "duration=") {
			durationText = strings.TrimPrefix(field, "duration=")
		}
	}
	if reportedStatus != status {
		t.Fatalf("workflow %q must finish as %s; its final report says %q\n%s", name, status, reportedStatus, out)
	}
	duration, err := time.ParseDuration(durationText)
	if err != nil || duration < 0 {
		t.Fatalf("workflow %q must report a valid, nonnegative elapsed duration; got %q (parse error: %v)\n%s", name, durationText, err, out)
	}
}
