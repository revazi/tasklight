# Tasklight Brand Assets — v2.1

A flat, high-contrast identity built around one consistent terminal-prompt and notification-beacon glyph.

## Runtime assets

These files are embedded or packaged by the Tasklight codebase:

- `tasklight-app-icon-1024.png` — default notification icon embedded by `assets.go` and used by supported notification providers.
- `Tasklight.icns` — macOS icon source embedded for the fallback sender app and copied into native helpers as the cache-safe `Tasklight-v2.icns` resource.

## Repository assets

- `tasklight-repo-banner-1600x640.png` — README and repository banner.
- `tasklight-github-avatar-1024.png` — avatar with a GitHub-safe circular crop.
- `tasklight-mark-transparent-1024.png` — standalone mark for dark backgrounds.

## Vector sources

- `tasklight-repo-banner.svg`
- `tasklight-app-icon.svg`
- `tasklight-mark-light.svg`
- `tasklight-mark-dark.svg`
- `tasklight-mark-mono.svg`

## Small icons

PNG exports at 16, 32, 48, 64, 128, 256, and 512 pixels are under `icons/`.

The archive preview, source `.iconset`, macOS metadata, and original archive are generation inputs or duplicates and are intentionally not retained in the repository.

## Palette

- Ink: `#111111`
- Paper: `#F4F0E6`
- Signal: `#FFD43B`
- Success: `#35B979`
- Failure: `#E95B5B`
