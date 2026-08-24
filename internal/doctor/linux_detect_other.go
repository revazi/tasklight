//go:build !linux
// +build !linux

package doctor

func detectLinuxDiagnostics() linuxDiagnostics {
	return linuxDiagnostics{}
}
