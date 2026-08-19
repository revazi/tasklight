#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

scripts/check-format.sh

npm --prefix npm/tasklight-cli ci --ignore-scripts --no-audit --no-fund
npm --prefix npm/tasklight-cli audit --omit=dev --audit-level=high

go mod tidy
if ! git diff --exit-code -- go.mod go.sum >/dev/null; then
  echo "go mod tidy changed tracked module files; commit the result." >&2
  git diff -- go.mod go.sum >&2
  exit 1
fi

go test ./...
go test -race ./...
go vet ./...
make build
