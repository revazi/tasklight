package doctor

import (
	"fmt"
	"io"
	"strings"
)

type macOSHelperDiagnostics struct {
	ExecutablePath        string
	BundleID              string
	Authorization         string
	Alerts                string
	Sounds                string
	SignatureValid        bool
	HardenedRuntime       bool
	SignatureDescription  string
	SignatureError        string
	RegistrationAttempted bool
	RegistrationOK        bool
	RegistrationError     string
	Quarantined           bool
	InspectionError       string
}

func writeMacOSHelperDiagnostics(w io.Writer, diagnostics macOSHelperDiagnostics) {
	if diagnostics.ExecutablePath != "" {
		okLine(w, "helper executable", diagnostics.ExecutablePath)
	}
	if diagnostics.SignatureValid && diagnostics.HardenedRuntime {
		okLine(w, "helper signature", diagnosticValue(diagnostics.SignatureDescription))
	} else if diagnostics.SignatureValid {
		warnLine(w, "helper signature", diagnosticValue(diagnostics.SignatureDescription)+"; hardened runtime missing")
	} else if diagnostics.SignatureError != "" {
		warnLine(w, "helper signature", diagnostics.SignatureError)
	}
	if diagnostics.Quarantined {
		warnLine(w, "helper quarantine", "quarantine attribute present; Gatekeeper may block the helper")
	} else {
		okLine(w, "helper quarantine", "not present")
	}
	if diagnostics.RegistrationAttempted {
		if diagnostics.RegistrationOK {
			okLine(w, "helper registration", diagnosticValue(diagnostics.BundleID))
		} else {
			warnLine(w, "helper registration", diagnostics.RegistrationError)
		}
	}

	switch diagnostics.Authorization {
	case "authorized", "provisional", "ephemeral":
		okLine(w, "notification authorization", diagnostics.Authorization)
	case "denied":
		warnLine(w, "notification authorization", "denied; enable Tasklight in System Settings > Notifications")
	case "not-determined":
		warnLine(w, "notification authorization", "not requested yet; the first notification will ask for permission")
	case "unavailable", "unknown", "":
		message := diagnostics.InspectionError
		if strings.TrimSpace(message) == "" {
			message = "could not determine notification settings"
		}
		warnLine(w, "notification authorization", message)
	default:
		warnLine(w, "notification authorization", diagnostics.Authorization)
	}
	if diagnostics.Alerts != "" {
		info(w, "notification alerts", diagnostics.Alerts)
	}
	if diagnostics.Sounds != "" {
		info(w, "notification sounds", diagnostics.Sounds)
	}
}

func parseMacOSHelperOutput(output string) map[string]string {
	values := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if ok && key != "" {
			values[key] = value
		}
	}
	return values
}

func macOSSignatureDescription(output string) string {
	description := "valid signature"
	if strings.Contains(output, "Signature=adhoc") {
		description = "valid ad hoc signature"
	} else {
		for _, line := range strings.Split(output, "\n") {
			if authority, ok := strings.CutPrefix(strings.TrimSpace(line), "Authority="); ok && authority != "" {
				description = fmt.Sprintf("valid Developer ID signature (%s)", authority)
				break
			}
		}
	}
	if strings.Contains(output, "runtime") {
		description += ", hardened runtime"
	}
	return description
}
