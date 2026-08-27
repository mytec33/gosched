package decode

import (
	"bytes"
	"testing"
)

func FuzzDecodeWorkflows(f *testing.F) {
	// LLM provided
	// go test ./internal/decode -fuzz=FuzzDecodeWorkflows -fuzztime=5m

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
		const sourceFileName = "fuzz-input.json"
		wfs, validationErrors, err := decodeWorkflows(bytes.NewReader(data), sourceFileName)
		if err != nil {
			if len(wfs) > 0 {
				t.Fatalf("decode error returned with %d trusted workflows", len(wfs))
			}
			return
		}

		if len(validationErrors) > 0 {
			if len(wfs) > 0 {
				t.Fatalf("validation errors returned with %d trusted workflows", len(wfs))
			}
			return
		}

		for i, wf := range wfs {
			if wf.SourceFile != sourceFileName {
				t.Fatalf("workflow %d source file = %q, want %q", i, wf.SourceFile, sourceFileName)
			}
			if wf.Name == "" {
				t.Fatalf("workflow %d has no name", i)
			}
			if len(wf.Steps) == 0 {
				t.Fatalf("workflow %d has no steps", i)
			}
		}
	})
}
