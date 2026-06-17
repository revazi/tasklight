# Tasklight Brand Assets

Canonical assets used by the Tasklight codebase and repository.

## Runtime notification assets

These files are embedded into the Tasklight Go binary by `assets.go`:

- `tasklight-app-icon-1024.png` — default notification icon used by `tasklight notify` and Linux `notify-send`.
- `Tasklight.icns` — macOS app icon used for the local `Tasklight.app` notification sender helper.

## Repository assets

- `tasklight-repo-banner-1600x640.png` — README/repository banner.
- `tasklight-github-avatar-1024.png` — GitHub organization/repository avatar candidate.
- `tasklight-mark-transparent-1024.png` — transparent standalone Tasklight mark.

Raw generated asset exports should not be referenced by the codebase. Keep runtime/repository assets here instead.
