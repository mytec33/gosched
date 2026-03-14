// Package types provides configuration types
package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"git.sr.ht/~mytec/gosched/internal/errs"
)

type ConfiguredArg struct {
	v string
}

func (c ConfiguredArg) String() string {
	return c.v
}

func (c ConfiguredArg) Configured() bool {
	return len(c.v) > 0
}

func (c *ConfiguredArg) UnmarshalJSON(b []byte) error {
	var tmp string

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	if tmp == "" {
		return errs.ErrEmpty
	}

	trimmed := strings.TrimSpace(tmp)

	if trimmed == "" {
		return errs.ErrWhitespaceAll
	}

	if trimmed != tmp {
		return fmt.Errorf("program argument value %q: %w", tmp, errs.ErrWhitespaceLeadingOrTrailing)
	}

	if len(tmp) > errs.MaxProgramArgsLengths {
		snippet := tmp
		if len(tmp) > 25 {
			snippet = tmp[:25]
		}
		return fmt.Errorf("program argument value %q...: %w", snippet, errs.ErrTooLong)
	}

	c.v = tmp
	return nil
}
