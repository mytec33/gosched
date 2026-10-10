// Package manifest provides schedules stored within a "manifest" file
package manifest

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mytec33/gosched/internal/logging"
)

var (
	ErrEmptyManifest = errors.New("manifest file must contain at least one file containing a workflow")
	ErrOpenManifest  = errors.New("unable to open manifest")
	ErrScanManifest  = errors.New("unable to scan manifest")
)

type Files []string

func ParseManifest(path string) (Files, error) {
	var schedules Files

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrOpenManifest, path, err)
	}
	defer func() {
		// This file is read-only; a close error does not affect the data already read.
		deferErr := file.Close()
		if deferErr != nil {
			logging.StdOut.Warn("configuration", "reason", "unable to close manifest file", "file", path, "error", deferErr)
		}
	}()

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

	if len(schedules) == 0 {
		return nil, ErrEmptyManifest
	}

	return schedules, nil
}
