//go:build darwin
// +build darwin

package doctor

import (
	"errors"
	"strings"
	"testing"
)

func TestDetectMacOSHelperDiagnostics(t *testing.T) {
	appPath := "/Applications/Tasklight.app"
	var commands []string
	diagnostics := detectMacOSHelperDiagnosticsWith(appPath, func(name string, args ...string) ([]byte, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		switch {
		case strings.HasSuffix(name, "/TasklightNotifier"):
			return []byte("bundle-id=dev.tasklight.Tasklight\nauthorization=authorized\nalerts=enabled\nsounds=disabled\n"), nil
		case name == "codesign" && len(args) > 0 && args[0] == "--display":
			return []byte("Signature=adhoc\nflags=0x10002(adhoc,runtime)"), nil
		case name == "xattr":
			return nil, errors.New("attribute not found")
		default:
			return nil, nil
		}
	})

	if diagnostics.ExecutablePath != appPath+"/Contents/MacOS/TasklightNotifier" {
		t.Fatalf("ExecutablePath = %q", diagnostics.ExecutablePath)
	}
	if diagnostics.BundleID != "dev.tasklight.Tasklight" || diagnostics.Authorization != "authorized" || diagnostics.Alerts != "enabled" || diagnostics.Sounds != "disabled" {
		t.Fatalf("helper settings = %#v", diagnostics)
	}
	if !diagnostics.SignatureValid || !diagnostics.HardenedRuntime || diagnostics.SignatureDescription != "valid ad hoc signature, hardened runtime" {
		t.Fatalf("signature diagnostics = %#v", diagnostics)
	}
	if !diagnostics.RegistrationAttempted || !diagnostics.RegistrationOK || diagnostics.Quarantined {
		t.Fatalf("registration/quarantine diagnostics = %#v", diagnostics)
	}
	if len(commands) != 5 {
		t.Fatalf("commands = %#v, want helper, codesign verify/display, registration, and xattr", commands)
	}
}
