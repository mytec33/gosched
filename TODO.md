# TODO

```
Consider indexing minute slots by workflow position instead of Workflow copies.

Keep s.workflows as the canonical list and store positions in byMinute (for
example, map[workflow.MinuteOfDay][]int). Minute queries resolve those positions
and continue returning independent deep copies through the public API.

An every-minute workflow currently contributes 1,440 Workflow struct copies.
Its nested steps and arguments share backing storage in the internal index;
they are not duplicated per slot. Integer positions could reduce the repeated
struct storage, but would not remove the allocation cost of public deep copies.

Measure memory and query costs with representative production schedules before
changing the representation. Preserve workflow order within each minute and
keep the canonical list stable for the schedule's lifetime. Reload should
continue constructing a complete replacement, including its own index.

Deferred idea from the Rust-port comparison; no implementation decision yet.
```

```
Evaluate JSON tag ownership after enabled/disabled work.

WorkflowRaw owns top-level JSON decoding, but nested structs such as Trigger,
Step, and RetryPolicy are still shared between raw decoding and trusted workflow
values. Decide later whether to split raw nested structs from trusted nested
structs, or keep the current shared shape and accept JSON tags on trusted values.
```

```
Track encoding/json/v2 stabilization.

When json/v2 becomes stable, evaluate replacing the current encoding/json
decode boundary. Relevant gosched benefits: duplicate JSON object names rejected
by default, invalid UTF-8 rejected by default, case-sensitive field matching by
default, RejectUnknownMembers for unknown fields, and simpler single-value input
handling. Keep WorkflowRaw.Validate as the authority for domain rules.
```

- ~~graceful scheduler shutdown

- think about Configuration as Code (CaC) might be how this scheduler can work

- ~~add flag -print-schedule=config~~
- ~~add flag -create-new to create a sample configuration file to work off of~~

### Validation
1. ~~Workflow names must be unqiue.~~
2. ~~Step names within a workflow must be unique.~~
