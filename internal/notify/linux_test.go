//go:build linux
// +build linux

package notify

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLinuxNotifierUsesNotifySend(t *testing.T) {
	var gotName string
	var gotArgs []string

	notifier := LinuxNotifier{
		getenv: func(string) string { return "" },
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
		getenv:   func(string) string { return "" },
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

func TestLinuxNotifierStartsOptInTmuxAction(t *testing.T) {
	var gotName string
	var gotArgs []string
	notifier := LinuxNotifier{
		getenv:   func(name string) string { return "1" },
		lookPath: func(string) (string, error) { return "/usr/bin/notify-send", nil },
		output: func(name string, args ...string) ([]byte, error) {
			return []byte("--action=ACTION --wait"), nil
		},
		start: func(name string, args ...string) error {
			gotName = name
			gotArgs = append([]string(nil), args...)
			return nil
		},
		run: func(string, ...string) error {
			t.Fatal("synchronous notify-send should not run for an action notification")
			return nil
		},
	}
	notification := Notification{
		Title:        "Tasklight",
		Message:      "done",
		ClickCommand: "tmux select-pane -t %5",
	}

	if err := notifier.Notify(notification); err != nil {
		t.Fatalf("Notify() error = %v, want nil", err)
	}
	if gotName != "/bin/sh" {
		t.Fatalf("command name = %q, want /bin/sh", gotName)
	}
	assertContainsArg(t, gotArgs, "tasklight-linux-action")
	assertContainsArg(t, gotArgs, notification.ClickCommand)
	assertContainsArg(t, gotArgs, "/usr/bin/notify-send")
	assertContainsArg(t, gotArgs, "--action=tasklight-return=Select tmux target")
	assertContainsArg(t, gotArgs, "--wait")
}

func TestStartLinuxActionRunsSelectedClickCommand(t *testing.T) {
	dir := t.TempDir()
	notifySend := filepath.Join(dir, "notify-send")
	if err := os.WriteFile(notifySend, []byte("#!/bin/sh\nprintf tasklight-return\n"), 0o700); err != nil {
		t.Fatalf("WriteFile(notify-send): %v", err)
	}
	clicked := filepath.Join(dir, "clicked")
	clickCommand := fmt.Sprintf("printf selected > %q", clicked)

	err := startLinuxAction(
		func(name string, args ...string) error { return exec.Command(name, args...).Run() },
		notifySend,
		[]string{"Tasklight", "done"},
		clickCommand,
	)
	if err != nil {
		t.Fatalf("startLinuxAction() error = %v", err)
	}
	content, err := os.ReadFile(clicked)
	if err != nil {
		t.Fatalf("ReadFile(clicked): %v", err)
	}
	if string(content) != "selected" {
		t.Fatalf("clicked content = %q, want selected", content)
	}
}

func TestLinuxNotifierFallsBackWhenActionsUnavailable(t *testing.T) {
	runCalled := false
	notifier := LinuxNotifier{
		getenv:   func(string) string { return "true" },
		lookPath: func(string) (string, error) { return "/usr/bin/notify-send", nil },
		output:   func(string, ...string) ([]byte, error) { return []byte("no actions"), nil },
		run: func(string, ...string) error {
			runCalled = true
			return nil
		},
		start: func(string, ...string) error {
			t.Fatal("action helper should not start without notify-send action support")
			return nil
		},
	}

	if err := notifier.Notify(Notification{Message: "done", ClickCommand: "tmux select-pane -t %5"}); err != nil {
		t.Fatalf("Notify() error = %v, want nil", err)
	}
	if !runCalled {
		t.Fatal("regular notify-send was not used as fallback")
	}
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
