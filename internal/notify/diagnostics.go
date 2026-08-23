package notify

import (
	"os"
	"path/filepath"
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
		DebugEnabled: os.Getenv("TASKLIGHT_FOCUS_DEBUG") != "",
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
