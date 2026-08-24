# OpenSSF Best Practices Self-Assessment

Tasklight uses the OpenSSF Best Practices passing criteria as a lightweight project checklist. This document records repository evidence and project-specific decisions; it is not an OpenSSF badge claim.

## Project basics

- [x] **Purpose and installation:** The [README](../README.md) describes the problem, installation methods, supported behavior, and CLI interface.
- [x] **Feedback and contributions:** GitHub issues accept bug reports and enhancements, and [CONTRIBUTING.md](../CONTRIBUTING.md) documents the pull-request process and validation expectations.
- [x] **Open-source license:** Tasklight is released under the OSI-approved [MIT License](../LICENSE).
- [x] **Public change history:** Source and interim changes are maintained in the public Git repository. Protected `main` requires pull requests, up-to-date branches, and required checks.
- [x] **English documentation:** Project documentation, issue templates, and contribution guidance are maintained in English.

## Supported platforms

The supported release targets are:

| Platform | Architectures | Notification provider |
| --- | --- | --- |
| macOS | arm64, amd64 | Bundled native helper, with `terminal-notifier` and `osascript` fallbacks |
| Linux | arm64, amd64 | `notify-send` |

Source builds require Go 1.19 or newer. npm installation requires Node.js 18 or newer. Platform capability differences and diagnostics are documented in the [README](../README.md).

## Reporting and maintenance

- [x] **Public reports:** Bugs and enhancement requests use the repository's searchable [GitHub issue tracker](https://github.com/revazi/tasklight/issues).
- [x] **Private vulnerability reports:** Security reports use [GitHub private vulnerability reporting](https://github.com/revazi/tasklight/security/advisories/new), as documented in [SECURITY.md](../SECURITY.md).
- [x] **Security response target:** The maintainer targets an initial private response within 14 days and prioritizes confirmed critical vulnerabilities.
- [x] **Supported versions:** Security fixes target the latest commit and latest release unless [SECURITY.md](../SECURITY.md) says otherwise.

## Build, test, and analysis

- [x] **Reproducible commands:** The Makefile and [CONTRIBUTING.md](../CONTRIBUTING.md) document formatting, tests, race detection, vetting, fuzz smoke tests, cross-compilation, and package checks.
- [x] **Continuous integration:** Pull requests run tests on macOS and Linux, race detection, vet, cross-compilation, package smoke tests, dependency review, and CodeQL.
- [x] **Tests accompany changes:** Significant behavior changes and bug fixes are expected to add or update automated tests.
- [x] **Static analysis:** `go vet`, CodeQL extended security queries, npm audit, and repository hygiene checks run before release.
- [x] **Dynamic analysis:** Go fuzz targets cover CLI/session parsing and shell command construction; race-enabled tests cover concurrent behavior.
- [x] **Dependency controls:** Go and npm dependencies are locked, GitHub Actions are pinned to commit SHAs, and Dependabot tracks updates.

## Releases and delivery

- [x] **Versioning:** User releases use Semantic Versioning and immutable `vX.Y.Z` Git tags.
- [x] **Release notes:** [CHANGELOG.md](../CHANGELOG.md) provides human-readable changes; the release workflow generates matching GitHub release notes. Release guidance requires naming fixed public vulnerabilities and their CVE or advisory identifiers when applicable.
- [x] **Provenance:** npm publication uses trusted publishing with short-lived OIDC credentials and npm provenance. No long-lived npm publishing token is stored.
- [x] **Artifact verification:** Release metadata, package contents, local tarball integrity, and installation are checked before publication.
- [x] **Transport security:** Source, dependencies, and release artifacts are delivered through HTTPS or authenticated Git transport.

## Project-specific applicability

Tasklight is a local CLI with no account system, network service, password storage, or cryptographic protocol. OpenSSF passing criteria concerning cryptographic key management, password hashing, and forward secrecy are therefore not applicable. Hashes used for package integrity and deterministic local filenames rely on standard library implementations rather than project-defined cryptography.

## Badge registration decision

- [ ] **OpenSSF Best Practices badge registration:** Tasklight is not currently registered at [bestpractices.dev](https://www.bestpractices.dev/), and no badge is claimed.

Registration requires a maintainer-authored external attestation and ongoing updates. It is deferred until the project has enough operating history to answer the criteria covering report response rates and maintenance over a 2–12 month period with evidence. Until then, the Scorecard `CII-Best-Practices` finding is accepted as an accurate indication that no external badge has been earned; this self-assessment remains the repository's actionable checklist.
