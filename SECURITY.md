# Security Policy

Tasklight is a local developer tool. It runs commands you explicitly ask it to run and sends local desktop notifications.

## Supported versions

Tasklight is pre-1.0. Security fixes will target the latest commit/release unless otherwise stated.

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability. Use [GitHub private vulnerability reporting](https://github.com/revazi/tasklight/security/advisories/new) so the report and follow-up remain private.

The maintainer targets an initial response within 14 days. If private vulnerability reporting is unavailable, contact the maintainer through [GitHub](https://github.com/revazi) to arrange a private channel without posting vulnerability details publicly.

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

Tasklight's lightweight OpenSSF criteria review and badge-registration decision are recorded in the [Best Practices self-assessment](docs/OPENSSF_BEST_PRACTICES.md).
