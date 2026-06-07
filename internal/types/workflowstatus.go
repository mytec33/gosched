package types

type WorkflowStatus struct {
	value string
}

var (
	WorkflowStatusCompleted = WorkflowStatus{value: "completed"}
	WorkflowStatusPartial   = WorkflowStatus{value: "partial"}
	WorkflowStatusFailed    = WorkflowStatus{value: "failed"}
	WorkflowStatusSkipped   = WorkflowStatus{value: "skipped"}
)

func (s WorkflowStatus) String() string {
	return s.value
}
