package cli

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"tasklight/internal/notify"
	"tasklight/internal/session"
)

func TestMain(m *testing.M) {
	detectFocusTarget = func(session.DetectOptions) session.FocusTarget {
		return session.FocusTarget{}
	}
	os.Exit(m.Run())
}

func TestExecuteRunSuccess(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	notifier := &recordingNotifier{}

	code := ExecuteWithNotifier([]string{"run", "--", "/bin/sh", "-c", "printf ok"}, strings.NewReader(""), &stdout, &stderr, notifier)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if got := stdout.String(); got != "ok" {
		t.Fatalf("stdout = %q, want %q", got, "ok")
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}

	assertNotificationCount(t, notifier, 1)
	got := notifier.notifications[0]
	if got.Title != "Tasklight" {
		t.Fatalf("notification title = %q, want Tasklight", got.Title)
	}
	if !strings.Contains(got.Subtitle, "✅ sh finished") {
		t.Fatalf("notification subtitle = %q, want success", got.Subtitle)
	}
	if got.Message != "" {
		t.Fatalf("notification message = %q, want empty", got.Message)
	}
}

func TestExecuteRunPreservesFailureExitCode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	notifier := &recordingNotifier{}

	code := ExecuteWithNotifier([]string{"run", "--", "/bin/sh", "-c", "exit 42"}, strings.NewReader(""), &stdout, &stderr, notifier)

	if code != 42 {
		t.Fatalf("code = %d, want 42; stderr=%q", code, stderr.String())
	}

	assertNotificationCount(t, notifier, 1)
	got := notifier.notifications[0]
	if !strings.Contains(got.Subtitle, "❌ sh failed") {
		t.Fatalf("notification subtitle = %q, want failure", got.Subtitle)
	}
	if got.Message != "Exit code: 42" {
		t.Fatalf("notification message = %q, want exit code", got.Message)
	}
}

func TestExecuteRunUsesCustomNameInNotification(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	notifier := &recordingNotifier{}

	code := ExecuteWithNotifier([]string{"run", "--name", "Demo task", "--", "/bin/sh", "-c", "exit 0"}, strings.NewReader(""), &stdout, &stderr, notifier)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}

	assertNotificationCount(t, notifier, 1)
	if got := notifier.notifications[0].Subtitle; !strings.Contains(got, "Demo task") {
		t.Fatalf("notification subtitle = %q, want custom name", got)
	}
}

func TestExecuteRunAddsFocusTargetToNotification(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	notifier := &recordingNotifier{}

	oldDetectFocusTarget := detectFocusTarget
	detectFocusTarget = func(opts session.DetectOptions) session.FocusTarget {
		if opts.ActivateApp != "Terminal" {
			t.Fatalf("ActivateApp = %q, want Terminal", opts.ActivateApp)
		}
		return session.FocusTarget{ActivateApp: "com.apple.Terminal"}
	}
	t.Cleanup(func() { detectFocusTarget = oldDetectFocusTarget })

	code := ExecuteWithNotifier([]string{"run", "--activate-app", "Terminal", "--", "/bin/sh", "-c", "exit 0"}, strings.NewReader(""), &stdout, &stderr, notifier)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}

	assertNotificationCount(t, notifier, 1)
	got := notifier.notifications[0]
	if got.ActivateApp != "com.apple.Terminal" {
		t.Fatalf("ActivateApp = %q, want com.apple.Terminal", got.ActivateApp)
	}
	if got.ClickCommand == "" {
		t.Fatal("ClickCommand is empty, want focus command")
	}
}

func TestExecuteRunNotificationFailureIsWarningOnly(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	notifier := &recordingNotifier{err: errors.New("boom")}

	code := ExecuteWithNotifier([]string{"run", "--", "/bin/sh", "-c", "exit 42"}, strings.NewReader(""), &stdout, &stderr, notifier)

	if code != 42 {
		t.Fatalf("code = %d, want 42", code)
	}
	if !strings.Contains(stderr.String(), "tasklight: warning: notification failed: boom") {
		t.Fatalf("stderr = %q, want notification warning", stderr.String())
	}
}

func TestExecuteRunMissingCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	notifier := &recordingNotifier{}

	code := ExecuteWithNotifier([]string{"run"}, strings.NewReader(""), &stdout, &stderr, notifier)

	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "missing command") {
		t.Fatalf("stderr = %q, want missing command", stderr.String())
	}
	assertNotificationCount(t, notifier, 0)
}

func TestExecuteRunRequiresSeparator(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	notifier := &recordingNotifier{}

	code := ExecuteWithNotifier([]string{"run", "/bin/sh", "-c", "exit 0"}, strings.NewReader(""), &stdout, &stderr, notifier)

	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "missing -- before command") {
		t.Fatalf("stderr = %q, want missing -- message", stderr.String())
	}
	assertNotificationCount(t, notifier, 0)
}

func TestExecuteNotify(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	notifier := &recordingNotifier{}

	oldDetectFocusTarget := detectFocusTarget
	detectFocusTarget = func(opts session.DetectOptions) session.FocusTarget {
		if opts.ActivateApp != "iTerm2" {
			t.Fatalf("ActivateApp = %q, want iTerm2", opts.ActivateApp)
		}
		return session.FocusTarget{ActivateApp: "com.googlecode.iterm2"}
	}
	t.Cleanup(func() { detectFocusTarget = oldDetectFocusTarget })

	code := ExecuteWithNotifier([]string{"notify", "--title", "Pi", "--subtitle", "✅ Ready", "--message", "Finished task", "--activate-app", "iTerm2", "--sound"}, strings.NewReader(""), &stdout, &stderr, notifier)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}
	assertNotificationCount(t, notifier, 1)
	got := notifier.notifications[0]
	if got.Title != "Pi" {
		t.Fatalf("Title = %q, want Pi", got.Title)
	}
	if got.Subtitle != "✅ Ready" {
		t.Fatalf("Subtitle = %q, want ready", got.Subtitle)
	}
	if got.Message != "Finished task" {
		t.Fatalf("Message = %q, want task message", got.Message)
	}
	if got.ActivateApp != "com.googlecode.iterm2" {
		t.Fatalf("ActivateApp = %q, want iTerm bundle", got.ActivateApp)
	}
	if !got.Sound {
		t.Fatal("Sound = false, want true")
	}
}

func TestExecuteNotifyRequiresText(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	notifier := &recordingNotifier{}

	code := ExecuteWithNotifier([]string{"notify"}, strings.NewReader(""), &stdout, &stderr, notifier)

	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "missing notification text") {
		t.Fatalf("stderr = %q, want missing notification text", stderr.String())
	}
	assertNotificationCount(t, notifier, 0)
}

func TestExecuteNotifyFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	notifier := &recordingNotifier{err: errors.New("boom")}

	code := ExecuteWithNotifier([]string{"notify", "--message", "hello"}, strings.NewReader(""), &stdout, &stderr, notifier)

	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "tasklight notify: notification failed: boom") {
		t.Fatalf("stderr = %q, want notification error", stderr.String())
	}
}

func TestExecuteRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	notifier := &recordingNotifier{}

	code := ExecuteWithNotifier([]string{"run", "--help"}, strings.NewReader(""), &stdout, &stderr, notifier)

	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "tasklight run [options] -- <command>") {
		t.Fatalf("stdout = %q, want run help", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertNotificationCount(t, notifier, 0)
}

func TestFormatDuration(t *testing.T) {
	tests := map[string]string{
		"0s":     "0s",
		"1ms":    "<1s",
		"1s":     "1s",
		"61s":    "1m 1s",
		"1h2m3s": "1h 2m 3s",
	}

	for input, want := range tests {
		got := formatDuration(mustParseDuration(t, input))
		if got != want {
			t.Fatalf("formatDuration(%s) = %q, want %q", input, got, want)
		}
	}
}

func mustParseDuration(t *testing.T, value string) time.Duration {
	t.Helper()
	duration, err := time.ParseDuration(value)
	if err != nil {
		t.Fatalf("ParseDuration(%q): %v", value, err)
	}
	return duration
}

type recordingNotifier struct {
	notifications []notify.Notification
	err           error
}

func (n *recordingNotifier) Notify(notification notify.Notification) error {
	n.notifications = append(n.notifications, notification)
	return n.err
}

func assertNotificationCount(t *testing.T, notifier *recordingNotifier, want int) {
	t.Helper()
	if got := len(notifier.notifications); got != want {
		t.Fatalf("notification count = %d, want %d: %#v", got, want, notifier.notifications)
	}
}
