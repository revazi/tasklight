# Changelog

## Unreleased

### Added

- `tasklight doctor --focus` diagnostics for notification providers, terminal/tmux targets, generated focus actions, and debug log paths.
- `tasklight run --idle <duration>` notifications for running tasks that stop producing stdout or stderr.
- `tasklight run --match <regexp>` attention notifications for matching stdout or stderr lines.
- Optional global and project-local TOML defaults for run and notification behavior.

## [0.1.1] - 2026-08-20

### Added

- Native macOS `Tasklight.app` notification helper bundled with the npm package.
- Tasklight notification identity, icon, click actions, and focus fallbacks.
- iTerm2 and tmux click-to-return support.
- Repository hygiene checks, CI, CodeQL, Scorecard, and Dependabot configuration.

## [0.1.0] - 2026-06-18

### Added

- `tasklight run -- <command>` command wrapper.
- Live stdout/stderr streaming and stdin forwarding.
- Child exit-code preservation.
- `tasklight notify` for direct notifications from scripts and integrations.
- `tasklight doctor` diagnostics.
- macOS notifications via optional `terminal-notifier` and `osascript` providers.
- Linux notifications via `notify-send`.
- npm distribution as `@tasklight/cli`.
