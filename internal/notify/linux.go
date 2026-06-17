//go:build linux
// +build linux

package notify

import (
	"os/exec"
	"strings"

	brandassets "github.com/revazi/tasklight/assets/brand"
)

type linuxCommandRunner func(name string, args ...string) error
type linuxPathLookup func(file string) (string, error)

// LinuxNotifier sends desktop notifications using notify-send.
type LinuxNotifier struct {
	run      linuxCommandRunner
	lookPath linuxPathLookup
}

func DefaultNotifier() Notifier {
	return LinuxNotifier{}
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
	args = append(args, title)

	body := linuxBody(notification)
	if body != "" {
		args = append(args, body)
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
