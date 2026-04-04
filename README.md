# goSched

A scheduler designed as simple replacement for Windows Task Scheduler with multiple
config files and operational visibility.

## Why

Windows Task Scheduler makes it difficult to:

- understand what runs at a given time
- reason about job collisions
- control execution order across multiple entries, especially those in different folders
- see program output and exit codes

gosched addresses this by:

- merging multiple config fles in declared order
- logging program output and exit codes
- providing operational view of scheduled work


## Key Concepts

- **Minute bucket**  
  Workflows are grouped by minute of execution.

- **Config precedence**  
`config1 + config2 != config2 + config1`  
Later files append workflows and implicitly take priority.

- **Failure policies**
- `abort` — stop remaining workflows in the bucket
- `continue` — proceed to next workflow

## Usage

```bash
gosched -schedule base.json -schedule site.json
```

Muliple `-schedule` flags are supported. Files are merged in the order provided.

## Output modes

### Config View

Displays workflows as defined in configuration files. Useful for verifying correctness of configuration.

Example:
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

### Operational View

Displays workflows grouped by execution time.

Example:
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

Useful for:

- understanding workload at a given time
- identifying scheduling collisions
- planning concurrency limits

## Roadmap

gosched is intentionally simple today. The following features are planned:

### Near Term

- Retry (basic)
- Concurrency limits per minute
- Improved logging of workflow execution

### Medium Term

- Retry policy refinement
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