# @tasklight/cli

[![npm version](https://img.shields.io/npm/v/%40tasklight%2Fcli.svg)](https://www.npmjs.com/package/@tasklight/cli)
[![npm downloads](https://img.shields.io/npm/dm/%40tasklight%2Fcli.svg)](https://www.npmjs.com/package/@tasklight/cli)
[![CI](https://github.com/revazi/tasklight/actions/workflows/ci.yml/badge.svg)](https://github.com/revazi/tasklight/actions/workflows/ci.yml)
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
