//go:build linux
// +build linux

package notify

import (
	"errors"
	"testing"
)

func TestLinuxNotifierUsesNotifySend(t *testing.T) {
	var gotName string
	var gotArgs []string

	notifier := LinuxNotifier{
		lookPath: func(file string) (string, error) {
			if file != "notify-send" {
				t.Fatalf("lookPath(%q), want notify-send", file)
			}
			return "/usr/bin/notify-send", nil
		},
		run: func(name string, args ...string) error {
			gotName = name
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	notification := Notification{
		Title:    "Tasklight",
		Subtitle: "✅ tests finished in 2s",
		Message:  "Exit code: 0",
		IconPath: "/tmp/tasklight-icon.png",
		Sound:    true,
	}

	if err := notifier.Notify(notification); err != nil {
		t.Fatalf("Notify() error = %v, want nil", err)
	}

	if gotName != "/usr/bin/notify-send" {
		t.Fatalf("command name = %q, want /usr/bin/notify-send", gotName)
	}
	assertContainsArgPair(t, gotArgs, "-a", "Tasklight")
	assertContainsArgPair(t, gotArgs, "-h", "string:sound-name:message-new-instant")
	assertContainsArgPair(t, gotArgs, "-i", notification.IconPath)
	assertContainsArg(t, gotArgs, "Tasklight")
	assertContainsArg(t, gotArgs, "✅ tests finished in 2s\nExit code: 0")
}

func TestLinuxNotifierDefaultsTitle(t *testing.T) {
	var gotArgs []string

	notifier := LinuxNotifier{
		lookPath: func(string) (string, error) { return "notify-send", nil },
		run: func(_ string, args ...string) error {
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	if err := notifier.Notify(Notification{Subtitle: "done"}); err != nil {
		t.Fatalf("Notify() error = %v, want nil", err)
	}

	assertContainsArg(t, gotArgs, "Tasklight")
}

func TestLinuxNotifierReturnsLookPathError(t *testing.T) {
	wantErr := errors.New("not found")
	notifier := LinuxNotifier{
		lookPath: func(string) (string, error) { return "", wantErr },
	}

	if err := notifier.Notify(Notification{}); !errors.Is(err, wantErr) {
		t.Fatalf("Notify() error = %v, want %v", err, wantErr)
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

func assertContainsArg(t *testing.T, args []string, value string) {
	t.Helper()
	for _, arg := range args {
		if arg == value {
			return
		}
	}
	t.Fatalf("args %#v do not contain %q", args, value)
}
