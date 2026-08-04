# Review findings — 2026-08-03

Merged from two independent reviews of the same commit (815977d). Every item below
was confirmed against source; measured output is included where it exists.

`DEFECT` = code does not match apparent intent. `POLICY` = needs a ruling, not a bug.

---

## P1 — Operational risk

- [ ] **`DEFECT` A step with no configured timeout runs unbounded.** `internal/runner/runner.go:27`

  ```go
  if step.Timeout.Duration() > 0 {
      ctx, cancel = context.WithTimeout(context.Background(), step.Timeout.Duration())
      cmd = exec.CommandContext(ctx, step.Program, step.Args...)
  } else {
      cmd = exec.Command(step.Program, step.Args...)   // no deadline, ever
  }
  ```

  Nothing in `Validate` requires `timeout`, so the unbounded branch is the default. History
  shows this is a survivor, not a design: plain `exec.Command` arrived with the initial
  scheduler (`68b5806`, 2025-12-06) and `CommandContext` was added *beside* it ten weeks
  later for per-step timeouts (`05919ea`, 2026-02-14). The pairing was never revisited.

  **Approach — fix in validation, not execution (Bill's call).** Materialize a default
  timeout in `WorkflowRaw.Validate` so the trusted `Workflow` never carries a zero. Caught
  far earlier, and the zero value never escapes the raw/trusted boundary that split exists
  to enforce.

  **It simplifies three places:**
  - `runStepCommand` — loses the `if/else`, the `var cmd` / `var ctx` / `var cancel`
    declarations, the `if cancel != nil` block, and the `ctx != nil` guard on line 41.
    One path.
  - `printStep` (`internal/schedule/print.go:94`) — the `"no timeout"` else branch becomes
    unreachable and can be deleted.
  - `printStep` line 104-107 — `details` always gets an element from that same if/else, so
    `len(details) > 0` is already always true and the bare-name `Fprintf` on line 107 is
    dead today, before any change.

  **Two traps:**
  - A default timeout alone does **not** guarantee `runStepCommand` returns. The deadline
    kills the child, but `CombinedOutput` blocks until the stdout/stderr pipes hit EOF, and
    a surviving grandchild holds the write end open. `cmd.WaitDelay` (Go 1.20+, repo is on
    1.25.1) is what actually forces the return. Same change, not a follow-up.
  - `print_test.go:192` and `:197` build a zero-timeout `Step` by hand and bypass
    validation, so they will keep **passing** while asserting an unreachable state. Delete
    the `"no timeout"` branch in the same change to turn a silent lie into two loud failures.

  **`POLICY`** — is the default generous-but-finite, or does `timeout` become required?
  Required converts the whole class into a startup config error. Also unresolved: no maximum
  on step timeout, and no maximum on step pause (retry pause is capped at 7200s).

  Suggested order: default at validation → delete the `"no timeout"` branch → `CommandContext`
  always → `WaitDelay` → scheduler-context threading as a separate pass (see next item).

- [ ] **`DEFECT` Shutdown can wait forever.** `internal/runner/scheduler.go:100`

  On SIGTERM the loop stops scheduling and calls `running.Wait()` with no grace period.
  Running workflows never receive the scheduler context:
  - `executeWorkflow(wf)` takes no `ctx` (`scheduler.go:157`)
  - steps without a configured timeout use plain `exec.Command` (`runner.go:31`)
  - retry pause is `time.Sleep` (`runner.go:91`), between-step pause is `time.Sleep` (`scheduler.go:213`)

  A single hung program blocks shutdown indefinitely.

  Fix: thread `ctx` through `runSchedulerTick` → `executeWorkflow` → `RunStepAttempt` →
  `runStepCommand`; use `exec.CommandContext` always; replace both sleeps with
  `select { case <-ctx.Done(): ... case <-time.After(d): }`; add a maximum grace period
  around `running.Wait()`.

  Note: `TODO.md` lists graceful shutdown as done. It is done for *scheduling*, not for
  *in-flight work*.

  **Trap — error classification breaks when the parent becomes the scheduler context.**
  Line 41 of `runner.go` currently treats a fired context as a timeout:

  ```go
  if ctx != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
  ```

  Once the parent is the scheduler's context there are two causes: `DeadlineExceeded` (the
  step blew its own timeout — a real failure) and `Canceled` (operator hit Ctrl-C — not a
  failure). As written, a shutdown-killed step falls past that branch and is logged as
  `status=failed`, so every clean shutdown produces a burst of spurious failures — and
  `RunStepRetries` starts retrying steps while the scheduler is trying to exit. The retry
  loop needs its own cancellation check, not just a cancellable sleep.

- [ ] **`DEFECT` Backward clock movement re-executes workflows.**
  `internal/workflow/minute_of_day.go:23` + `internal/runner/scheduler.go:68`

  `MinutesSince` is `(m - previous + 1440) % 1440`, so any backward step reads as a
  near-full-day forward jump, lands in the `default` branch, and runs the tick again.

  ```
  midnight rollover 23:59 -> 00:00     MinutesSince = 1        correct
  DST fall-back    02:00 -> 01:00      MinutesSince = 1380     re-runs
  NTP step back 5m 10:05 -> 10:00      MinutesSince = 1435     re-runs
  suspend exactly 24h 09:00 -> 09:00   MinutesSince = 0        silently suppressed
  ```

  The NTP case is the likely one — a routine backward correction re-executes that
  minute's workflows. Midnight rollover itself is correct.

  Fix: compare monotonic wall-clock instants rather than minute-of-day arithmetic, or
  refuse to run when `now` precedes the last processed instant.

---

## P2 — Resource and disclosure

- [ ] **`DEFECT` Command output is unbounded in memory.** `internal/runner/runner.go:33`

  `cmd.CombinedOutput()` buffers all stdout+stderr, then the whole thing is passed to
  logging. A runaway program can OOM the scheduler.

  Fix: bounded buffer or streaming with an explicit cap and a truncation marker.

- [ ] **`DEFECT` `NewWorkflowLogger` drops the workflow name.** `internal/logging/logging.go:22`

  The `workflowName` parameter is accepted and never used; only `wfRunID` is attached.
  Go does not warn on unused parameters. Compounding it, terminal abort errors are
  emitted by the global `logging.StdOut` (`scheduler.go:149`), so they carry neither
  workflow name nor run ID. Concurrent failures are hard to attribute.

  Fix: `StdOut.With(slog.String("workflow", workflowName), slog.String("wfRunID", wfID))`
  and route every event in the run through it.

---

## P3 — Correctness and consistency

- [ ] **`DEFECT` Mode flags silently override each other.** `cmd/gosched/main.go:51`, `:102`, `:110`

  `gosched -manifest m.json -run-once "Backup" -print-schedule config` prints the
  schedule and exits 0; `-run-once` is discarded with no message. `-new-config` beats
  everything. Precedence is whatever order the `if` blocks sit in.

  Fix: reject more than one mode flag with `ExitInvalidArgs`.

- [ ] **`POLICY` Duplicate-name matching differs by level.**

  ```go
  // internal/schedule/validation.go:33 — workflows
  wfKey := strings.TrimSpace(strings.ToLower(wf.Name))
  // internal/workflow/workflow_raw.go:217 — steps
  _, exists := stepNames[step.Name]
  ```

  Workflows `"Build"` / `"build"` are duplicates. Steps `"Build"` / `"build"` are not.
  The workflow side is deliberate — `schedule_test.go` tests mixed case and padding.

- [ ] **`POLICY` Retry block accepted with a non-retry `onFailure`.**
  `internal/workflow/workflow_raw.go:114`, `:120`

  `onFailure: "retry"` with no retry block errors. The reverse — `onFailure: "abort"`
  *with* a retry block — is accepted and copied into the trusted workflow. The
  analogous `disabledReason` contradiction has an explicit `ErrDisabledReasonNotAllowed`.

- [ ] **`POLICY` Length caps count bytes, not runes.** `internal/workflow/workflow_raw.go:205`

  `len(trimmed) > maxLength`. With `MaxWorkflowNameLength = 256`, a name of 128
  two-byte characters hits the cap. Same question you settled for cadence.

---

## P4 — Nits

- [ ] `internal/workflow/workflow_raw.go:214` — `var errors []error` shadows the imported
      `errors` package inside that function. Compiles today; a trap for the next edit.
- [ ] `internal/workflow/workflow_raw.go:148` — `for i, steps := range raw.Steps` names a
      single element plural.
- [ ] `internal/workflow/workflow_raw.go:34-37` — `ErrRetryCountNegative` and
      `ErrRetryCountTooLarge` carry identical message text; likewise the two pause errors.
      Distinct sentinels, indistinguishable output.
- [ ] `internal/workflow/minute_of_day.go:22`, `:35` — pointer receivers on
      `MinuteOfDay int`. `MinuteOfDayFromTime(now).MinutesSince(prev)` will not compile
      (`cannot call pointer method`); `scheduler.go:65` works only because it assigns to a
      variable first. Ties into the deferred value-semantics pass.

---

## Architectural — decide, then document

- [ ] Skipped minutes are logged but never replayed after suspension or a forward clock jump.
- [ ] A suspension of exactly 24h reads as `MinutesSince = 0` and is suppressed as a duplicate.
- [ ] Workflow locks are process-local and vanish on restart, so two scheduler instances can
      run the same workflow concurrently.

Reasonable constraints for a lightweight single-instance scheduler — worth stating in
`README.md` if intentional.

---

## Ruled closed — working as intended

**Logging step args and command output.** `internal/runner/runner.go:58`, `:65`

Raised as a disclosure risk (credentials, tokens, URLs, customer data in logs).
**Ruled by Bill, 2026-08-03: intended behaviour, closed.** He does not write credentials
or other sensitive values into workflow configuration, and knowing exactly what the
program ran when it failed is the point — that record is what makes a failure
reproducible. Do not re-raise.

Note: the unbounded-memory item in P2 is a separate concern (volume, not sensitivity) and
is unaffected by this ruling. Any output cap added there should preserve args in full.

---

## Confirmed good

- JSON decoding rejects unknown fields and trailing content.
- Invalid workflows never reach the trusted schedule.
- Commands use `exec.Command(program, args...)`, not a shell — no ordinary shell injection.
- `SafeMap` locking is mutex-protected.
- Schedule expansion replaces its prior map, so it is idempotent.
- Midnight rollover arithmetic is correct.
- Full test suite passes: `ok` for `cmd/testprog`, `internal/decode`, `internal/runner`,
  `internal/schedule`, `internal/workflow`.
