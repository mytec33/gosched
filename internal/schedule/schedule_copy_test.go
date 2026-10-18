package schedule_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mytec33/gosched/internal/schedule"
	"github.com/mytec33/gosched/internal/workflow"
)

func TestScheduleWorkflowGettersReturnIndependentCopies(t *testing.T) {
	manifestFile := filepath.Join("testdata", "schedule", "protected_manifest.txt")
	sched, errs := schedule.New(manifestFile)
	if len(errs) > 0 {
		t.Fatalf("construct schedule: %v", errs)
	}

	original, errs := schedule.New(manifestFile)
	if len(errs) > 0 {
		t.Fatalf("construct reference schedule: %v", errs)
	}
	want := original.Workflows()
	if len(want) != 2 {
		t.Fatalf("expected 2 fixture workflows, got %d", len(want))
	}

	minute, err := workflow.ParseMinuteOfDay("00:00")
	if err != nil {
		t.Fatal(err)
	}

	copies := [][]workflow.Workflow{sched.Workflows(), sched.WorkflowsAtMinute(minute)}
	for _, workflows := range copies {
		if !reflect.DeepEqual(workflows, want) {
			t.Fatal("getter did not return the expected workflows")
		}
		for i := range workflows {
			workflows[i].Name = "changed workflow"
			for j := range workflows[i].Steps {
				workflows[i].Steps[j].Program = "changed program"
				for k := range workflows[i].Steps[j].Args {
					workflows[i].Steps[j].Args[k] = "changed argument"
				}
			}
		}
	}

	named, err := sched.GetWorkflowByName(want[0].Name)
	if err != nil {
		t.Fatalf("get workflow by name: %v", err)
	}
	if !reflect.DeepEqual(named, want[0]) {
		t.Fatal("name lookup did not return the expected workflow")
	}
	named.Name = "changed workflow"
	for j := range named.Steps {
		named.Steps[j].Program = "changed program"
		for k := range named.Steps[j].Args {
			named.Steps[j].Args[k] = "changed argument"
		}
	}

	got := sched.Workflows()
	if !reflect.DeepEqual(got, want) {
		t.Error("changing returned data modified the schedule's workflows")
	}
	got = sched.WorkflowsAtMinute(minute)
	if !reflect.DeepEqual(got, want) {
		t.Error("changing returned data modified the schedule's minute index")
	}
}
