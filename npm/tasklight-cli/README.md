# @tasklight/cli

[![npm version](https://img.shields.io/npm/v/%40tasklight%2Fcli.svg)](https://www.npmjs.com/package/@tasklight/cli)
[![npm downloads](https://img.shields.io/npm/dm/%40tasklight%2Fcli.svg)](https://www.npmjs.com/package/@tasklight/cli)
[![CI](https://github.com/revazi/tasklight/actions/workflows/ci.yml/badge.svg)](https://github.com/revazi/tasklight/actions/workflows/ci.yml)
[![CodeQL](https://github.com/revazi/tasklight/actions/workflows/codeql.yml/badge.svg)](https://github.com/revazi/tasklight/actions/workflows/codeql.yml)
[![license: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../../LICENSE)

npm package for the Tasklight CLI.

Tasklight notifies you when long-running developer tasks finish, fail, or need attention.

## Install

```bash
npm install -g @tasklight/cli
```

Then:

```bash
tasklight --version
tasklight doctor
tasklight run -- pnpm test
tasklight run --idle 5m -- your-agent "continue implementation"
tasklight run --match 'approve|waiting|failed' -- your-agent "implement feature"
tasklight notify --subtitle "✅ Done" --message "Finished"
```

`--idle` sends one notification when a running command produces no stdout or stderr for the configured duration. Output resets the timer and re-arms detection without stopping the command.

`--match` accepts a Go regular expression and sends one attention notification for the first matching stdout or stderr line. Invalid expressions fail before the command starts.

## Configuration

Optional defaults can be stored in `$XDG_CONFIG_HOME/tasklight/config.toml` (or `~/.config/tasklight/config.toml`) and overridden by a project-local `.tasklight.toml` and then CLI flags:

```toml
[run]
activate_app = "iTerm2"
sound = false
idle = "5m"
match = "approve|waiting|failed"

[notify]
activate_app = "iTerm2"
sound = true
```

Invalid configuration fails before Tasklight runs the command or sends a notification.

## Platform support

This package currently bundles prebuilt Tasklight binaries for:

- macOS arm64
- macOS x64
- Linux arm64
- Linux x64

Windows is not packaged yet.

## macOS notifications

This npm package bundles a tiny native `Tasklight.app` notification helper for proper Tasklight notification identity, custom icon support, and reliable click behavior. The helper is ad hoc signed with the hardened runtime and verified during packaging; it is not represented as an Apple-notarized standalone application.

After installing on macOS, open **System Settings → Notifications** and enable **Tasklight** once. This is required after installing v0.2.1 because its cache-safe notification identity is new. Run `tasklight doctor` afterward to confirm authorization.

If the native helper is unavailable, Tasklight can use `terminal-notifier` as an optional fallback:

```bash
brew install terminal-notifier
```

Without the native helper or `terminal-notifier`, Tasklight falls back to built-in `osascript` notifications.

If click-to-focus returns to the wrong terminal or tmux pane, generate a pasteable diagnostic report:

```bash
tasklight doctor --focus
```

The report shows the selected provider, helper signature and quarantine state, notification authorization, terminal/tmux target, generated focus action, and opt-in debug log paths. If authorization is denied, enable Tasklight under **System Settings → Notifications**. Set `TASKLIGHT_FOCUS_DEBUG=1` only while reproducing a focus issue, and review generated focus logs before sharing them.

See the [native macOS helper guide](https://github.com/revazi/tasklight/blob/main/docs/MACOS_HELPER.md) for lifecycle timeouts, manual checks, logging details, and the future Developer ID/notarization plan.

## Linux notification dependency

Linux desktop notifications use `notify-send`.

```bash
# Ubuntu/Debian
sudo apt install libnotify-bin

# Fedora
sudo dnf install libnotify

# Arch
sudo pacman -S libnotify
```

Linux notifications are informational by default. Inside tmux, `TASKLIGHT_LINUX_ACTIONS=1` enables a best-effort **Select tmux target** action when `notify-send` and the notification daemon advertise action support. Tasklight does not invoke optional X11 window-control tools. Run `tasklight doctor` for desktop, D-Bus, notification-service, action, optional-tool, and tmux diagnostics; use `tasklight doctor --focus` for the generated tmux target details.

## Development

From the Tasklight repository root:

```bash
make npm-package
make package-smoke
```

Run these package checks on macOS for publish-ready artifacts because the tarball includes the native macOS notification helper.

Package-local commands are also available:

```bash
npm --prefix npm/tasklight-cli run build:vendor
npm --prefix npm/tasklight-cli run check
npm --prefix npm/tasklight-cli run pack:check
npm --prefix npm/tasklight-cli run package:smoke
```

## Source

https://github.com/revazi/tasklight
