# Changelog

## Unreleased

Initial public version of Tasklight.

### Added

- `tasklight run -- <command>` command wrapper.
- Live stdout/stderr streaming and stdin forwarding.
- Child exit-code preservation.
- `tasklight notify` for direct notifications from scripts/integrations.
- `tasklight doctor` diagnostics.
- macOS notifications via `osascript` with optional `terminal-notifier` enhancements.
- Linux notifications via `notify-send`.
- Bundled Tasklight notification icon and macOS sender helper registration.
- Best-effort app activation and tmux pane return.
- Separate `pi-tasklight` integration package support through `tasklight notify`.
