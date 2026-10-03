package schedule

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

var fixturePrintScheduleCOnfig = `
testdata/config_print1.json:
1: 1h 11:46  Workflow 1 (abort)
		1: step 1 (timeout 30s)
2: 1h 11:45  Workflow 2 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)

testdata/config_print2.json:
3: 1h 11:45  Workflow 3 (continue)
		1: step 1 (timeout 30s)
4: 1h 12:47  Workflow 4 (retry)
		1: step 1 (timeout 30m0s)
5: 1h 12:47  Workflow 5 (retry) <--- disabled: demonstrating disabled display
		1: step 1 (timeout 30m0s)

`

func TestPrintScheduleConfig(t *testing.T) {
	manifestFile := filepath.Join("testdata", "config_manifest.txt")
	s, errorList := New(manifestFile)
	if len(errorList) > 0 {
		t.Fatalf("expected no errors, got %v\n", errorList)
	}

	var buf bytes.Buffer
	if err := s.printScheduleConfig(&buf); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if got != fixturePrintScheduleCOnfig {
		t.Fatalf("unexpected output\ngot:\n%s\nwant:\n%s", got, fixturePrintScheduleCOnfig)
	}
}

func TestDisplayReason(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "short reason unchanged", input: "maintenance", want: "maintenance"},
		{name: "exactly max unchanged", input: strings.Repeat("a", 40), want: strings.Repeat("a", 40)},
		{name: "over max truncated with ellipsis", input: strings.Repeat("a", 41), want: strings.Repeat("a", 40) + "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := displayReason(tt.input)
			if got != tt.want {
				t.Fatalf("displayReason(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

var fixturePrintScheduleOperational = `1: 11:45  Workflow 1 (abort)
		1: step 1 (timeout 30s, pause 5s)
		2: step 2 (timeout 30s)
2: 11:45  Workflow 2 (continue)
		1: step 1 (timeout 30s)

1: 11:46  Workflow 3 (abort)
		1: step 1 (timeout 30s)

1: 12:47  Workflow 1 at 12:47 (retry)
		1: step 1 (timeout 30m0s)
2: 12:47  Workflow 4 (retry) <--- disabled: disabled to test enable functionality
		1: step 1 (timeout 30m0s)
`

func TestPrintScheduleOperational(t *testing.T) {
	manifestFile := filepath.Join("testdata", "operational_manifest.txt")
	s, errorList := New(manifestFile)
	if len(errorList) > 0 {
		t.Fatalf("expected no errors, got %v\n", errorList)
	}

	var buf bytes.Buffer
	if err := s.printScheduleOperational(&buf); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if got != fixturePrintScheduleOperational {
		t.Fatalf("unexpected output\ngot:\n%s\nwant:\n%s", got, fixturePrintScheduleOperational)
	}
}

var fixturePrintVariationsConfig = `
testdata/print_variations.json:
1: 1d 06:00  Step display variations (abort)
		1: timeout and pause (timeout 30s, pause 5s)
		2: timeout only (timeout 30s)
2: 1d 07:00  Disabled workflow (continue) <--- disabled: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...
		1: disabled step (timeout 30s)

`

var fixturePrintVariationsOperational = `1: 06:00  Step display variations (abort)
		1: timeout and pause (timeout 30s, pause 5s)
		2: timeout only (timeout 30s)

1: 07:00  Disabled workflow (continue) <--- disabled: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...
		1: disabled step (timeout 30s)
`

func TestPrintScheduleVariations(t *testing.T) {
	manifestFile := filepath.Join("testdata", "print_variations_manifest.txt")
	s, errorList := New(manifestFile)
	if len(errorList) > 0 {
		t.Fatalf("expected no errors, got %v", errorList)
	}

	tests := []struct {
		method string
		want   string
	}{
		{method: "config", want: fixturePrintVariationsConfig},
		{method: "operational", want: fixturePrintVariationsOperational},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			var buf bytes.Buffer
			if err := s.Print(tt.method, &buf); err != nil {
				t.Fatalf("print %s schedule: %v", tt.method, err)
			}

			if got := buf.String(); got != tt.want {
				t.Fatalf("unexpected output\ngot:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

// failPrintWriter fails at a chosen write and counts any writes after failure.
type failPrintWriter struct {
	writes int
	failAt int
	err    error
}

func (w *failPrintWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes >= w.failAt {
		return 0, w.err
	}
	return len(p), nil
}

func TestPrintWriteErrors(t *testing.T) {
	s, errs := New(filepath.Join("testdata", "print_variations_manifest.txt"))
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	want := errors.New("output write failed")
	for _, method := range []string{"config", "operational"} {
		// Count successful writes so every output position is tested, including separators.
		counter := &failPrintWriter{failAt: int(^uint(0) >> 1), err: want}
		if err := s.Print(method, counter); err != nil {
			t.Fatal(err)
		}
		for failAt := 1; failAt <= counter.writes; failAt++ {
			t.Run(fmt.Sprintf("%s/write-%d", method, failAt), func(t *testing.T) {
				w := &failPrintWriter{failAt: failAt, err: want}
				if err := s.Print(method, w); !errors.Is(err, want) {
					t.Fatalf("got %v, want %v", err, want)
				}
				if w.writes != failAt {
					t.Fatalf("continued writing after failure: %d writes, want %d", w.writes, failAt)
				}
			})
		}
	}
}
