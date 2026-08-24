package doctor

import (
	"fmt"
	"io"
)

type linuxDiagnostics struct {
	Desktop            string
	SessionType        string
	WaylandDisplay     string
	X11Display         string
	DBusSession        bool
	NotificationServer string
	NotificationError  string
	NotifyActions      bool
	NotifyActionsKnown bool
	ActionsEnabled     bool
	GDBusPath          string
	WMCtrlPath         string
	XDoToolPath        string
	BusctlPath         string
}

func writeLinuxDiagnostics(w io.Writer, diagnostics linuxDiagnostics) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Linux desktop integration")
	info(w, "desktop", diagnosticValue(diagnostics.Desktop))
	info(w, "session type", diagnosticValue(diagnostics.SessionType))
	info(w, "Wayland display", diagnosticValue(diagnostics.WaylandDisplay))
	info(w, "X11 display", diagnosticValue(diagnostics.X11Display))
	if diagnostics.DBusSession {
		okLine(w, "D-Bus session", "available")
	} else {
		warnLine(w, "D-Bus session", "not detected; desktop notifications may be unavailable")
	}
	if diagnostics.NotificationServer != "" {
		okLine(w, "notification service", diagnostics.NotificationServer)
	} else if diagnostics.NotificationError != "" {
		warnLine(w, "notification service", diagnostics.NotificationError)
	} else {
		warnLine(w, "notification service", "not queried; gdbus and a session bus are required")
	}

	actionSupport := "not advertised"
	if !diagnostics.NotifyActionsKnown {
		actionSupport = "unknown"
	} else if diagnostics.NotifyActions {
		actionSupport = "advertised by notify-send"
	}
	info(w, "notification actions", actionSupport)
	if diagnostics.ActionsEnabled && diagnostics.NotifyActions {
		warnLine(w, "click-to-focus", "best-effort opt-in enabled; notification actions run the captured tmux command")
	} else if diagnostics.ActionsEnabled {
		warnLine(w, "click-to-focus", "opt-in requested, but notify-send action support was not detected")
	} else {
		warnLine(w, "click-to-focus", "disabled by default; set TASKLIGHT_LINUX_ACTIONS=1 for best-effort tmux actions")
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Optional Linux focus tools")
	optionalToolLine(w, "gdbus", diagnostics.GDBusPath, "D-Bus diagnostics/actions")
	optionalToolLine(w, "wmctrl", diagnostics.WMCtrlPath, "X11 window activation")
	optionalToolLine(w, "xdotool", diagnostics.XDoToolPath, "X11 input/window control")
	optionalToolLine(w, "busctl", diagnostics.BusctlPath, "D-Bus diagnostics")
}

func optionalToolLine(w io.Writer, name string, path string, purpose string) {
	if path != "" {
		okLine(w, name, path+" ("+purpose+")")
		return
	}
	info(w, name, "not found (optional; "+purpose+")")
}
