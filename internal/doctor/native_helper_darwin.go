//go:build darwin
// +build darwin

package doctor

import (
	"os"
	"path/filepath"
	"strings"
)

func macOSNativeHelperPath() string {
	if override := os.Getenv("TASKLIGHT_MACOS_HELPER"); override != "" {
		if strings.EqualFold(override, "none") || override == "-" {
			return ""
		}
		if strings.HasSuffix(override, ".app") && doctorAppBundleExists(override) {
			return override
		}
		if doctorFileExists(override) {
			return override
		}
		return ""
	}

	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	executableDir := filepath.Dir(executable)
	for _, candidate := range []string{
		filepath.Join(executableDir, "Tasklight.app"),
		filepath.Join(executableDir, "..", "Tasklight.app"),
	} {
		if doctorAppBundleExists(candidate) {
			return candidate
		}
	}
	return ""
}

func doctorAppBundleExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	return doctorFileExists(filepath.Join(path, "Contents", "MacOS", "TasklightNotifier"))
}

func doctorFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
