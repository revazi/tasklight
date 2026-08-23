package doctor

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/revazi/tasklight/internal/notify"
	"github.com/revazi/tasklight/internal/session"
)

// FocusReport contains the state needed to troubleshoot notification
// click-to-focus behavior.
type FocusReport struct {
	Platform     string
	GoVersion    string
	Target       session.FocusTarget
	ClickCommand string
	Notification notify.FocusDiagnostics
}

// RunFocus prints a pasteable focus/session diagnostic report.
func RunFocus(w io.Writer) int {
	target := session.Detect(session.DetectOptions{})
	clickCommand := target.ClickCommand()
	report := FocusReport{
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		GoVersion:    runtime.Version(),
		Target:       target,
		ClickCommand: clickCommand,
		Notification: notify.DiagnoseFocus(clickCommand),
	}
	WriteFocusReport(w, report)
	return 0
}

// WriteFocusReport renders a deterministic diagnostic report.
func WriteFocusReport(w io.Writer, report FocusReport) {
	fmt.Fprintln(w, "Tasklight focus diagnostics")
	fmt.Fprintln(w)
	info(w, "Platform", diagnosticValue(report.Platform))
	info(w, "Go", diagnosticValue(report.GoVersion))

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Notification delivery")
	info(w, "selected provider", diagnosticValue(report.Notification.Provider))
	info(w, "provider path", diagnosticValue(report.Notification.ProviderPath))
	info(w, "native helper path", diagnosticValue(report.Notification.NativeHelperPath))
	info(w, "click actions", yesNo(report.Notification.SupportsClick))

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Terminal target")
	info(w, "activate app", diagnosticValue(report.Target.ActivateApp))
	info(w, "terminal name", diagnosticValue(report.Target.Terminal.Name))
	info(w, "terminal bundle ID", diagnosticValue(report.Target.Terminal.BundleID))
	info(w, "iTerm session ID", diagnosticValue(report.Target.Terminal.ITermSessionID))
	info(w, "Terminal session ID", diagnosticValue(report.Target.Terminal.TerminalSession))
	info(w, "terminal/client tty", diagnosticValue(report.Target.Terminal.ClientTTY))

	fmt.Fprintln(w)
	fmt.Fprintln(w, "tmux target")
	tmux := report.Target.Tmux
	info(w, "inside tmux", yesNo(tmux != nil))
	if tmux == nil {
		tmux = &session.TmuxTarget{}
	}
	info(w, "tmux socket", diagnosticValue(tmux.Socket))
	info(w, "tmux client", diagnosticValue(tmux.ClientName))
	info(w, "tmux client tty", diagnosticValue(tmux.ClientTTY))
	info(w, "tmux session", diagnosticValue(tmux.Session))
	info(w, "tmux window index", diagnosticValue(tmux.WindowIndex))
	info(w, "tmux window ID", diagnosticValue(tmux.WindowID))
	info(w, "tmux pane index", diagnosticValue(tmux.PaneIndex))
	info(w, "tmux pane ID", diagnosticValue(tmux.PaneID))

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Generated focus action")
	clickCommand := report.ClickCommand
	block(w, "click command", clickCommand)
	providerCommand := report.Notification.ExecutionCommand
	if providerCommand != "" && providerCommand == clickCommand {
		providerCommand = "(same as click command)"
	}
	block(w, "provider command", providerCommand)
	info(w, "generated script path", diagnosticValue(report.Notification.ScriptPath))

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Debug logging")
	info(w, "TASKLIGHT_FOCUS_DEBUG", enabledDisabled(report.Notification.DebugEnabled))
	info(w, "native helper log", diagnosticValue(report.Notification.NativeHelperLogPath))
	info(w, "focus execution log", diagnosticValue(report.Notification.FocusLogPath))
	if !report.Notification.DebugEnabled {
		info(w, "enable logs", "set TASKLIGHT_FOCUS_DEBUG=1 when sending a notification")
	}
}

func diagnosticValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(not detected)"
	}
	return value
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func enabledDisabled(value bool) string {
	if value {
		return "enabled"
	}
	return "disabled"
}

func block(w io.Writer, name string, value string) {
	value = diagnosticValue(value)
	lines := strings.Split(value, "\n")
	if len(lines) == 1 {
		info(w, name, lines[0])
		return
	}
	fmt.Fprintf(w, "  • %s\n", name)
	for _, line := range lines {
		fmt.Fprintf(w, "      %s\n", line)
	}
}
