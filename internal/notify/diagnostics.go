package notify

import (
	"os"
	"path/filepath"
	"strings"
)

// FocusDiagnostics describes how the selected notification provider will handle
// a generated click-to-focus command.
type FocusDiagnostics struct {
	Provider            string
	ProviderPath        string
	NativeHelperPath    string
	SupportsClick       bool
	ExecutionCommand    string
	ScriptPath          string
	DebugEnabled        bool
	NativeHelperLogPath string
	FocusLogPath        string
}

func newFocusDiagnostics() FocusDiagnostics {
	return FocusDiagnostics{
		DebugEnabled: environmentFlagEnabled(os.Getenv("TASKLIGHT_FOCUS_DEBUG")),
	}
}

func environmentFlagEnabled(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func focusLogPaths() (nativeHelper string, focusExecution string) {
	cacheDir := tasklightCacheDir()
	return filepath.Join(cacheDir, "native-helper.log"), filepath.Join(cacheDir, "focus", "focus.log")
}

func tasklightCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil || cacheDir == "" {
		cacheDir = os.TempDir()
	}
	return filepath.Join(cacheDir, "tasklight")
}
