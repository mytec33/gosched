//go:build linux

package platform

func MaxPathLength() int {
	return 255
}
