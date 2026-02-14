// Package platform provides functions based on platform which are compiled in based on platform
package platform

import (
	"errors"
	"strings"
)

var (
	ErrPlatformLeadingOrTrailingWhitespace = errors.New("leading or trailing whitespace")
	ErrPlatformTooLong                     = errors.New("too long")
)

func ValidateProgramPath(s string) error {
	if strings.TrimSpace(s) != s {
		return ErrPlatformLeadingOrTrailingWhitespace
	}
	if len(s) > maxPathLength() {
		return ErrPlatformTooLong
	}
	return nil
}
