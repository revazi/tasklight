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

func TestFocusTargetClickCommandIncludesTmuxSwitchClientAndItermFocus(t *testing.T) {
	target := FocusTarget{
		ActivateApp: "com.googlecode.iterm2",
		Terminal: TerminalTarget{
			ITermSessionID: "59571519-5F27-46C1-95C1-2A6E9A2286F0",
			ClientTTY:      "/dev/ttys000",
		},
		Tmux: &TmuxTarget{
			Socket:     "/tmp/tmux-501/default",
			ClientName: "/dev/ttys000",
			ClientTTY:  "/dev/ttys000",
			Session:    "project",
			WindowID:   "@10",
			PaneID:     "%55",
		},
	}

	command := target.ClickCommand()

	wantParts := []string{
		"tmux",
		"'-S' '/tmp/tmux-501/default'",
		"'switch-client' '-c' '/dev/ttys000' '-t' 'project'",
		"'select-window' '-t' '@10'",
		"'select-pane' '-t' '%55'",
		"osascript",
		"com.googlecode.iterm2",
		"59571519-5F27-46C1-95C1-2A6E9A2286F0",
		"/dev/ttys000",
		"select aWindow",
		"select aTab",
		"select aSession",
	}
	for _, want := range wantParts {
		if !strings.Contains(command, want) {
			t.Fatalf("ClickCommand() = %q, want to contain %q", command, want)
		}
	}
}

func TestFocusTargetClickCommandFallsBackToOpenForTerminal(t *testing.T) {
	target := FocusTarget{ActivateApp: "com.apple.Terminal"}

	command := target.ClickCommand()

	if !strings.Contains(command, "'open' '-b' 'com.apple.Terminal'") {
		t.Fatalf("ClickCommand() = %q, want open fallback", command)
	}
}

func TestTerminalFocusCommandUsesClientTTY(t *testing.T) {
	target := FocusTarget{
		ActivateApp: "com.apple.Terminal",
		Terminal:    TerminalTarget{ClientTTY: "/dev/ttys123"},
	}

	command := target.ClickCommand()

	for _, want := range []string{"com.apple.Terminal", "/dev/ttys123", "set selected of aTab to true", "set index of aWindow to 1"} {
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

func TestNormalizeITermSessionID(t *testing.T) {
	input := "w0t0p0:59571519-5F27-46C1-95C1-2A6E9A2286F0"
	want := "59571519-5F27-46C1-95C1-2A6E9A2286F0"

	if got := normalizeITermSessionID(input); got != want {
		t.Fatalf("normalizeITermSessionID(%q) = %q, want %q", input, got, want)
	}
}
