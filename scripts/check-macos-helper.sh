#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP="${1:-}"
TARGET="${2:-}"
EXPECTED_VERSION="${3:-}"

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "macOS helper checks require macOS" >&2
  exit 2
fi
if [[ -z "$APP" || ! -d "$APP" ]]; then
  echo "usage: $0 <Tasklight.app> [darwin-arm64|darwin-amd64] [version]" >&2
  exit 2
fi

PLIST="$APP/Contents/Info.plist"
EXECUTABLE="$APP/Contents/MacOS/TasklightNotifier"
ICON="$APP/Contents/Resources/Tasklight-v2.icns"

for path in "$PLIST" "$EXECUTABLE" "$ICON"; do
  if [[ ! -f "$path" ]]; then
    echo "missing native helper file: $path" >&2
    exit 1
  fi
done
if [[ ! -x "$EXECUTABLE" ]]; then
  echo "native helper executable is not executable: $EXECUTABLE" >&2
  exit 1
fi

plutil -lint "$PLIST" >/dev/null
bundle_id="$(plutil -extract CFBundleIdentifier raw "$PLIST")"
version="$(plutil -extract CFBundleShortVersionString raw "$PLIST")"
icon_name="$(plutil -extract CFBundleIconFile raw "$PLIST")"
if [[ "$bundle_id" != "dev.tasklight.notifier" ]]; then
  echo "unexpected native helper bundle ID: $bundle_id" >&2
  exit 1
fi
if [[ "$icon_name" != "Tasklight-v2" ]]; then
  echo "unexpected native helper icon resource: $icon_name" >&2
  exit 1
fi
if ! cmp -s "$ROOT/assets/brand/Tasklight.icns" "$ICON"; then
  echo "native helper icon does not match the repository brand asset" >&2
  exit 1
fi
if [[ -n "$EXPECTED_VERSION" && "$version" != "$EXPECTED_VERSION" ]]; then
  echo "native helper version $version does not match package version $EXPECTED_VERSION" >&2
  exit 1
fi

codesign --verify --deep --strict "$APP"
signature="$(codesign --display --verbose=4 "$APP" 2>&1)"
if [[ "$signature" != *"Signature=adhoc"* ]]; then
  echo "native helper must use an explicit ad hoc signature" >&2
  exit 1
fi
if [[ "$signature" != *"runtime"* ]]; then
  echo "native helper signature is missing the hardened runtime flag" >&2
  exit 1
fi
if xattr -p com.apple.quarantine "$APP" >/dev/null 2>&1; then
  echo "native helper build unexpectedly carries a quarantine attribute" >&2
  exit 1
fi

architectures="$(lipo -archs "$EXECUTABLE")"
case "$TARGET" in
  darwin-arm64) expected_arch="arm64" ;;
  darwin-amd64) expected_arch="x86_64" ;;
  "") expected_arch="" ;;
  *) echo "unsupported helper target: $TARGET" >&2; exit 2 ;;
esac
if [[ -n "$expected_arch" && " $architectures " != *" $expected_arch "* ]]; then
  echo "native helper architectures ($architectures) do not include $expected_arch" >&2
  exit 1
fi

host_arch="$(uname -m)"
if [[ "$host_arch" == "x86_64" ]]; then
  host_target="darwin-amd64"
else
  host_target="darwin-$host_arch"
fi
if [[ -n "$TARGET" && "$TARGET" != "$host_target" ]]; then
  echo "validated $TARGET bundle structure and signature (runtime checks skipped on $host_arch)"
  exit 0
fi

TMP="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP"
}
trap cleanup EXIT

run_with_timeout() {
  "$@" &
  local command_pid=$!
  (
    sleep 5
    if kill -0 "$command_pid" >/dev/null 2>&1; then
      kill -TERM "$command_pid" >/dev/null 2>&1 || true
    fi
  ) &
  local watchdog_pid=$!
  local status=0
  wait "$command_pid" || status=$?
  kill "$watchdog_pid" >/dev/null 2>&1 || true
  wait "$watchdog_pid" >/dev/null 2>&1 || true
  return "$status"
}

run_with_timeout "$EXECUTABLE" --help >"$TMP/help.txt"
grep -F "TasklightNotifier doctor" "$TMP/help.txt" >/dev/null
run_with_timeout "$EXECUTABLE" self-test --timeout 0.05 >"$TMP/self-test.txt"
grep -Fx "timeout-exit=ok" "$TMP/self-test.txt" >/dev/null

mkdir -p "$TMP/home"
CFFIXED_USER_HOME="$TMP/home" TASKLIGHT_FOCUS_DEBUG=0 run_with_timeout "$EXECUTABLE" self-test --timeout 0 >/dev/null
DEBUG_LOG="$TMP/home/Library/Caches/tasklight/native-helper.log"
if [[ -e "$DEBUG_LOG" ]]; then
  echo "TASKLIGHT_FOCUS_DEBUG=0 unexpectedly enabled helper logging" >&2
  exit 1
fi
CFFIXED_USER_HOME="$TMP/home" TASKLIGHT_FOCUS_DEBUG=1 run_with_timeout "$EXECUTABLE" self-test --timeout 0 >/dev/null
if [[ ! -f "$DEBUG_LOG" || "$(stat -f '%Lp' "$DEBUG_LOG")" != "600" ]]; then
  echo "native helper debug log was not created with mode 0600" >&2
  exit 1
fi

run_with_timeout "$EXECUTABLE" doctor --timeout 3 >"$TMP/doctor.txt"
grep -Fx "bundle-id=dev.tasklight.notifier" "$TMP/doctor.txt" >/dev/null
grep -E '^authorization=(authorized|denied|not-determined|provisional|ephemeral|unknown|unavailable)$' "$TMP/doctor.txt" >/dev/null
grep -E '^alerts=(enabled|disabled|not-supported|unknown)$' "$TMP/doctor.txt" >/dev/null
grep -E '^sounds=(enabled|disabled|not-supported|unknown)$' "$TMP/doctor.txt" >/dev/null

echo "validated native helper $version ($architectures), signing, diagnostics, and bounded timeout exit"
