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
tasklight notify --subtitle "✅ Done" --message "Finished"
```

## Platform support

This package currently bundles prebuilt Tasklight binaries for:

- macOS arm64
- macOS x64
- Linux arm64
- Linux x64

Windows is not packaged yet.

## macOS notifications

This npm package bundles a tiny native `Tasklight.app` notification helper for proper Tasklight notification identity, custom icon support, and reliable click behavior.

If the native helper is unavailable, Tasklight can use `terminal-notifier` as an optional fallback:

```bash
brew install terminal-notifier
```

Without the native helper or `terminal-notifier`, Tasklight falls back to built-in `osascript` notifications.

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
