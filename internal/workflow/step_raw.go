package workflow

type StepRaw struct {
	Name    string          `json:"name"`
	Program string          `json:"program"`
	Args    []string        `json:"args"`
	Timeout *ConfigDuration `json:"timeout"`
	Pause   ConfigDuration  `json:"pause"`
}
