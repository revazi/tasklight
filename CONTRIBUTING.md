# Contributing

Thanks for your interest in Tasklight.

Tasklight is early-stage. The main priority is keeping the CLI small, predictable, and safe.

## Maintenance and review policy

Tasklight is actively maintained as a pre-1.0, single-maintainer project.

### Triage and support

- Public bug reports are triaged on a best-effort basis, with an initial acknowledgement target of 30 days. Reproducible regressions and supported-platform breakage take priority over enhancements.
- Private vulnerability reports follow the process and 14-day initial response target in [SECURITY.md](SECURITY.md).
- The supported line is the latest release and current `main`. Pre-1.0 releases may make breaking changes, and older releases do not receive routine backports. A critical security fix may be backported when practical.
- Supported release platforms are macOS and Linux on arm64 and amd64. Other source builds are community-supported unless documented otherwise.
- If active maintenance stops, the maintainer will update the repository status or archive the repository rather than continuing to imply active support.

### Change control

- Changes to `main` go through pull requests. Branch protection is enforced for administrators, requires branches to be current, and blocks merging until configured CI and CodeQL checks pass. Direct pushes, force pushes, and branch deletion are blocked.
- External contributions receive maintainer review before merge. Review considers behavior, tests, security boundaries, platform impact, compatibility, and documentation.
- Maintainer-authored changes should seek independent human review when available for security fixes, release/publishing workflows, GitHub token permissions, shell-command construction, and native helper changes.
- Because requiring one approval would deadlock a single-maintainer repository, routine maintainer-authored pull requests may merge without an approving review after all required checks pass. Automated checks are a merge gate, not a claim of independent human review, and sham or automated approvals are not used to improve metrics.
- Pull requests are normally squash-merged after their validation and review requirements are satisfied. Material review findings should be resolved or explicitly documented before merge.

## Development setup

Requirements:

- Go 1.19 or newer
- Node.js 18 or newer for npm package checks
- macOS or Linux

Common checks:

```bash
make check
make fuzz-smoke
./bin/tasklight doctor
```

`make fuzz-smoke` runs bounded fuzzing for shell command construction and CLI argument parsing. Override the per-target duration with `FUZZTIME=30s make fuzz-smoke`.

Package checks:

```bash
make npm-package
make package-smoke
```

`make npm-package` should be run on macOS for publish-ready artifacts because the npm package includes the native macOS notification helper.

## Release process

Releases are built from `vX.Y.Z` tags by [the release workflow](.github/workflows/release.yml). The workflow verifies the tag and changelog, runs the repository checks, builds the native macOS helper, smoke-tests the packed npm package, publishes with npm provenance, and creates a GitHub release with generated notes.

Release notes must identify any publicly known runtime vulnerability fixed by the release, including its CVE or advisory identifier when one exists.

Before tagging a release:

1. Update the version in both `npm/tasklight-cli/package.json` and `npm/tasklight-cli/package-lock.json`.
2. Move the relevant changelog items under a `## [X.Y.Z] - YYYY-MM-DD` heading.
3. Merge the release changes to `main` through a pull request.
4. Run the **Release** workflow manually on that commit. Manual dispatch is always a dry run: it performs verification and `npm publish --dry-run`, but cannot publish or create a GitHub release.
5. Create and push an annotated tag only after the dry run passes:

   ```bash
   git switch main
   git pull --ff-only
   git tag -a vX.Y.Z -m "vX.Y.Z"
   git push origin vX.Y.Z
   ```

The publish job targets the protected GitHub environment named `npm`. Configure that environment with required reviewers. In npm package settings, configure a trusted publisher for repository `revazi/tasklight`, workflow `release.yml`, and environment `npm`. The workflow intentionally uses short-lived OIDC credentials and does not read a long-lived `NPM_TOKEN` secret.

For a local metadata check and package dry run, use:

```bash
./scripts/check-release.sh
make check
make npm-package
make package-smoke
npm publish ./npm/tasklight-cli --dry-run --access public
```

## Guidelines

- Preserve child command behavior: live output, stdin forwarding, Ctrl+C behavior, and exit code.
- Keep notification failures non-fatal for `tasklight run`.
- Do not add telemetry.
- Do not store command output by default.
- Avoid shell execution unless the user explicitly asks for shell behavior.
- Keep platform-specific behavior behind small provider interfaces.
- Add or update automated tests for bug fixes and significant behavior changes; add fuzz seeds or targets for parser and command-construction boundaries when appropriate.

## Before opening a PR

Run:

```bash
make check
```

If your change affects notifications, also run:

```bash
./bin/tasklight notify --subtitle "✅ Test" --message "Tasklight notification test"
./bin/tasklight doctor
```
