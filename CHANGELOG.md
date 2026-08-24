# Changelog

## Unreleased

## [0.2.1] - 2026-08-24

### Fixed

- Force macOS to load the v2 notification icon by using a cache-safe bundle identity and icon resource name, and prevent npm package builds from registering temporary helper bundles.

## [0.2.0] - 2026-08-24

### Added

- `tasklight doctor --focus` diagnostics for notification providers, terminal/tmux targets, generated focus actions, and debug log paths.
- `tasklight run --idle <duration>` notifications for running tasks that stop producing stdout or stderr.
- `tasklight run --match <regexp>` attention notifications for matching stdout or stderr lines.
- Optional global and project-local TOML defaults for run and notification behavior.
- Linux desktop, D-Bus notification service, action-capability, and optional focus-tool diagnostics.
- Opt-in best-effort Linux notification actions that return to a captured tmux target.
- Native macOS helper signing, authorization, registration, quarantine, and lifecycle diagnostics.
- Tasklight v2.1 brand identity across repository and notification assets.
- Homebrew tap distribution with a source-built native macOS helper.

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
