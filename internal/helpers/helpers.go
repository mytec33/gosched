// Package helpers provides shared functions across packages
package helpers

import (
	"time"
)

func SecondsDuration(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}
