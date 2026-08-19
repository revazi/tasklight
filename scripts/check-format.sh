#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

status=0

go_files="$(find . \
  -path './.git' -prune -o \
  -path './.fallow' -prune -o \
  -path './bin' -prune -o \
  -path './npm/tasklight-cli/vendor' -prune -o \
  -type f -name '*.go' -print)"

if [[ -n "$go_files" ]]; then
  unformatted="$(printf '%s\n' "$go_files" | xargs gofmt -l)"
  if [[ -n "$unformatted" ]]; then
    echo "Go files need gofmt:" >&2
    printf '%s\n' "$unformatted" >&2
    status=1
  fi
fi

for script in scripts/*.sh; do
  [[ -e "$script" ]] || continue
  bash -n "$script"
done

node --check npm/tasklight-cli/bin/tasklight.js >/dev/null

git diff --check

exit "$status"
