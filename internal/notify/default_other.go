//go:build !darwin && !linux
// +build !darwin,!linux

package notify

func DefaultNotifier() Notifier {
	return NoopNotifier{}
}

// DiagnoseFocus reports that the current platform has no notification provider.
func DiagnoseFocus(_ string) FocusDiagnostics {
	diagnostics := newFocusDiagnostics()
	diagnostics.Provider = "unavailable"
	return diagnostics
}
