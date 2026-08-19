# Changelog

## Unreleased

Initial public version of Tasklight.

### Added

- `tasklight run -- <command>` command wrapper.
- Live stdout/stderr streaming and stdin forwarding.
- Child exit-code preservation.
- `tasklight notify` for direct notifications from scripts/integrations.
- `tasklight doctor` diagnostics.
- macOS notifications via bundled native `Tasklight.app`, with optional `terminal-notifier` and `osascript` fallbacks.
- Linux notifications via `notify-send`.
- Bundled Tasklight notification icon and macOS sender helper registration.
- iTerm2 + tmux click-to-return support with best-effort app activation elsewhere.
- Separate `pi-tasklight` integration package support through `tasklight notify`.
