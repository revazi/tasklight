#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-}"
REPOSITORY="${TASKLIGHT_REPOSITORY:-revazi/tasklight}"
TAP_REPOSITORY="${TASKLIGHT_TAP_REPOSITORY:-revazi/homebrew-tap}"
TAG="v$VERSION"

if [[ ! "$VERSION" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
  echo "usage: $0 <X.Y.Z>" >&2
  exit 2
fi

local_version="$(node -p "require('$ROOT/npm/tasklight-cli/package.json').version")"
[[ "$local_version" == "$VERSION" ]] || {
  echo "local package version is $local_version; expected $VERSION" >&2
  exit 1
}

git -C "$ROOT" rev-parse --verify "refs/tags/$TAG" >/dev/null
published_version="$(npm view @tasklight/cli version)"
[[ "$published_version" == "$VERSION" ]] || {
  echo "npm latest is $published_version; expected $VERSION" >&2
  exit 1
}

release_tag="$(gh release view "$TAG" --repo "$REPOSITORY" --json tagName,isDraft,isPrerelease --jq 'select(.isDraft == false and .isPrerelease == false) | .tagName')"
[[ "$release_tag" == "$TAG" ]] || {
  echo "published GitHub release does not match $TAG" >&2
  exit 1
}

TMP="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP"
}
trap cleanup EXIT

gh release download "$TAG" --repo "$REPOSITORY" --pattern tasklight.rb --dir "$TMP/release"
gh api "repos/$TAP_REPOSITORY/contents/Formula/tasklight.rb" --jq .content | tr -d '\n' | base64 --decode > "$TMP/tap-tasklight.rb"
cmp "$TMP/release/tasklight.rb" "$TMP/tap-tasklight.rb"
grep -F "releases/download/$TAG/tasklight-$TAG-source.tar.gz" "$TMP/tap-tasklight.rb" >/dev/null

printf 'verified Tasklight %s across local metadata, Git tag, GitHub release, npm, and Homebrew tap\n' "$VERSION"
