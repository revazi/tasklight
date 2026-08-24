package doctor

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteMacOSHelperDiagnostics(t *testing.T) {
	diagnostics := macOSHelperDiagnostics{
		ExecutablePath:        "/Applications/Tasklight.app/Contents/MacOS/TasklightNotifier",
		BundleID:              "dev.tasklight.notifier",
		Authorization:         "authorized",
		Alerts:                "enabled",
		Sounds:                "disabled",
		SignatureValid:        true,
		HardenedRuntime:       true,
		SignatureDescription:  "valid ad hoc signature, hardened runtime",
		RegistrationAttempted: true,
		RegistrationOK:        true,
	}

	var output bytes.Buffer
	writeMacOSHelperDiagnostics(&output, diagnostics)
	got := output.String()
	for _, want := range []string{
		"helper executable",
		"TasklightNotifier",
		"helper signature",
		"ad hoc signature, hardened runtime",
		"helper quarantine",
		"not present",
		"helper registration",
		"dev.tasklight.notifier",
		"notification authorization",
		"authorized",
		"notification alerts",
		"enabled",
		"notification sounds",
		"disabled",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("macOS diagnostics missing %q:\n%s", want, got)
		}
	}
}

func TestWriteMacOSHelperDiagnosticsExplainsDeniedAuthorization(t *testing.T) {
	var output bytes.Buffer
	writeMacOSHelperDiagnostics(&output, macOSHelperDiagnostics{
		Authorization: "denied",
		Quarantined:   true,
	})
	got := output.String()
	for _, want := range []string{"denied", "System Settings > Notifications", "Gatekeeper may block"} {
		if !strings.Contains(got, want) {
			t.Fatalf("macOS diagnostics missing %q:\n%s", want, got)
		}
	}
}

func TestParseMacOSHelperOutput(t *testing.T) {
	values := parseMacOSHelperOutput("bundle-id=dev.tasklight.notifier\nauthorization=not-determined\nalerts=enabled\ninvalid\n")
	if values["bundle-id"] != "dev.tasklight.notifier" || values["authorization"] != "not-determined" || values["alerts"] != "enabled" {
		t.Fatalf("parseMacOSHelperOutput() = %#v", values)
	}
}

func TestMacOSSignatureDescription(t *testing.T) {
	if got := macOSSignatureDescription("Signature=adhoc\nflags=0x10002(adhoc,runtime)"); got != "valid ad hoc signature, hardened runtime" {
		t.Fatalf("ad hoc signature description = %q", got)
	}
	if got := macOSSignatureDescription("Authority=Developer ID Application: Example\n"); got != "valid Developer ID signature (Developer ID Application: Example)" {
		t.Fatalf("Developer ID signature description = %q", got)
	}
}
