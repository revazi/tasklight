package doctor

import (
	"bytes"
	"strings"
	"testing"

	"github.com/revazi/tasklight/internal/notify"
	"github.com/revazi/tasklight/internal/session"
)

func TestWriteFocusReportIncludesCoreDetectionData(t *testing.T) {
	report := FocusReport{
		Platform:  "darwin/arm64",
		GoVersion: "go1.test",
		Target: session.FocusTarget{
			ActivateApp: "com.googlecode.iterm2",
			Terminal: session.TerminalTarget{
				Name:            "iTerm2",
				BundleID:        "com.googlecode.iterm2",
				ITermSessionID:  "session-123",
				TerminalSession: "terminal-456",
				ClientTTY:       "/dev/ttys001",
			},
			Tmux: &session.TmuxTarget{
				Socket:      "/tmp/tmux/default",
				ClientName:  "/dev/ttys001",
				ClientTTY:   "/dev/ttys001",
				Session:     "project",
				WindowIndex: "2",
				WindowID:    "@10",
				PaneIndex:   "3",
				PaneID:      "%55",
			},
		},
		ClickCommand: "tmux select-pane -t %55",
		Notification: notify.FocusDiagnostics{
			Provider:            "native macOS helper",
			ProviderPath:        "/opt/tasklight/Tasklight.app",
			NativeHelperPath:    "/opt/tasklight/Tasklight.app",
			SupportsClick:       true,
			ExecutionCommand:    "focus-command",
			ScriptPath:          "/tmp/tasklight/focus/action.sh",
			DebugEnabled:        true,
			NativeHelperLogPath: "/tmp/tasklight/native-helper.log",
			FocusLogPath:        "/tmp/tasklight/focus/focus.log",
		},
	}

	var output bytes.Buffer
	WriteFocusReport(&output, report)
	got := output.String()

	wantParts := []string{
		"Tasklight focus diagnostics",
		"selected provider",
		"native macOS helper",
		"terminal bundle ID",
		"com.googlecode.iterm2",
		"iTerm session ID",
		"session-123",
		"terminal/client tty",
		"/dev/ttys001",
		"tmux socket",
		"/tmp/tmux/default",
		"tmux session",
		"project",
		"tmux window ID",
		"@10",
		"tmux pane ID",
		"%55",
		"click command",
		"select-pane",
		"provider command",
		"focus-command",
		"generated script path",
		"/tmp/tasklight/focus/action.sh",
		"native helper log",
		"/tmp/tasklight/native-helper.log",
		"focus execution log",
		"/tmp/tasklight/focus/focus.log",
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("focus report missing %q:\n%s", want, got)
		}
	}
}

func TestWriteFocusReportShowsMissingData(t *testing.T) {
	var output bytes.Buffer
	WriteFocusReport(&output, FocusReport{})
	got := output.String()

	for _, want := range []string{"inside tmux", "no", "click actions", "(not detected)", "enable logs"} {
		if !strings.Contains(got, want) {
			t.Fatalf("focus report missing %q:\n%s", want, got)
		}
	}
}
