package session

import (
	"strings"
	"testing"
)

func TestNormalizeActivateApp(t *testing.T) {
	tests := map[string]string{
		"Terminal":           "com.apple.Terminal",
		"iTerm2":             "com.googlecode.iterm2",
		"WezTerm":            "com.github.wez.wezterm",
		"VS Code":            "com.microsoft.VSCode",
		"Cursor":             "com.todesktop.230313mzl4w4u92",
		"com.example.Custom": "com.example.Custom",
	}

	for input, want := range tests {
		if got := NormalizeActivateApp(input); got != want {
			t.Fatalf("NormalizeActivateApp(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestFocusTargetClickCommandIncludesTmuxAndActivateApp(t *testing.T) {
	target := FocusTarget{
		ActivateApp: "com.apple.Terminal",
		Tmux: &TmuxTarget{
			Socket:   "/tmp/tmux-501/default",
			WindowID: "@10",
			PaneID:   "%55",
		},
	}

	command := target.ClickCommand()

	wantParts := []string{
		"tmux",
		"'-S' '/tmp/tmux-501/default'",
		"'select-window' '-t' '@10'",
		"'select-pane' '-t' '%55'",
		"osascript",
		`'tell application id "com.apple.Terminal" to activate'`,
	}
	for _, want := range wantParts {
		if !strings.Contains(command, want) {
			t.Fatalf("ClickCommand() = %q, want to contain %q", command, want)
		}
	}
}

func TestFocusTargetClickCommandEscapesShellValues(t *testing.T) {
	target := TmuxTarget{
		Session:     "project's tests",
		WindowIndex: "1",
		PaneID:      "%55",
	}

	command := target.ClickCommand()

	if !strings.Contains(command, "'project'\\''s tests:1'") {
		t.Fatalf("ClickCommand() = %q, want escaped session", command)
	}
}

func TestParseTmuxSocket(t *testing.T) {
	input := "/private/tmp/tmux-501/default,16652,7"
	want := "/private/tmp/tmux-501/default"

	if got := parseTmuxSocket(input); got != want {
		t.Fatalf("parseTmuxSocket(%q) = %q, want %q", input, got, want)
	}
}
