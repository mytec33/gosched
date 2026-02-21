//go:build openbsd

// Package platform provides details about the underlying platform.
package platform

func MaxPathLength() int {
	return 1024
}
