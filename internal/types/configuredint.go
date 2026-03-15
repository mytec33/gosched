// Package types provides configuration types
package types

import (
	"encoding/json"
	"time"

	"git.sr.ht/~mytec/gosched/internal/errs"
)

type ConfiguredInt struct {
	v int
}

func (c ConfiguredInt) Int() int {
	return c.v
}

func (c ConfiguredInt) Configured() bool {
	return c.v > 0
}

func (c ConfiguredInt) Duration() time.Duration {
	return time.Duration(c.Int()) * time.Second
}

func NewConfiguredInt(i int) ConfiguredInt {
	return ConfiguredInt{v: i}
}

func (c *ConfiguredInt) UnmarshalJSON(b []byte) error {
	var tmp int

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	if tmp < 0 {
		return errs.ErrNegativeNumber
	}

	c.v = tmp
	return nil
}
