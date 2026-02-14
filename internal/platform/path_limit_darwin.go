//go:build darwin

package platform

func maxPathLength() int {
	return 1024
}
