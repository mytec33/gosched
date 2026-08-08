package workflow

type RetryPolicy struct {
	NumberRetries int `json:"numberRetries"`
	PauseSeconds  int `json:"pauseSeconds"`
}
