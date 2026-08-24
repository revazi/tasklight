# Native macOS Notification Helper

The npm package includes `Tasklight.app`, a small Swift helper built from [`helpers/macos/TasklightNotifier/TasklightNotifier.swift`](../helpers/macos/TasklightNotifier/TasklightNotifier.swift). It gives notifications the Tasklight bundle identity and icon and receives notification click callbacks. Tasklight falls back to `terminal-notifier` and then `osascript` when the helper is unavailable.

## Notification permission and registration

The helper uses bundle identifier `dev.tasklight.notifier`. macOS owns notification authorization for that identity. This stable notifier identity was introduced in v0.2.1 because macOS continued serving the pre-v2 icon from the old `dev.tasklight.Tasklight` identity even after the helper, bundle version, registration, and on-disk icon had been updated. The terminal-notifier fallback uses the separate `dev.tasklight.sender.v2` identity so the native helper and fallback app cannot compete for one LaunchServices registration.

Run:

```bash
tasklight doctor
tasklight doctor --focus
```

On macOS, the report verifies the selected helper executable, code signature, hardened-runtime flag, quarantine state, LaunchServices registration, notification authorization, alert setting, and sound setting. The authorization check does not request permission or send a notification.

If authorization is `not-determined`, send the first Tasklight notification and accept the macOS prompt. Upgrading from v0.2.0 to v0.2.1 may show this prompt once because the cache-safe notifier identity is new. If authorization is `denied`, open **System Settings → Notifications → Tasklight** and enable notifications. Re-run `tasklight doctor` afterward.

### Stale notification icon

The helper bundled in the globally installed npm package supplies the notification icon; editing or rebuilding the repository does not update that installed package. Check both the executable and selected helper before troubleshooting caches:

```bash
command -v tasklight
tasklight --version
tasklight doctor
```

Upgrade `@tasklight/cli` when a new brand release is published. The upgraded binary replaces its cached PNG by content, the fallback sender carries a new brand bundle version, and the native helper carries the package version. Brand v2 also uses new native-helper and fallback-sender bundle identities plus the `Tasklight-v2.icns` resource name so LaunchServices and iconservices cannot reuse the pre-v2 notification icon. Running an older globally installed Tasklight can restore its older embedded PNG to the same cache path. Avoid deleting system notification databases as a first step.

## Process lifecycle

The helper has bounded waits and does not run as a daemon:

- A notification without a click command exits after a one-second delivery grace period.
- A notification with a click command waits for a response for at most one hour by default, then exits.
- A notification click starts the captured focus command, completes the callback, and exits.
- A notification-center relaunch without arguments waits up to 30 seconds for the pending response, then exits.
- Authorization and notification-delivery API calls each fail after five seconds rather than waiting indefinitely.

The build check exercises a notification-free `self-test` mode and fails if the bounded timeout process does not exit within five seconds.

Manual lifecycle checks for a source build:

```bash
make macos-helper
HELPER=./bin/Tasklight.app/Contents/MacOS/TasklightNotifier

# Does not request authorization or display a notification.
"$HELPER" doctor --timeout 3
"$HELPER" self-test --timeout 0.1

# Visible manual timeout test; do not click and confirm it exits after 5 seconds.
time "$HELPER" notify --message "Tasklight timeout test" --timeout 5

# Visible click test.
rm -f /tmp/tasklight-click-test
"$HELPER" notify \
  --message "Click to test Tasklight" \
  --click-command 'touch /tmp/tasklight-click-test' \
  --timeout 30
```

After the click test, `/tmp/tasklight-click-test` should exist and the helper process should have exited.

## Debug logging

Debug logging is disabled by default. Enable it only for a reproduction by setting `TASKLIGHT_FOCUS_DEBUG` to `1`, `true`, `yes`, or `on` on the command that sends the notification:

```bash
TASKLIGHT_FOCUS_DEBUG=1 tasklight notify --message "Focus test"
TASKLIGHT_FOCUS_DEBUG=1 tasklight doctor --focus
```

Logs are stored under the user cache directory:

- `~/Library/Caches/tasklight/native-helper.log`
- `~/Library/Caches/tasklight/focus/focus.log`

The native log is mode `0600`, is truncated after it grows beyond 1 MiB, and records lifecycle metadata rather than notification title/body text or the click command. Generated focus scripts and their directory are mode `0700`; the focus execution log is created with a restrictive umask and can include the generated shell command and terminal/tmux identifiers. Review both files before sharing them. Unset the variable after troubleshooting.

## Signing and notarization decision

### npm distribution for v0.2.x

Ad hoc signing is sufficient for the helper bundled inside the npm package for v0.2.x:

- The build removes extended attributes, enables the hardened runtime, ad hoc signs the final app bundle, and fails unless `codesign --verify --deep --strict` succeeds.
- Package construction and installation are tested on macOS.
- The npm tarball is published through trusted OIDC publishing with npm provenance and no long-lived npm token.
- The helper is not represented as a separately downloadable, Apple-identified application.

An ad hoc signature protects bundle integrity after signing but does **not** establish the publisher's Apple identity and is not a substitute for notarization. Users should obtain the npm package from the documented `@tasklight/cli` package.

### Future standalone distribution

A future standalone `.app`, archive, installer, or direct-download release should not reuse the ad hoc policy. Before offering one, the release process must:

1. Sign the helper with a Developer ID Application certificate and the hardened runtime.
2. Use a secure timestamp and keep certificate/notary credentials outside the repository.
3. Submit the final distribution artifact with `xcrun notarytool`.
4. Staple the successful notarization ticket where the artifact format supports it.
5. Verify with `codesign --verify --deep --strict` and `spctl --assess` on a clean macOS system.
6. Document certificate rotation and CI secret-recovery procedures.

This decision should be revisited as part of any Homebrew or direct-download distribution work.
