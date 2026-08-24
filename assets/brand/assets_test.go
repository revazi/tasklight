package brandassets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMacOSInfoPlistUsesBrandVersion(t *testing.T) {
	plist := macOSInfoPlist("dev.tasklight.Tasklight")
	for _, want := range []string{"<string>2.1</string>", "<string>2</string>"} {
		if !strings.Contains(plist, want) {
			t.Fatalf("macOSInfoPlist() missing %q", want)
		}
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
