// Package platform provides functions based on platform which are compiled in based on platform
package platform

import (
	"errors"
	"strings"
)

var (
	ErrPlatformEmpty                       = errors.New("cannot be empty")
	ErrPlatformWhitespaceAll               = errors.New("cannot be all whitespace")
	ErrPlatformWhitespaceLeadingOrTrailing = errors.New("leading or trailing whitespace")
	ErrPlatformTooLong                     = errors.New("too long")
)

func ValidateProgramPath(s string) error {
	if s == "" {
		return ErrPlatformEmpty
	}

	if strings.TrimSpace(s) == "" {
		return ErrPlatformWhitespaceAll
	}

	if strings.TrimSpace(s) != s {
		return ErrPlatformWhitespaceLeadingOrTrailing
	}

	if len(s) > maxPathLength() {
		return ErrPlatformTooLong
	}

	return nil
}
