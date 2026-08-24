//go:build linux
// +build linux

package doctor

import (
	"errors"
	"reflect"
	"testing"
)

func TestDetectLinuxDiagnostics(t *testing.T) {
	environment := map[string]string{
		"XDG_CURRENT_DESKTOP":      "GNOME",
		"XDG_SESSION_TYPE":         "wayland",
		"WAYLAND_DISPLAY":          "wayland-0",
		"DISPLAY":                  ":1",
		"DBUS_SESSION_BUS_ADDRESS": "unix:path=/run/user/1000/bus",
		"TASKLIGHT_LINUX_ACTIONS":  "yes",
	}
	paths := map[string]string{
		"notify-send": "/usr/bin/notify-send",
		"gdbus":       "/usr/bin/gdbus",
		"busctl":      "/usr/bin/busctl",
	}
	var commands [][]string

	diagnostics := detectLinuxDiagnosticsWith(
		func(name string) string { return environment[name] },
		func(name string) (string, error) {
			if path := paths[name]; path != "" {
				return path, nil
			}
			return "", errors.New("not found")
		},
		func(name string, args ...string) ([]byte, error) {
			commands = append(commands, append([]string{name}, args...))
			if name == "/usr/bin/notify-send" {
				return []byte("  -A, --action=ACTION  Specifies the actions to display"), nil
			}
			return []byte("('GNOME Shell', 'GNOME', '46', '1.2')"), nil
		},
	)

	if diagnostics.Desktop != "GNOME" || diagnostics.SessionType != "wayland" || diagnostics.WaylandDisplay != "wayland-0" || diagnostics.X11Display != ":1" {
		t.Fatalf("desktop diagnostics = %#v", diagnostics)
	}
	if !diagnostics.DBusSession || diagnostics.NotificationServer == "" || !diagnostics.NotifyActionsKnown || !diagnostics.NotifyActions || !diagnostics.ActionsEnabled {
		t.Fatalf("notification diagnostics = %#v", diagnostics)
	}
	if diagnostics.GDBusPath != "/usr/bin/gdbus" || diagnostics.BusctlPath != "/usr/bin/busctl" || diagnostics.WMCtrlPath != "" || diagnostics.XDoToolPath != "" {
		t.Fatalf("tool diagnostics = %#v", diagnostics)
	}
	if len(commands) != 2 {
		t.Fatalf("commands = %#v, want notify-send help and gdbus query", commands)
	}
	wantServerCommand := []string{
		"/usr/bin/gdbus",
		"call",
		"--session",
		"--dest", "org.freedesktop.Notifications",
		"--object-path", "/org/freedesktop/Notifications",
		"--method", "org.freedesktop.Notifications.GetServerInformation",
	}
	if !reflect.DeepEqual(commands[1], wantServerCommand) {
		t.Fatalf("server command = %#v, want %#v", commands[1], wantServerCommand)
	}
}

func TestDetectLinuxDiagnosticsInfersSessionType(t *testing.T) {
	diagnostics := detectLinuxDiagnosticsWith(
		func(name string) string {
			if name == "DISPLAY" {
				return ":0"
			}
			return ""
		},
		func(string) (string, error) { return "", errors.New("not found") },
		func(string, ...string) ([]byte, error) { return nil, errors.New("not found") },
	)

	if diagnostics.SessionType != "x11 (inferred)" {
		t.Fatalf("SessionType = %q, want inferred x11", diagnostics.SessionType)
	}
}
