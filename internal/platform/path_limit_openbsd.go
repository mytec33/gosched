//go:build openbsd

package platform

func MaxPathLength() int {
	return 1024
}
