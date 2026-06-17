package brandassets

import (
	"bytes"
	_ "embed"
	"html"
	"os"
	"path/filepath"
)

//go:embed tasklight-app-icon-1024.png
var defaultIcon []byte

//go:embed Tasklight.icns
var macOSAppIcon []byte

const defaultIconFileName = "tasklight-app-icon-1024.png"
const macOSAppName = "Tasklight.app"

// DefaultIconPath returns a filesystem path to the bundled Tasklight app icon.
//
// Notification providers such as terminal-notifier and notify-send need an icon
// path, not embedded bytes, so the asset is written to the user's cache dir on
// demand. TASKLIGHT_ICON can override the bundled icon.
func DefaultIconPath() string {
	if override := os.Getenv("TASKLIGHT_ICON"); override != "" {
		return override
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil || cacheDir == "" {
		cacheDir = os.TempDir()
	}

	iconDir := filepath.Join(cacheDir, "tasklight")
	if err := os.MkdirAll(iconDir, 0o755); err != nil {
		return ""
	}

	iconPath := filepath.Join(iconDir, defaultIconFileName)
	if shouldWriteIcon(iconPath) {
		if err := os.WriteFile(iconPath, defaultIcon, 0o644); err != nil {
			return ""
		}
	}

	return iconPath
}

// DefaultMacOSAppBundle returns a filesystem path to a minimal Tasklight.app
// bundle with the bundled Tasklight icon and the provided bundle identifier.
//
// macOS notification providers use the sender application's bundle identity for
// the left-side notification icon. A loose PNG is not enough for that slot.
func DefaultMacOSAppBundle(bundleID string) string {
	if bundleID == "" {
		return ""
	}

	baseDir, err := os.UserConfigDir()
	if err != nil || baseDir == "" {
		baseDir = os.TempDir()
	}

	appPath := filepath.Join(baseDir, "Tasklight", macOSAppName)
	contentsPath := filepath.Join(appPath, "Contents")
	macOSPath := filepath.Join(contentsPath, "MacOS")
	resourcesPath := filepath.Join(contentsPath, "Resources")

	if err := os.MkdirAll(macOSPath, 0o755); err != nil {
		return ""
	}
	if err := os.MkdirAll(resourcesPath, 0o755); err != nil {
		return ""
	}

	infoPlistPath := filepath.Join(contentsPath, "Info.plist")
	executablePath := filepath.Join(macOSPath, "tasklight-notification-helper")
	iconPath := filepath.Join(resourcesPath, "Tasklight.icns")

	if !writeFileIfNeeded(infoPlistPath, []byte(macOSInfoPlist(bundleID)), 0o644) {
		return ""
	}
	if !writeFileIfNeeded(executablePath, []byte("#!/bin/sh\nexit 0\n"), 0o755) {
		return ""
	}
	if err := os.Chmod(executablePath, 0o755); err != nil {
		return ""
	}
	if !writeFileIfNeeded(iconPath, macOSAppIcon, 0o644) {
		return ""
	}

	return appPath
}

func shouldWriteIcon(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return info.Size() != int64(len(defaultIcon))
}

func writeFileIfNeeded(path string, content []byte, perm os.FileMode) bool {
	current, err := os.ReadFile(path)
	if err == nil && bytes.Equal(current, content) {
		return true
	}
	return os.WriteFile(path, content, perm) == nil
}

func macOSInfoPlist(bundleID string) string {
	escapedBundleID := html.EscapeString(bundleID)
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>en</string>
	<key>CFBundleDisplayName</key>
	<string>Tasklight</string>
	<key>CFBundleExecutable</key>
	<string>tasklight-notification-helper</string>
	<key>CFBundleIconFile</key>
	<string>Tasklight</string>
	<key>CFBundleIdentifier</key>
	<string>` + escapedBundleID + `</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>Tasklight</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>0.1.0</string>
	<key>CFBundleVersion</key>
	<string>1</string>
	<key>LSBackgroundOnly</key>
	<true/>
</dict>
</plist>
`
}
