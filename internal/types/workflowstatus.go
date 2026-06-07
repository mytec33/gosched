package types

type WorkflowStatus struct {
	value string
}

func WorkflowStatusCompleted() WorkflowStatus {
	return WorkflowStatus{value: "completed"}
}

func WorkflowStatusPartial() WorkflowStatus {
	return WorkflowStatus{value: "partial"}
}

func WorkflowStatusFailed() WorkflowStatus {
	return WorkflowStatus{value: "failed"}
}

func WorkflowStatusSkipped() WorkflowStatus {
	return WorkflowStatus{value: "skipped"}
}

func (s WorkflowStatus) String() string {
	return s.value
}
