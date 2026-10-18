package workflow

type WorkflowStatus struct {
	value string
}

var (
	StatusCompleted = WorkflowStatus{value: "completed"}
	StatusPartial   = WorkflowStatus{value: "partial"}
	StatusFailed    = WorkflowStatus{value: "failed"}
	StatusSkipped   = WorkflowStatus{value: "skipped"}
	StatusUnknown   = WorkflowStatus{value: "unknown"}
)

func (s WorkflowStatus) String() string {
	return s.value
}
