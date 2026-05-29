# goSched

gosched is a task scheduler designed as a simple replacement for Windows Task Scheduler. Written in Go, it's easy to deploy across myriad systems.

## Why

Managing large numbers of scheduled tasks with Windows Task Scheduler can become difficult when execution order, visibility, and operational troubleshooting are spread across many independent scheduler entries.

gosched addresses this by:

- merging multiple config files in declared order
- executing workflow steps sequentially
- logging program output and exit codes
- providing operational visibility into scheduled work


## Key Concepts

- **Minute bucket**  
  Workflows are grouped by minute of execution.

- **Config precedence**  
`config1 + config2 != config2 + config1`  
Later files append workflows and implicitly take priority.

- **Workflow failure policies**
- `abort` — stop the current workflow after a failed step; use when later steps depend on earlier outputs.
- `continue` — keep running later steps after a failed step; use when steps are independent and partial success is useful.
- `retry` — if a step fails, retry the step `numberRetries` times, pausing `pauseSeconds` between retries. After retries complete (successful or exhausted), continue to the next step.

## Usage

```bash
gosched -manifest ./manifest.txt
```

Manifest files are plain text lists of schedule JSON files.

- Blank lines are ignored.
- Lines beginning with # are ignored and treated as comments.
- Separate manifest files can be used for production, testing, or troubleshooting.

### Display the two configurations as they were imported.

```bash
gosched -manifest ./manifest.txt -print-schedule config
```

In this example the manifest file contains two configurations: config1.json and config2.json.

```
1: 11:45  Workflow 1 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)
2: 11:45  Workflow 2 (continue)
		1: step 1 (timeout 30s)
3: 11:46  Workflow 3 (abort)
		1: step 1 (timeout 30s)
4: 11:45  Workflow 1 - Test 2 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)
5: 11:45  Workflow 2 - Test 2 (continue)
		1: step 1 (timeout 30s)
6: 11:46  Workflow 3 - Test 2 (abort)
		1: step 1 (timeout 30s)
7: 13:45  Workflow 1 - Test 2 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)
```

### Display the configuration ordered by time.

```bash
gosched -schedule config1.json -schedule config2.json -print-schedule operational
```

```
1: 11:45  Workflow 1 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)
2: 11:45  Workflow 2 (continue)
		1: step 1 (timeout 30s)
3: 11:45  Workflow 1 - Test 2 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)
4: 11:45  Workflow 2 - Test 2 (continue)
		1: step 1 (timeout 30s)

1: 11:46  Workflow 3 (abort)
		1: step 1 (timeout 30s)
2: 11:46  Workflow 3 - Test 2 (abort)
		1: step 1 (timeout 30s)

1: 13:45  Workflow 1 - Test 2 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)

```

## Roadmap

gosched is intentionally simple today. The following features are planned:

### Near Term

- Concurrency limits per minute
- Improved logging of workflow execution

### Medium Term

- Visibility into running vs queued workflows
- Additional output modes

### Long Term

- Smarter scheduling strategies
- Resource-aware execution (CPU/load considerations)

## Status

v1.0 — usable

- Multi-config support implemented
- Deterministic execution model
- Operational schedule view available
- Used in production

Actively in development.