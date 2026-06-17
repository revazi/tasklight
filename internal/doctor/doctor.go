package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	brandassets "tasklight/assets/brand"
)

const tasklightSenderBundleID = "dev.tasklight.Tasklight"

// Run writes environment diagnostics and returns a process exit code.
func Run(w io.Writer) int {
	failures := 0

	fmt.Fprintln(w, "Tasklight doctor")
	fmt.Fprintln(w)
	info(w, "Platform", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
	info(w, "Go", runtime.Version())
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Notifications")
	switch runtime.GOOS {
	case "darwin":
		if path, ok := lookPath("osascript"); ok {
			okLine(w, "osascript", path)
		} else {
			failures++
			failLine(w, "osascript", "missing; basic macOS notifications will not work")
		}

		if path, ok := lookPath("terminal-notifier"); ok {
			okLine(w, "terminal-notifier", path)
			checkMacOSSenderApp(w)
		} else {
			warnLine(w, "terminal-notifier", "optional but recommended: brew install terminal-notifier")
		}

	case "linux":
		if path, ok := lookPath("notify-send"); ok {
			okLine(w, "notify-send", path)
		} else {
			failures++
			failLine(w, "notify-send", "missing; install libnotify-bin/libnotify with your package manager")
		}

	default:
		failures++
		failLine(w, "notification provider", "unsupported platform; notifications will be disabled")
	}

	if iconPath := brandassets.DefaultIconPath(); iconPath != "" {
		okLine(w, "bundled icon", iconPath)
	} else {
		warnLine(w, "bundled icon", "could not prepare cached icon")
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Focus/session integration")
	if path, ok := lookPath("tmux"); ok {
		okLine(w, "tmux", path)
	} else {
		warnLine(w, "tmux", "not found; tmux pane return will be unavailable")
	}
	if pane := os.Getenv("TMUX_PANE"); pane != "" {
		okLine(w, "current tmux pane", pane)
	} else {
		warnLine(w, "current tmux pane", "not inside tmux")
	}

	fmt.Fprintln(w)
	if failures == 0 {
		okLine(w, "result", "Tasklight looks ready")
		return 0
	}
	failLine(w, "result", fmt.Sprintf("%d required check(s) failed", failures))
	return 1
}

func checkMacOSSenderApp(w io.Writer) {
	appPath := brandassets.DefaultMacOSAppBundle(tasklightSenderBundleID)
	if appPath == "" {
		warnLine(w, "Tasklight.app sender", "could not prepare local sender app bundle")
		return
	}
	okLine(w, "Tasklight.app sender", appPath)

	lsregister := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
	if _, err := os.Stat(lsregister); err != nil {
		warnLine(w, "LaunchServices registration", "lsregister not found")
		return
	}
	if err := exec.Command(lsregister, "-f", appPath).Run(); err != nil {
		warnLine(w, "LaunchServices registration", err.Error())
		return
	}
	okLine(w, "LaunchServices registration", tasklightSenderBundleID)
}

func lookPath(name string) (string, bool) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", false
	}
	return path, true
}

func info(w io.Writer, name string, message string) {
	fmt.Fprintf(w, "  • %-28s %s\n", name, message)
}

func okLine(w io.Writer, name string, message string) {
	fmt.Fprintf(w, "  ✓ %-28s %s\n", name, message)
}

func warnLine(w io.Writer, name string, message string) {
	fmt.Fprintf(w, "  ! %-28s %s\n", name, message)
}

func failLine(w io.Writer, name string, message string) {
	fmt.Fprintf(w, "  ✗ %-28s %s\n", name, strings.TrimSpace(message))
}
