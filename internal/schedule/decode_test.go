package schedule

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
)

const validOneWorkflowOneStep = `
[
  {
    "name": "Workflow 1",
    "time": "10:35",
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": "--sleep 10 --role daily-slot-ratings"
      }
    ]
  }
]
`

const validOneWorkflowTwoSteps = `
[
  {
    "name": "Workflow 1",
    "time": "10:35",
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": "--sleep 10 --role daily-slot-ratings"
      },
      {
        "name": "modified",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": "--sleep 50 --role modified-slot-ratings"
      }
    ]
  }
]
`

const validTwoWorkflows = `
[
  {
    "name": "Workflow 1",
    "time": "10:35",
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": "--sleep 10 --role daily-slot-ratings"
      },
      {
        "name": "modified",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": "--sleep 50 --role modified-slot-ratings"
      }
    ]
  },
  {
    "name": "Workflow 2",
    "time": "10:35",
    "steps": [
      {
        "name": "daily",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": "--sleep 5 --role daily-table-ratings"
      },
      {
        "name": "modified",
        "program": "/Users/user/some_path/go/gosched/testprog",
        "args": "--sleep 25 --role modified-table-ratings"
      }
    ]
  }
]
`

func TestDecode_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "empty", json: ``},
		{name: "syntax error 1", json: `{`},
		{name: "syntax error 2", json: `{}{}`},
		{name: "unknown field", json: `[{"foo": "bar"}]"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)

			_, _, err := DecodeSchedule(r)
			if err == nil {
				t.Fatal("expected error, got no error")
			} else if !errors.Is(err, ErrDecodeSchedule) {
				t.Fatalf("%v: expected ErrDecodeSchedule, got %v", tt.name, err)
			}
		})
	}
}

func TestDecode_ValidInput(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "One work flow, one step", json: validOneWorkflowOneStep},
		{name: "One work flow, two steps", json: validOneWorkflowTwoSteps},
		{name: "Two work flows", json: validTwoWorkflows},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.json)

			_, _, err := DecodeSchedule(r)
			if err != nil {
				t.Fatalf("%v: expected no error, got %v", tt.name, err)
			}
		})
	}
}

func TestReadScheduleFile_OK(t *testing.T) {
	filename := writeTempFile(t, validOneWorkflowOneStep)

	_, _, err := ReadScheduleFile(filename)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestReadScheduleFile_OpenError(t *testing.T) {
	_, _, err := ReadScheduleFile("/path/that/does/not/exist.json")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "open workflows file") {
		t.Fatalf("expected open context, got %v", err)
	}
}

func TestReadScheduleFile_DecodeErrorIncludesFilename(t *testing.T) {
	filename := writeTempFile(t, "{") // intentionally invalid JSON

	_, _, err := ReadScheduleFile(filename)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "decode workflows file") {
		t.Fatalf("expected decode context, got %v", err)
	}
	if !strings.Contains(err.Error(), strconv.Quote(filename)) {
		t.Fatalf("expected filename in error, got %v", err)
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "schedule-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}

	_, err = f.WriteString(content)
	if err != nil {
		f.Close()
		t.Fatalf("WriteString: %v", err)
	}

	err = f.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	return f.Name()
}
