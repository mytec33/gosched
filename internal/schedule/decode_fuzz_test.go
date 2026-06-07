package schedule

import (
	"bytes"
	"io"
	"testing"
)

func FuzzDecodeWorkflows(f *testing.F) {
	// LLM provided
	// go test ./internal/schedule -fuzz=FuzzDecodeWorkflows -fuzztime=5m

	f.Add([]byte(`[{"name":"x","trigger":{"every":"1d","beginAt":"00:00"},"onFailure":"continue","steps":[{"name":"s","program":"echo"}]}]`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`{"not":"array"}`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		s, validationErrors, err := DecodeWorkflows(bytes.NewReader(data))
		if err != nil || len(validationErrors) > 0 {
			return
		}

		s.PrintScheduleConfig(io.Discard)
		s.PrintScheduleOperational(io.Discard)

		for _, wf := range s.Workflows() {
			_ = wf.Trigger.String()
			_ = WorkflowAbortsOnFailure(wf)
			_ = WorkflowContinuesOnFailure(wf)
			_ = WorkflowRetriesOnFailure(wf)
		}
	})
}
