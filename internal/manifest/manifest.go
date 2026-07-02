// Package manifest provides schedules stored within a "manifest" file
package manifest

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrOpenManifest = errors.New("unable to open manifest")
	ErrScanManifest = errors.New("unable to scan manifest")
)

type WorkflowFiles []string

func (s *WorkflowFiles) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *WorkflowFiles) String() string {
	return fmt.Sprintf("%v", *s)
}

func ParseManifest(path string) (WorkflowFiles, error) {
	var schedules WorkflowFiles

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrOpenManifest, path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		schedules = append(schedules, line)
	}

	err = scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("%w: %q: %w", ErrScanManifest, path, err)
	}

	return schedules, nil
}
