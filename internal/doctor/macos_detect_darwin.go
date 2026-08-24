//go:build darwin
// +build darwin

package doctor

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type macOSCommandOutput func(name string, args ...string) ([]byte, error)

func detectMacOSHelperDiagnostics(helperPath string) macOSHelperDiagnostics {
	return detectMacOSHelperDiagnosticsWith(helperPath, runMacOSDiagnosticCommand)
}

func detectMacOSHelperDiagnosticsWith(helperPath string, run macOSCommandOutput) macOSHelperDiagnostics {
	diagnostics := macOSHelperDiagnostics{}
	if helperPath == "" {
		return diagnostics
	}

	appBundle := strings.HasSuffix(helperPath, ".app")
	executablePath := helperPath
	if appBundle {
		executablePath = filepath.Join(helperPath, "Contents", "MacOS", "TasklightNotifier")
	}
	diagnostics.ExecutablePath = executablePath

	if output, err := run(executablePath, "doctor", "--timeout", "3"); err != nil {
		diagnostics.InspectionError = commandDiagnosticError(output, err)
	} else {
		values := parseMacOSHelperOutput(string(output))
		diagnostics.BundleID = values["bundle-id"]
		diagnostics.Authorization = values["authorization"]
		diagnostics.Alerts = values["alerts"]
		diagnostics.Sounds = values["sounds"]
	}

	signatureTarget := helperPath
	if output, err := run("codesign", "--verify", "--deep", "--strict", signatureTarget); err != nil {
		diagnostics.SignatureError = commandDiagnosticError(output, err)
	} else {
		diagnostics.SignatureValid = true
		details, _ := run("codesign", "--display", "--verbose=4", signatureTarget)
		diagnostics.HardenedRuntime = strings.Contains(string(details), "runtime")
		diagnostics.SignatureDescription = macOSSignatureDescription(string(details))
	}

	if appBundle {
		diagnostics.RegistrationAttempted = true
		lsregister := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
		if output, err := run(lsregister, "-f", helperPath); err != nil {
			diagnostics.RegistrationError = commandDiagnosticError(output, err)
		} else {
			diagnostics.RegistrationOK = true
		}
	}

	if _, err := run("xattr", "-p", "com.apple.quarantine", helperPath); err == nil {
		diagnostics.Quarantined = true
	}
	return diagnostics
}

func runMacOSDiagnosticCommand(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

func commandDiagnosticError(output []byte, err error) string {
	message := strings.TrimSpace(string(output))
	if message == "" {
		return err.Error()
	}
	return fmt.Sprintf("%s (%v)", message, err)
}
