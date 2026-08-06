package decode

import (
	"bytes"
	"io"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/schedule"
	"git.sr.ht/~mytec/gosched/internal/workflow"
)

func FuzzDecodeWorkflows(f *testing.F) {
	// LLM provided
	// go test ./internal/decode -fuzz=FuzzDecodeWorkflows -fuzztime=5m

	f.Add([]byte(validOneWorkflowOneStep))
	f.Add([]byte(validFullExample))
	f.Add([]byte(`[]`))
	f.Add([]byte(`{"not":"array"}`))
	f.Add([]byte(``))

	f.Add([]byte(`[
  {
    "name":"durations",
    "enabled":true,
    "trigger":{"every":"15m","beginAt":"23:59"},
    "onFailure":"continue",
    "steps":[{
      "name":"s",
      "program":"echo",
      "timeout":"30s",
      "pause":"5s"
    }]
  }
]`))

	// Valid structure with a malformed typed leaf.
	f.Add([]byte(`[
  {
    "name":"x",
    "enabled":true,
    "trigger":{"every":"1d","beginAt":"25:99"},
    "onFailure":"continue",
    "steps":[{"name":"s","program":"echo"}]
  }
]`))

	// Deep unknown field.
	f.Add([]byte(`[
  {
    "name":"x",
    "enabled":true,
    "trigger":{"every":"1d","beginAt":"00:00"},
    "onFailure":"continue",
    "steps":[{"name":"s","program":"echo","unknown":true}]
  }
]`))

	// Valid document followed by trailing content.
	f.Add([]byte(`[] []`))

	f.Fuzz(func(t *testing.T, data []byte) {
		sourceFileName := "not used"
		wfs, validationErrors, err := DecodeWorkflows(bytes.NewReader(data), sourceFileName)
		if err != nil || len(validationErrors) > 0 {
			return
		}

		sched := schedule.FromWorkflows(wfs)

		// Gate on schedule validation exactly as main.go does before expanding.
		if len(sched.Validate()) > 0 {
			return
		}
		if err := sched.ExpandSchedule(); err != nil {
			return
		}

		sched.Print("config", io.Discard)
		sched.Print("operational", io.Discard)

		for _, wf := range sched.Workflows() {
			_ = wf.Trigger.String()
			_, _ = sched.GetWorkflowByName(wf.Name)
		}
		// Exercise the byMinute index the runner reads from.
		for m := workflow.MinuteOfDay(0); m < workflow.MinuteOfDay(1440); m++ {
			_ = sched.WorkflowsAtMinute(m)
		}
	})
}
