package workflow

import "fmt"

type RetryPolicy struct {
	NumberRetries int `json:"numberRetries"`
	PauseSeconds  int `json:"pauseSeconds"`
}

func (t *RetryPolicy) String() string {
	return fmt.Sprintf("%v %v", t.NumberRetries, t.PauseSeconds)
}
