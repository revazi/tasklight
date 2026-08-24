//go:build darwin
// +build darwin

package notify

import (
	"crypto/sha256"
	"encoding/hex"
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
type commandStarter func(name string, args ...string) error
type pathLookup func(file string) (string, error)
type senderBundleIDProvider func() string
type nativeHelperProvider func() string

// MacOSNotifier sends macOS notifications.
//
// It prefers terminal-notifier when installed because it can run an action when
// the notification is clicked. It falls back to osascript for dependency-free
// notifications without click actions.
type MacOSNotifier struct {
	run              commandRunner
	start            commandStarter
	lookPath         pathLookup
	senderBundleID   senderBundleIDProvider
	nativeHelperPath nativeHelperProvider
}

func DefaultNotifier() Notifier {
	return MacOSNotifier{}
}

// DiagnoseFocus reports the provider and execution path that would be used for
// a click-to-focus command without sending a notification.
func DiagnoseFocus(clickCommand string) FocusDiagnostics {
	return diagnoseFocus(clickCommand, defaultNativeHelperPath, exec.LookPath)
}

func diagnoseFocus(clickCommand string, nativeHelperPath nativeHelperProvider, lookPath pathLookup) FocusDiagnostics {
	diagnostics := newFocusDiagnostics()
	diagnostics.NativeHelperLogPath, diagnostics.FocusLogPath = focusLogPaths()
	diagnostics.NativeHelperPath = nativeHelperPath()

	if diagnostics.NativeHelperPath != "" {
		diagnostics.Provider = "native macOS helper"
		diagnostics.ProviderPath = diagnostics.NativeHelperPath
		diagnostics.SupportsClick = true
		diagnostics.ExecutionCommand = clickCommand
		return diagnostics
	}

	if terminalNotifierPath, err := lookPath("terminal-notifier"); err == nil {
		diagnostics.Provider = "terminal-notifier"
		diagnostics.ProviderPath = terminalNotifierPath
		diagnostics.SupportsClick = true
		diagnostics.ExecutionCommand = clickCommand
		if shouldWrapExecuteCommand(clickCommand) {
			diagnostics.ScriptPath = executeScriptPath(clickCommand)
			if diagnostics.ScriptPath != "" {
				diagnostics.ExecutionCommand = shellJoin([]string{"/bin/sh", diagnostics.ScriptPath})
			}
		}
		return diagnostics
	}

	diagnostics.Provider = "osascript fallback"
	if osascriptPath, err := lookPath("osascript"); err == nil {
		diagnostics.ProviderPath = osascriptPath
	}
	return diagnostics
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

	nativeHelperPath := defaultNativeHelperPath
	if n.nativeHelperPath != nil {
		nativeHelperPath = n.nativeHelperPath
	}
	if helperPath := nativeHelperPath(); helperPath != "" {
		start := n.start
		if start == nil {
			start = startCommand
		}
		return startNativeHelper(start, helperPath, nativeHelperArgs(notification, title))
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

func startNativeHelper(start commandStarter, helperPath string, args []string) error {
	if strings.HasSuffix(helperPath, ".app") {
		openArgs := []string{"-n", helperPath, "--args"}
		openArgs = append(openArgs, args...)
		return start("open", openArgs...)
	}
	return start(helperPath, args...)
}

func nativeHelperArgs(notification Notification, title string) []string {
	args := []string{"notify", "--title", title}
	if notification.Subtitle != "" {
		args = append(args, "--subtitle", notification.Subtitle)
	}
	if notification.Message != "" {
		args = append(args, "--message", notification.Message)
	}
	if notification.ClickCommand != "" {
		args = append(args, "--click-command", notification.ClickCommand)
	}
	if notification.Sound {
		args = append(args, "--sound")
	}
	return args
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
	if notification.ClickCommand == "" && notification.ActivateApp != "" && looksLikeBundleID(notification.ActivateApp) {
		args = append(args, "-activate", notification.ActivateApp)
	}
	if senderBundleID != "" && notification.ClickCommand == "" {
		args = append(args, "-sender", senderBundleID)
	} else if iconURL := terminalNotifierIconURL(notification); iconURL != "" {
		args = append(args, "-appIcon", iconURL)
	}
	if notification.ClickCommand != "" {
		args = append(args, "-execute", terminalNotifierExecuteCommand(notification.ClickCommand))
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

func defaultNativeHelperPath() string {
	if override := os.Getenv("TASKLIGHT_MACOS_HELPER"); override != "" {
		if strings.EqualFold(override, "none") || override == "-" {
			return ""
		}
		if strings.HasSuffix(override, ".app") && appBundleExists(override) {
			return override
		}
		if fileExists(override) {
			return override
		}
		return ""
	}

	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	executableDir := filepath.Dir(executable)
	candidates := []string{
		filepath.Join(executableDir, "Tasklight.app"),
		filepath.Join(executableDir, "..", "Tasklight.app"),
	}
	for _, candidate := range candidates {
		if appBundleExists(candidate) {
			return candidate
		}
	}
	return ""
}

func appBundleExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	return fileExists(filepath.Join(path, "Contents", "MacOS", "TasklightNotifier"))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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

func terminalNotifierExecuteCommand(command string) string {
	if command == "" {
		return ""
	}
	if !shouldWrapExecuteCommand(command) {
		return command
	}
	if scriptPath := writeExecuteScript(command); scriptPath != "" {
		return shellJoin([]string{"/bin/sh", scriptPath})
	}
	return command
}

func shouldWrapExecuteCommand(command string) bool {
	return strings.Contains(command, "\n") || len(command) > 512
}

func writeExecuteScript(command string) string {
	scriptPath := executeScriptPath(command)
	if scriptPath == "" {
		return ""
	}
	scriptDir := filepath.Dir(scriptPath)
	if err := os.MkdirAll(scriptDir, 0o700); err != nil {
		return ""
	}
	if err := os.Chmod(scriptDir, 0o700); err != nil {
		return ""
	}

	content := "#!/bin/sh\n" + focusDebugPrelude(scriptDir) + command + "\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0o700); err != nil {
		return ""
	}
	if err := os.Chmod(scriptPath, 0o700); err != nil {
		return ""
	}
	return scriptPath
}

func executeScriptPath(command string) string {
	if command == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(command))
	return filepath.Join(tasklightCacheDir(), "focus", hex.EncodeToString(sum[:8])+".sh")
}

func focusDebugPrelude(scriptDir string) string {
	if !environmentFlagEnabled(os.Getenv("TASKLIGHT_FOCUS_DEBUG")) {
		return ""
	}
	logPath := filepath.Join(scriptDir, "focus.log")
	return "umask 077\nexec >> " + shellQuote(logPath) + " 2>&1\nset -x\ndate\n"
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

func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuote(arg)
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func startCommand(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}

func runCommand(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}
