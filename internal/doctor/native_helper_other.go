//go:build !darwin
// +build !darwin

package doctor

func macOSNativeHelperPath() string {
	return ""
}
