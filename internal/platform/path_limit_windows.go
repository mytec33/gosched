//go:build windows

package platform

func MaxPathLength() int {
	return 260
}
