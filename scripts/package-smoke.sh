#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PKG="$ROOT/npm/tasklight-cli"
TMP="$(mktemp -d)"
TARBALL=""

cleanup() {
  if [[ -n "$TARBALL" && -f "$PKG/$TARBALL" ]]; then
    rm -f "$PKG/$TARBALL"
  fi
  rm -rf "$TMP"
}
trap cleanup EXIT

cd "$PKG"
TARBALL="$(npm pack --silent | tail -n 1)"

npm install \
  --prefix "$TMP" \
  --ignore-scripts \
  --no-audit \
  --no-fund \
  "$PKG/$TARBALL"

"$TMP/node_modules/.bin/tasklight" --version
"$TMP/node_modules/.bin/tasklight" doctor
