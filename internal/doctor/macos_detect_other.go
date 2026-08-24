//go:build !darwin
// +build !darwin

package doctor

func detectMacOSHelperDiagnostics(string) macOSHelperDiagnostics {
	return macOSHelperDiagnostics{}
}
