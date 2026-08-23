#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PACKAGE_JSON="$ROOT/npm/tasklight-cli/package.json"
PACKAGE_LOCK="$ROOT/npm/tasklight-cli/package-lock.json"
CHANGELOG="$ROOT/CHANGELOG.md"
REF_TYPE="${1:-${GITHUB_REF_TYPE:-}}"
REF_NAME="${2:-${GITHUB_REF_NAME:-}}"

read -r package_version lock_version < <(
  node - "$PACKAGE_JSON" "$PACKAGE_LOCK" <<'NODE'
const packageJson = require(process.argv[2]);
const packageLock = require(process.argv[3]);
process.stdout.write(`${packageJson.version} ${packageLock.version}\n`);
NODE
)

if [[ "$package_version" != "$lock_version" ]]; then
  echo "package.json version ($package_version) does not match package-lock.json ($lock_version)." >&2
  exit 1
fi

if [[ ! "$package_version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
  echo "Package version must use X.Y.Z format; found: $package_version" >&2
  exit 1
fi

if [[ "$REF_TYPE" == "tag" ]]; then
  if [[ ! "$REF_NAME" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
    echo "Release tag must use vX.Y.Z format; found: $REF_NAME" >&2
    exit 1
  fi

  expected_tag="v$package_version"
  if [[ "$REF_NAME" != "$expected_tag" ]]; then
    echo "Release tag ($REF_NAME) does not match package version ($package_version)." >&2
    exit 1
  fi
else
  echo "No release tag supplied; validating v$package_version as a dry run."
fi

escaped_version="${package_version//./\\.}"
if ! grep -Eq "^## \\[$escaped_version\\]([[:space:]]|$)" "$CHANGELOG"; then
  echo "CHANGELOG.md is missing a '## [$package_version]' release heading." >&2
  exit 1
fi

echo "Release metadata is valid for v$package_version."
