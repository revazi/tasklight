//go:build linux
// +build linux

package notify

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	brandassets "github.com/revazi/tasklight/assets/brand"
)

type linuxCommandRunner func(name string, args ...string) error
type linuxCommandStarter func(name string, args ...string) error
type linuxCommandOutput func(name string, args ...string) ([]byte, error)
type linuxPathLookup func(file string) (string, error)

// LinuxNotifier sends desktop notifications using notify-send.
type LinuxNotifier struct {
	run      linuxCommandRunner
	start    linuxCommandStarter
	output   linuxCommandOutput
	lookPath linuxPathLookup
	getenv   func(string) string
}

func DefaultNotifier() Notifier {
	return LinuxNotifier{}
}

// DiagnoseFocus reports the Linux notification provider. Basic notify-send
// notifications do not currently expose a portable click action.
func DiagnoseFocus(clickCommand string) FocusDiagnostics {
	diagnostics := newFocusDiagnostics()
	diagnostics.Provider = "notify-send"
	if path, err := exec.LookPath("notify-send"); err == nil {
		diagnostics.ProviderPath = path
		if clickCommand != "" && linuxActionsEnabled(nil) {
			if help, err := runLinuxCommandOutput(path, "--help"); err == nil && strings.Contains(string(help), "--action") {
				diagnostics.SupportsClick = true
				diagnostics.ExecutionCommand = clickCommand
			}
		}
	}
	return diagnostics
}

func (n LinuxNotifier) Notify(notification Notification) error {
	title := notification.Title
	if title == "" {
		title = "Tasklight"
	}

	run := n.run
	if run == nil {
		run = runLinuxCommand
	}

	lookPath := n.lookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	notifySendPath, err := lookPath("notify-send")
	if err != nil {
		return err
	}

	args := []string{"-a", "Tasklight"}
	if notification.Sound {
		args = append(args, "-h", "string:sound-name:message-new-instant")
	}
	if iconPath := linuxNotificationIconPath(notification); iconPath != "" {
		args = append(args, "-i", iconPath)
	}
	useAction := false
	if notification.ClickCommand != "" && linuxActionsEnabled(n.getenv) {
		output := n.output
		if output == nil {
			output = runLinuxCommandOutput
		}
		if help, err := output(notifySendPath, "--help"); err == nil && strings.Contains(string(help), "--action") {
			useAction = true
			args = append(args, "--action=tasklight-return=Select tmux target", "--wait")
		}
	}

	args = append(args, title)
	body := linuxBody(notification)
	if body != "" {
		args = append(args, body)
	}

	if useAction {
		start := n.start
		if start == nil {
			start = startLinuxCommand
		}
		return startLinuxAction(start, notifySendPath, args, notification.ClickCommand)
	}
	return run(notifySendPath, args...)
}

func linuxNotificationIconPath(notification Notification) string {
	if notification.IconPath != "" {
		return notification.IconPath
	}
	return brandassets.DefaultIconPath()
}

func linuxBody(notification Notification) string {
	parts := make([]string, 0, 2)
	if notification.Subtitle != "" {
		parts = append(parts, notification.Subtitle)
	}
	if notification.Message != "" {
		parts = append(parts, notification.Message)
	}
	return strings.Join(parts, "\n")
}

func runLinuxCommand(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

func runLinuxCommandOutput(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func linuxActionsEnabled(getenv func(string) string) bool {
	if getenv == nil {
		getenv = os.Getenv
	}
	switch strings.ToLower(strings.TrimSpace(getenv("TASKLIGHT_LINUX_ACTIONS"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func startLinuxAction(start linuxCommandStarter, notifySendPath string, args []string, clickCommand string) error {
	const script = `click_command=$1
shift
choice="$("$@")"
if [ "$choice" = "tasklight-return" ]; then
	exec /bin/sh -c "$click_command"
fi`
	helperArgs := []string{"-c", script, "tasklight-linux-action", clickCommand, notifySendPath}
	helperArgs = append(helperArgs, args...)
	return start("/bin/sh", helperArgs...)
}

func startLinuxCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = nil
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
