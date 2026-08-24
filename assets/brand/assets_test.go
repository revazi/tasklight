package brandassets

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMacOSInfoPlistUsesBrandVersion(t *testing.T) {
	plist := macOSInfoPlist("dev.tasklight.sender.v2")
	for _, want := range []string{"<string>2.1</string>", "<string>2</string>", "<string>Tasklight-v2</string>", "<string>dev.tasklight.sender.v2</string>"} {
		if !strings.Contains(plist, want) {
			t.Fatalf("macOSInfoPlist() missing %q", want)
		}
	}
}

func TestDefaultMacOSAppBundleUsesCacheSafeIconResource(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))

	appPath := DefaultMacOSAppBundle("dev.tasklight.sender.v2")
	if appPath == "" {
		t.Fatal("DefaultMacOSAppBundle() returned an empty path")
	}

	plist, err := os.ReadFile(filepath.Join(appPath, "Contents", "Info.plist"))
	if err != nil {
		t.Fatalf("ReadFile(Info.plist): %v", err)
	}
	for _, want := range []string{"<string>Tasklight-v2</string>", "<string>dev.tasklight.sender.v2</string>"} {
		if !strings.Contains(string(plist), want) {
			t.Fatalf("Info.plist missing %q", want)
		}
	}

	icon, err := os.ReadFile(filepath.Join(appPath, "Contents", "Resources", macOSIconResourceFileName))
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", macOSIconResourceFileName, err)
	}
	if !bytes.Equal(icon, macOSAppIcon) {
		t.Fatal("fallback sender icon does not match the embedded brand asset")
	}

	legacyIcon := filepath.Join(appPath, "Contents", "Resources", "Tasklight.icns")
	if err := os.WriteFile(legacyIcon, []byte("stale"), 0o644); err != nil {
		t.Fatalf("WriteFile(legacy icon): %v", err)
	}
	if got := DefaultMacOSAppBundle("dev.tasklight.sender.v2"); got != appPath {
		t.Fatalf("DefaultMacOSAppBundle() = %q after upgrade, want %q", got, appPath)
	}
	if _, err := os.Stat(legacyIcon); !os.IsNotExist(err) {
		t.Fatalf("legacy icon resource still exists: %v", err)
	}
}

func TestShouldWriteIconComparesContent(t *testing.T) {
	original := defaultIcon
	defaultIcon = []byte("new-icon")
	t.Cleanup(func() { defaultIcon = original })

	path := filepath.Join(t.TempDir(), "icon.png")
	if !shouldWriteIcon(path) {
		t.Fatal("shouldWriteIcon() = false for a missing icon")
	}
	if err := os.WriteFile(path, []byte("new-icon"), 0o600); err != nil {
		t.Fatalf("WriteFile(same icon): %v", err)
	}
	if shouldWriteIcon(path) {
		t.Fatal("shouldWriteIcon() = true for identical content")
	}
	if err := os.WriteFile(path, []byte("old-icon"), 0o600); err != nil {
		t.Fatalf("WriteFile(stale icon): %v", err)
	}
	if !shouldWriteIcon(path) {
		t.Fatal("shouldWriteIcon() = false for stale content with the same size")
	}
}
