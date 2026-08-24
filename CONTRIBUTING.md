# Contributing

Thanks for your interest in Tasklight.

Tasklight is early-stage. The main priority is keeping the CLI small, predictable, and safe.

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
