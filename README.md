# Tasklight

Tasklight watches long-running developer tasks and lights up when they finish, fail, stall, or need input.

This repository currently implements the first CLI wrapper milestone: `tasklight run -- <command>` runs a command transparently, streams output live, forwards stdin, and returns the child command's exit code.

## Usage

```bash
# Run tests and keep the same exit code
tasklight run -- pnpm test

# Run Python tests
tasklight run -- pytest

# Run a coding-agent task
tasklight run -- pi "fix this failing test"

# Run from a specific directory
tasklight run --cwd frontend -- pnpm build
```

The `--` separator is required so Tasklight flags are not confused with child command flags.

## Development

This project uses Go.

```bash
# Run tests
go test ./...

# Run vet
go vet ./...

# Build a local binary
go build -o bin/tasklight ./cmd/tasklight

# Show help
go run ./cmd/tasklight --help
```

## Current scope

Implemented:

- `tasklight run -- <command>`
- live stdout/stderr streaming
- stdin forwarding
- child exit-code preservation
- `--cwd` and `--name` parsing for the `run` command

Planned next:

- macOS notifications on success/failure
- idle/stuck detection
- match-based attention detection
- optional terminal activation support
