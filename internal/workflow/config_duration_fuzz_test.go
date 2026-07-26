package workflow

import (
	"errors"
	"strings"
	"testing"
)

// LLM provided
// go test ./internal/workflow -fuzz=FuzzParseConfigDuration -fuzztime=5m

func FuzzParseConfigDuration(f *testing.F) {
	for _, seed := range []string{
		"0s",
		"15H",
		"90m",
		"1.5h",
		"1h30m15s",
		"3m2h",
		"",
		"-1h",
		"+2s",
		"1",
		"1x",
		"999999999999999999999h",
		"１h",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		duration, err := ParseConfigDuration(input)

		if strings.HasPrefix(input, "-") ||
			strings.HasPrefix(input, "+") {
			if !errors.Is(
				err,
				ErrConfigDurationSignInvalid,
			) {
				t.Fatalf(
					"ParseConfigDuration(%q) error = %v, want %v",
					input,
					err,
					ErrConfigDurationSignInvalid,
				)
			}

			return
		}

		if err != nil {
			return
		}

		if duration.Duration() < 0 {
			t.Fatalf(
				"ParseConfigDuration(%q) produced negative duration %v",
				input,
				duration.Duration(),
			)
		}

		canonical := duration.String()

		reparsed, err := ParseConfigDuration(canonical)
		if err != nil {
			t.Fatalf(
				"ParseConfigDuration(%q) produced %q, which failed to parse: %v",
				input,
				canonical,
				err,
			)
		}

		if reparsed.Duration() != duration.Duration() {
			t.Fatalf(
				"ParseConfigDuration(%q) round trip = %v, want %v",
				input,
				reparsed.Duration(),
				duration.Duration(),
			)
		}
	})
}
