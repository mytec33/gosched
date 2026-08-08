package workflow

import (
	"testing"
	"time"
)

func TestTriggerString(t *testing.T) {
	want := "1h 10:35"

	testTime, err := time.Parse("15:04", "10:35")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cadence, err := ParseCadence("1h")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	trigger := Trigger{Every: cadence, BeginAt: MinuteOfDayFromTime(testTime)}
	got := trigger.String()
	if want != got {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
