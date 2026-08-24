//go:build linux
// +build linux

package doctor

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

type linuxOutputRunner func(name string, args ...string) ([]byte, error)

func detectLinuxDiagnostics() linuxDiagnostics {
	return detectLinuxDiagnosticsWith(os.Getenv, exec.LookPath, linuxCommandOutput)
}

func detectLinuxDiagnosticsWith(getenv func(string) string, lookPath func(string) (string, error), output linuxOutputRunner) linuxDiagnostics {
	diagnostics := linuxDiagnostics{
		Desktop:        firstNonEmpty(getenv("XDG_CURRENT_DESKTOP"), getenv("DESKTOP_SESSION")),
		SessionType:    strings.TrimSpace(getenv("XDG_SESSION_TYPE")),
		WaylandDisplay: strings.TrimSpace(getenv("WAYLAND_DISPLAY")),
		X11Display:     strings.TrimSpace(getenv("DISPLAY")),
		DBusSession:    strings.TrimSpace(getenv("DBUS_SESSION_BUS_ADDRESS")) != "",
		ActionsEnabled: linuxActionsEnabled(getenv("TASKLIGHT_LINUX_ACTIONS")),
	}
	if diagnostics.SessionType == "" {
		switch {
		case diagnostics.WaylandDisplay != "":
			diagnostics.SessionType = "wayland (inferred)"
		case diagnostics.X11Display != "":
			diagnostics.SessionType = "x11 (inferred)"
		}
	}

	diagnostics.GDBusPath = optionalPath(lookPath, "gdbus")
	diagnostics.WMCtrlPath = optionalPath(lookPath, "wmctrl")
	diagnostics.XDoToolPath = optionalPath(lookPath, "xdotool")
	diagnostics.BusctlPath = optionalPath(lookPath, "busctl")

	if notifySendPath := optionalPath(lookPath, "notify-send"); notifySendPath != "" {
		if help, err := output(notifySendPath, "--help"); err == nil {
			diagnostics.NotifyActionsKnown = true
			diagnostics.NotifyActions = strings.Contains(string(help), "--action")
		}
	}

	if diagnostics.DBusSession && diagnostics.GDBusPath != "" {
		server, err := output(
			diagnostics.GDBusPath,
			"call",
			"--session",
			"--dest", "org.freedesktop.Notifications",
			"--object-path", "/org/freedesktop/Notifications",
			"--method", "org.freedesktop.Notifications.GetServerInformation",
		)
		if err != nil {
			diagnostics.NotificationError = strings.TrimSpace(string(server))
			if diagnostics.NotificationError == "" {
				diagnostics.NotificationError = err.Error()
			}
		} else {
			diagnostics.NotificationServer = strings.TrimSpace(string(server))
		}
	}

	return diagnostics
}

func linuxCommandOutput(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

func optionalPath(lookPath func(string) (string, error), name string) string {
	path, err := lookPath(name)
	if err != nil {
		return ""
	}
	return path
}

func linuxActionsEnabled(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
