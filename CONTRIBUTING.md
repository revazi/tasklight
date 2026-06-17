# Contributing

Thanks for your interest in Tasklight.

Tasklight is early-stage. The main priority is keeping the CLI small, predictable, and safe.

## Development setup

Requirements:

- Go 1.19 or newer
- macOS or Linux

Common checks:

```bash
go test ./...
go vet ./...
go build -o bin/tasklight ./cmd/tasklight
./bin/tasklight doctor
```

## Guidelines

- Preserve child command behavior: live output, stdin forwarding, Ctrl+C behavior, and exit code.
- Keep notification failures non-fatal for `tasklight run`.
- Do not add telemetry.
- Do not store command output by default.
- Avoid shell execution unless the user explicitly asks for shell behavior.
- Keep platform-specific behavior behind small provider interfaces.

## Before opening a PR

Run:

```bash
go test ./...
go vet ./...
go build -o bin/tasklight ./cmd/tasklight
```

If your change affects notifications, also run:

```bash
./bin/tasklight notify --subtitle "✅ Test" --message "Tasklight notification test"
./bin/tasklight doctor
```
