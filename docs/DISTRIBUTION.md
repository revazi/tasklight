# Distribution

Tasklight uses two supported installation channels beginning with v0.2.0.

## npm

The npm package is the primary prebuilt distribution:

```bash
npm install -g @tasklight/cli
```

It contains Go binaries for macOS/Linux on arm64/amd64 and the ad hoc signed native macOS notification helper. npm publication uses trusted OIDC publishing and provenance.

## Homebrew tap

The supported Homebrew command is:

```bash
brew install revazi/tap/tasklight
```

The formula is maintained in `revazi/homebrew-tap`. It builds the tagged Tasklight source with Homebrew's Go toolchain. On macOS it also builds the native helper locally with the installed Xcode toolchain, applies the same hardened-runtime/ad hoc checks, and installs a wrapper that selects that helper. Linux builds install the CLI and use the documented `notify-send` dependency.

Building the helper locally avoids presenting an unnotarized, directly downloaded `.app` as an Apple-identified application. The future standalone signing/notarization requirements remain documented in [MACOS_HELPER.md](MACOS_HELPER.md).

## Release formula flow

For a version tag, the release workflow:

1. Verifies package, tag, and changelog metadata.
2. Runs repository, package, native-helper, and installation checks.
3. Creates a source archive from the checked-out immutable tag and records its SHA-256.
4. Generates and validates a versioned `tasklight.rb` formula that uses that release asset.
5. Publishes npm through the protected environment.
6. Creates the matching GitHub release with the source archive, `tasklight.rb`, and `source-checksum.txt` attached.

After the GitHub release exists, copy the generated formula into `Formula/tasklight.rb` in `revazi/homebrew-tap`, review its source URL and checksum, and test:

```bash
brew update
brew upgrade revazi/tap/tasklight
# or, for a clean validation
brew uninstall tasklight
brew install revazi/tap/tasklight
tasklight --version
tasklight doctor
brew test revazi/tap/tasklight
```

The Git tag, GitHub release, npm `latest`, formula version, and formula source URL must all agree before announcing a release. After updating the tap, run:

```bash
./scripts/verify-release-state.sh X.Y.Z
```

## Direct downloads and install scripts

Tasklight does not currently publish a `curl | sh` installer. Avoiding an install script keeps executable placement, upgrades, and uninstallation under npm or Homebrew package-manager control.

GitHub automatically provides source archives for each release. Prebuilt standalone macOS app archives are intentionally deferred until Developer ID signing and notarization are available. See [MACOS_HELPER.md](MACOS_HELPER.md).
