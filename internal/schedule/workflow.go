// Package schedule provides schedule layout and locking
package schedule

import "errors"

type Workflow struct {
	Name  string `json:"name"`
	Time  string `json:"time"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Name    string `json:"name"`
	Program string `json:"program"`
	Args    string `json:"args"`
}

func (w Workflow) Validate() []error {

	return []error{errors.New("foo")}
}
