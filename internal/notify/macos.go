//go:build darwin
// +build darwin

package notify

import (
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	brandassets "github.com/revazi/tasklight/assets/brand"
)

const displayNotificationScript = `on run argv
	set notificationMessage to item 1 of argv
	set notificationTitle to item 2 of argv
	set notificationSubtitle to item 3 of argv

	if notificationSubtitle is "" then
		display notification notificationMessage with title notificationTitle
	else
		display notification notificationMessage with title notificationTitle subtitle notificationSubtitle
	end if
end run`

const tasklightSenderBundleID = "dev.tasklight.Tasklight"

type commandRunner func(name string, args ...string) error
type pathLookup func(file string) (string, error)
type senderBundleIDProvider func() string

// MacOSNotifier sends macOS notifications.
//
// It prefers terminal-notifier when installed because it can run an action when
// the notification is clicked. It falls back to osascript for dependency-free
// notifications without click actions.
type MacOSNotifier struct {
	run            commandRunner
	lookPath       pathLookup
	senderBundleID senderBundleIDProvider
}

func DefaultNotifier() Notifier {
	return MacOSNotifier{}
}

func (n MacOSNotifier) Notify(notification Notification) error {
	title := notification.Title
	if title == "" {
		title = "Tasklight"
	}

	run := n.run
	if run == nil {
		run = runCommand
	}

	lookPath := n.lookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	if terminalNotifierPath, err := lookPath("terminal-notifier"); err == nil {
		senderBundleID := defaultTerminalNotifierSenderBundleID
		if n.senderBundleID != nil {
			senderBundleID = n.senderBundleID
		}
		return run(terminalNotifierPath, terminalNotifierArgs(notification, title, senderBundleID())...)
	}

	return run(
		"osascript",
		"-e",
		displayNotificationScript,
		notification.Message,
		title,
		notification.Subtitle,
	)
}

func terminalNotifierArgs(notification Notification, title string, senderBundleID string) []string {
	message := notification.Message
	if message == "" {
		if notification.ClickCommand != "" || notification.ActivateApp != "" {
			message = "Click to return"
		} else {
			message = "Done"
		}
	}

	args := []string{"-title", title, "-message", message}
	if notification.Subtitle != "" {
		args = append(args, "-subtitle", notification.Subtitle)
	}
	if notification.ActivateApp != "" && looksLikeBundleID(notification.ActivateApp) {
		args = append(args, "-activate", notification.ActivateApp)
	}
	if senderBundleID != "" {
		args = append(args, "-sender", senderBundleID)
	} else if iconURL := terminalNotifierIconURL(notification); iconURL != "" {
		args = append(args, "-appIcon", iconURL)
	}
	if notification.ClickCommand != "" {
		args = append(args, "-execute", notification.ClickCommand)
	}
	if notification.Sound {
		args = append(args, "-sound", "default")
	}

	return args
}

var tasklightSenderRegistration struct {
	once sync.Once
	id   string
}

func defaultTerminalNotifierSenderBundleID() string {
	if override := os.Getenv("TASKLIGHT_SENDER_BUNDLE_ID"); override != "" {
		if strings.EqualFold(override, "none") || override == "-" {
			return ""
		}
		return override
	}

	tasklightSenderRegistration.once.Do(func() {
		appPath := brandassets.DefaultMacOSAppBundle(tasklightSenderBundleID)
		if appPath == "" {
			return
		}
		if err := registerMacOSAppBundle(appPath); err != nil {
			return
		}
		tasklightSenderRegistration.id = tasklightSenderBundleID
	})

	return tasklightSenderRegistration.id
}

func registerMacOSAppBundle(appPath string) error {
	lsregister := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
	return exec.Command(lsregister, "-f", appPath).Run()
}

func terminalNotifierIconURL(notification Notification) string {
	iconPath := notification.IconPath
	if iconPath == "" {
		iconPath = brandassets.DefaultIconPath()
	}
	if iconPath == "" {
		return ""
	}
	if strings.Contains(iconPath, "://") {
		return iconPath
	}
	absolutePath, err := filepath.Abs(iconPath)
	if err != nil {
		absolutePath = iconPath
	}
	return (&url.URL{Scheme: "file", Path: absolutePath}).String()
}

func looksLikeBundleID(value string) bool {
	for _, r := range value {
		if r == '.' {
			return true
		}
	}
	return false
}

func runCommand(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}
