#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PKG="$ROOT/npm/tasklight-cli"
TMP="$(mktemp -d)"
TARBALL=""
TARBALL_PATH=""

cleanup() {
  if [[ -n "$TARBALL" && -f "$PKG/$TARBALL" ]]; then
    rm -f "$PKG/$TARBALL"
  fi
  rm -rf "$TMP"
}
trap cleanup EXIT

cd "$PKG"
pack_result="$(npm pack --json)"
read -r TARBALL expected_integrity < <(
  node -e '
    const result = JSON.parse(process.argv[1]);
    if (!Array.isArray(result) || result.length !== 1) process.exit(1);
    process.stdout.write(`${result[0].filename} ${result[0].integrity}\n`);
  ' "$pack_result"
)

if [[ -z "$TARBALL" || "$TARBALL" == */* || "$TARBALL" != *.tgz ]]; then
  echo "npm pack returned an invalid tarball name: $TARBALL" >&2
  exit 1
fi
TARBALL_PATH="$PKG/$TARBALL"

actual_integrity="$(
  node -e '
    const crypto = require("node:crypto");
    const fs = require("node:fs");
    const digest = crypto.createHash("sha512").update(fs.readFileSync(process.argv[1])).digest("base64");
    process.stdout.write(`sha512-${digest}`);
  ' "$TARBALL_PATH"
)"
if [[ "$actual_integrity" != "$expected_integrity" ]]; then
  echo "Tarball integrity mismatch for $TARBALL" >&2
  exit 1
fi

# The package source is the freshly generated local tarball above, not a
# registry dependency. An isolated cache plus --offline prevents registry
# substitution while npm exec still exercises package extraction and bin shims.
(
  cd "$TMP"
  export npm_config_cache="$TMP/npm-cache"
  export npm_config_update_notifier=false

  npm exec --offline --yes --ignore-scripts --package="$TARBALL_PATH" -- tasklight --version
  npm exec --offline --yes --ignore-scripts --package="$TARBALL_PATH" -- tasklight doctor
)
