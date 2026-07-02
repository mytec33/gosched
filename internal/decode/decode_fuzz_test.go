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

	f.Add([]byte(`[{"name":"x","trigger":{"every":"1d","beginAt":"00:00"},"onFailure":"continue","steps":[{"name":"s","program":"echo"}]}]`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`{"not":"array"}`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		wfs, validationErrors, err := DecodeWorkflowFile(bytes.NewReader(data))
		if err != nil || len(validationErrors) > 0 {
			return
		}

		sched := schedule.FromWorkflows(wfs)
		sched.Print("config", io.Discard)
		sched.Print("operational", io.Discard)

		for _, wf := range sched.Workflows() {
			_ = wf.Trigger.String()
			_ = workflow.WorkflowAbortsOnFailure(wf)
			_ = workflow.WorkflowContinuesOnFailure(wf)
			_ = workflow.WorkflowRetriesOnFailure(wf)
		}
	})
}
