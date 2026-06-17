# Security Policy

Tasklight is a local developer tool. It runs commands you explicitly ask it to run and sends local desktop notifications.

## Supported versions

Tasklight is pre-1.0. Security fixes will target the latest commit/release unless otherwise stated.

## Reporting a vulnerability

Please report security issues privately by contacting the maintainer rather than opening a public issue.

If GitHub private vulnerability reporting is enabled for this repository, use that. Otherwise, contact the maintainer through GitHub: https://github.com/revazi

## Security principles

- No telemetry.
- No cloud service.
- No command output upload.
- No command output storage by default.
- Default command execution avoids an implicit shell.
- Notification provider failures should not alter child command exit codes.
