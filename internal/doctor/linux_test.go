package doctor

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteLinuxDiagnosticsIncludesDesktopCapabilities(t *testing.T) {
	diagnostics := linuxDiagnostics{
		Desktop:            "GNOME",
		SessionType:        "wayland",
		WaylandDisplay:     "wayland-0",
		DBusSession:        true,
		NotificationServer: "('GNOME Shell', 'GNOME', '46', '1.2')",
		NotifyActions:      true,
		NotifyActionsKnown: true,
		GDBusPath:          "/usr/bin/gdbus",
		WMCtrlPath:         "/usr/bin/wmctrl",
	}

	var output bytes.Buffer
	writeLinuxDiagnostics(&output, diagnostics)
	got := output.String()

	for _, want := range []string{
		"Linux desktop integration",
		"GNOME",
		"wayland",
		"wayland-0",
		"D-Bus session",
		"GNOME Shell",
		"advertised by notify-send",
		"click-to-focus",
		"disabled by default",
		"TASKLIGHT_LINUX_ACTIONS=1",
		"Optional Linux focus tools",
		"/usr/bin/gdbus",
		"/usr/bin/wmctrl",
		"xdotool",
		"not found",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Linux diagnostics missing %q:\n%s", want, got)
		}
	}
}
