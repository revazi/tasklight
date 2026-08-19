#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${1:-$(mktemp -d)}"
VERSION="${VERSION:-dev}"
LDFLAGS="-X github.com/revazi/tasklight/internal/cli.Version=$VERSION"

targets=(
  darwin/arm64
  darwin/amd64
  linux/arm64
  linux/amd64
)

mkdir -p "$OUT_DIR"

for target in "${targets[@]}"; do
  goos="${target%/*}"
  goarch="${target#*/}"
  output="$OUT_DIR/tasklight-$goos-$goarch"
  echo "building $goos/$goarch"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -ldflags "$LDFLAGS" -o "$output" "$ROOT/cmd/tasklight"
done

echo "cross-compiled binaries in $OUT_DIR"
