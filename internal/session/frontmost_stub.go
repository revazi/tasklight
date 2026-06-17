//go:build !darwin
// +build !darwin

package session

func detectFrontmostBundleID() string {
	return ""
}
