package workflow

type Step struct {
	Name    string
	Program string
	Args    []string
	Timeout ConfigDuration
	Pause   ConfigDuration
}
