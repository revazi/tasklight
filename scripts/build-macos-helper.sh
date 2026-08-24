#!/usr/bin/env bash
set -euo pipefail
umask 022

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${1:-$ROOT/bin}"
TARGET="${2:-darwin-$(uname -m | sed 's/x86_64/amd64/')}"
APP="$OUT_DIR/Tasklight.app"
CONTENTS="$APP/Contents"
MACOS="$CONTENTS/MacOS"
RESOURCES="$CONTENTS/Resources"
SOURCE="$ROOT/helpers/macos/TasklightNotifier/TasklightNotifier.swift"
ICON="$ROOT/assets/brand/Tasklight.icns"
EXECUTABLE="$MACOS/TasklightNotifier"
VERSION="${TASKLIGHT_VERSION:-}"
if [[ -z "$VERSION" ]]; then
  VERSION="$(sed -nE 's/^[[:space:]]*"version":[[:space:]]*"([0-9]+\.[0-9]+\.[0-9]+)".*/\1/p' "$ROOT/npm/tasklight-cli/package.json" | head -n 1)"
fi
if [[ ! "$VERSION" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
  echo "invalid Tasklight helper version: $VERSION" >&2
  exit 2
fi

case "$TARGET" in
  darwin-arm64) SWIFT_TARGET="arm64-apple-macosx13.0" ;;
  darwin-amd64) SWIFT_TARGET="x86_64-apple-macosx13.0" ;;
  *) echo "unsupported macOS helper target: $TARGET" >&2; exit 2 ;;
esac

rm -rf "$APP"
mkdir -p "$MACOS" "$RESOURCES"
cp "$ICON" "$RESOURCES/Tasklight.icns"

cat > "$CONTENTS/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>en</string>
	<key>CFBundleDisplayName</key>
	<string>Tasklight</string>
	<key>CFBundleExecutable</key>
	<string>TasklightNotifier</string>
	<key>CFBundleIconFile</key>
	<string>Tasklight</string>
	<key>CFBundleIdentifier</key>
	<string>dev.tasklight.Tasklight</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>Tasklight</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>$VERSION</string>
	<key>CFBundleVersion</key>
	<string>$VERSION</string>
	<key>LSUIElement</key>
	<true/>
	<key>NSUserNotificationAlertStyle</key>
	<string>alert</string>
</dict>
</plist>
PLIST

xcrun swiftc \
  -O \
  -target "$SWIFT_TARGET" \
  -framework AppKit \
  -framework UserNotifications \
  -o "$EXECUTABLE" \
  "$SOURCE"

chmod 0755 "$EXECUTABLE"
chmod 0644 "$CONTENTS/Info.plist" "$RESOURCES/Tasklight.icns"

if ! command -v codesign >/dev/null 2>&1; then
  echo "codesign is required to package the native macOS helper" >&2
  exit 1
fi
find "$APP" -exec xattr -c {} +
codesign --force --sign - --options runtime --timestamp=none "$APP"
codesign --verify --deep --strict "$APP"

LSREGISTER="/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
if [[ -x "$LSREGISTER" ]]; then
  "$LSREGISTER" -f "$APP" >/dev/null
fi

"$ROOT/scripts/check-macos-helper.sh" "$APP" "$TARGET" "$VERSION" >/dev/null

echo "$APP"
