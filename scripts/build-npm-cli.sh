#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PKG="$ROOT/npm/tasklight-cli"
VERSION="$(node -e 'process.stdout.write(require(process.argv[1]).version)' "$PKG/package.json")"
LDFLAGS="-X github.com/revazi/tasklight/internal/cli.Version=$VERSION"

build_target() {
  local goos="$1"
  local goarch="$2"
  local dir="$3"

  mkdir -p "$PKG/vendor/$dir"
  echo "building $dir"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -ldflags "$LDFLAGS" -o "$PKG/vendor/$dir/tasklight" "$ROOT/cmd/tasklight"
  chmod +x "$PKG/vendor/$dir/tasklight"
}

build_target darwin arm64 darwin-arm64
build_target darwin amd64 darwin-amd64
build_target linux arm64 linux-arm64
build_target linux amd64 linux-amd64

chmod +x "$PKG/bin/tasklight.js"

echo "built @tasklight/cli $VERSION vendor binaries"
