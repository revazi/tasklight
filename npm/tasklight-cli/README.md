# @tasklight/cli

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
tasklight notify --subtitle "✅ Done" --message "Finished"
```

## Platform support

This package currently bundles prebuilt Tasklight binaries for:

- macOS arm64
- macOS x64
- Linux arm64
- Linux x64

Windows is not packaged yet.

## macOS optional enhancement

Tasklight works on macOS with built-in `osascript` notifications. For better notification identity, custom icon support, and click-to-focus behavior, install:

```bash
brew install terminal-notifier
```

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

## Development

From the Tasklight repository root:

```bash
npm --prefix npm/tasklight-cli run build:vendor
npm --prefix npm/tasklight-cli run test:local
npm --prefix npm/tasklight-cli run pack:check
```

## Source

https://github.com/revazi/tasklight
