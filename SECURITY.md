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

## GitHub Actions token policy

Workflow-level `GITHUB_TOKEN` permissions default to none. Every job declares only the permissions it needs:

- Repository checkout jobs receive `contents: read`.
- CodeQL and Scorecard SARIF upload jobs receive `security-events: write`.
- Scorecard publication and npm trusted publishing receive `id-token: write`.
- The tag-only GitHub release job receives `contents: write`, which the GitHub Releases API requires. That job runs only the fixed `gh release create` command after the separately verified, approval-gated npm publish job; it does not check out or execute repository code and does not run third-party actions.

The `contents: write` grant is intentionally isolated from checkout, builds, package publication, and other repository-controlled code. Removing it would prevent GitHub release creation; replacing it with a broader personal access token would weaken the security model.
