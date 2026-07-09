# TODO

```
Evaluate JSON tag ownership after enabled/disabled work.

WorkflowRaw owns top-level JSON decoding, but nested structs such as Trigger,
Step, and RetryPolicy are still shared between raw decoding and trusted workflow
values. Decide later whether to split raw nested structs from trusted nested
structs, or keep the current shared shape and accept JSON tags on trusted values.
```

- ~~graceful scheduler shutdown

- think about Configuration as Code (CaC) might be how this scheduler can work

- ~~add flag -print-schedule=config~~
- ~~add flag -create-new to create a sample configuration file to work off of~~

### Validation
1. ~~Workflow names must be unqiue.~~
2. ~~Step names within a workflow must be unique.~~