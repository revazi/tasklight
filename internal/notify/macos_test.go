//go:build darwin
// +build darwin

package notify

import (
	"errors"
	"testing"
)

func TestMacOSNotifierUsesTerminalNotifierWhenAvailable(t *testing.T) {
	var gotName string
	var gotArgs []string

	notifier := MacOSNotifier{
		lookPath: func(file string) (string, error) {
			if file != "terminal-notifier" {
				t.Fatalf("lookPath(%q), want terminal-notifier", file)
			}
			return "/opt/homebrew/bin/terminal-notifier", nil
		},
		run: func(name string, args ...string) error {
			gotName = name
			gotArgs = append([]string(nil), args...)
			return nil
		},
		senderBundleID: func() string { return "dev.tasklight.Tasklight" },
	}

	notification := Notification{
		Title:        "Custom Title",
		Subtitle:     "subtitle with \"quotes\"",
		Message:      "line one\nline two",
		ActivateApp:  "com.apple.Terminal",
		ClickCommand: "echo clicked",
		IconPath:     "/tmp/tasklight-icon.png",
		Sound:        true,
	}

	if err := notifier.Notify(notification); err != nil {
		t.Fatalf("Notify() error = %v, want nil", err)
	}

	if gotName != "/opt/homebrew/bin/terminal-notifier" {
		t.Fatalf("command name = %q, want terminal-notifier path", gotName)
	}
	assertContainsArgPair(t, gotArgs, "-title", notification.Title)
	assertContainsArgPair(t, gotArgs, "-subtitle", notification.Subtitle)
	assertContainsArgPair(t, gotArgs, "-message", notification.Message)
	assertContainsArgPair(t, gotArgs, "-activate", notification.ActivateApp)
	assertContainsArgPair(t, gotArgs, "-sender", "dev.tasklight.Tasklight")
	assertNotContainsArg(t, gotArgs, "-appIcon")
	assertContainsArgPair(t, gotArgs, "-execute", notification.ClickCommand)
	assertContainsArgPair(t, gotArgs, "-sound", "default")
}

func TestMacOSNotifierUsesClickMessageForEmptyTerminalNotifierMessage(t *testing.T) {
	var gotArgs []string

	notifier := MacOSNotifier{
		lookPath: func(string) (string, error) { return "terminal-notifier", nil },
		run: func(_ string, args ...string) error {
			gotArgs = append([]string(nil), args...)
			return nil
		},
		senderBundleID: func() string { return "" },
	}

	if err := notifier.Notify(Notification{ClickCommand: "echo clicked"}); err != nil {
		t.Fatalf("Notify() error = %v, want nil", err)
	}

	assertContainsArgPair(t, gotArgs, "-message", "Click to return")
}

func TestMacOSNotifierUsesAppIconWhenNoSenderBundleID(t *testing.T) {
	var gotArgs []string

	notifier := MacOSNotifier{
		lookPath: func(string) (string, error) { return "terminal-notifier", nil },
		run: func(_ string, args ...string) error {
			gotArgs = append([]string(nil), args...)
			return nil
		},
		senderBundleID: func() string { return "" },
	}

	if err := notifier.Notify(Notification{Message: "hello", IconPath: "/tmp/tasklight-icon.png"}); err != nil {
		t.Fatalf("Notify() error = %v, want nil", err)
	}

	assertContainsArgPair(t, gotArgs, "-appIcon", "file:///tmp/tasklight-icon.png")
}

func TestMacOSNotifierFallsBackToAppleScript(t *testing.T) {
	var gotName string
	var gotArgs []string

	notifier := MacOSNotifier{
		lookPath: func(string) (string, error) { return "", errors.New("not found") },
		run: func(name string, args ...string) error {
			gotName = name
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	notification := Notification{
		Title:    "Custom Title",
		Subtitle: "subtitle with \"quotes\"",
		Message:  "line one\nline two",
		IconPath: "/tmp/tasklight-icon.png",
	}

	if err := notifier.Notify(notification); err != nil {
		t.Fatalf("Notify() error = %v, want nil", err)
	}

	if gotName != "osascript" {
		t.Fatalf("command name = %q, want osascript", gotName)
	}
	if len(gotArgs) != 5 {
		t.Fatalf("args len = %d, want 5: %#v", len(gotArgs), gotArgs)
	}
	if gotArgs[0] != "-e" {
		t.Fatalf("first arg = %q, want -e", gotArgs[0])
	}
	if gotArgs[2] != notification.Message {
		t.Fatalf("message arg = %q, want %q", gotArgs[2], notification.Message)
	}
	if gotArgs[3] != notification.Title {
		t.Fatalf("title arg = %q, want %q", gotArgs[3], notification.Title)
	}
	if gotArgs[4] != notification.Subtitle {
		t.Fatalf("subtitle arg = %q, want %q", gotArgs[4], notification.Subtitle)
	}
	for _, arg := range gotArgs {
		if arg == "-appIcon" || arg == notification.IconPath {
			t.Fatalf("osascript fallback args %#v should not include icon args", gotArgs)
		}
	}
}

func TestMacOSNotifierDefaultsTitle(t *testing.T) {
	var gotArgs []string

	notifier := MacOSNotifier{
		lookPath: func(string) (string, error) { return "", errors.New("not found") },
		run: func(_ string, args ...string) error {
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	if err := notifier.Notify(Notification{Message: "hello"}); err != nil {
		t.Fatalf("Notify() error = %v, want nil", err)
	}

	if got := gotArgs[3]; got != "Tasklight" {
		t.Fatalf("default title = %q, want Tasklight", got)
	}
}

func assertContainsArgPair(t *testing.T, args []string, flag string, value string) {
	t.Helper()
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == value {
			return
		}
	}
	t.Fatalf("args %#v do not contain %s %q", args, flag, value)
}

func assertNotContainsArg(t *testing.T, args []string, value string) {
	t.Helper()
	for _, arg := range args {
		if arg == value {
			t.Fatalf("args %#v should not contain %q", args, value)
		}
	}
}
