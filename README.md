# Quick App Switcher

Quick App Switcher is a lightweight Windows `Alt+Tab` replacement written in Go. It keeps its own MRU window order, shows a fast native overlay, and stays out of the way in the system tray while it runs in the background.

## What It Does

- Replaces the native `Alt+Tab` switching flow with an MRU-driven window switcher.
- Starts a session on the first `Tab` press while `Alt` is held.
- Preselects the previous window first, so repeated `Alt+Tab` presses quickly bounce between your two most recent windows.
- Shows a centered overlay on the active monitor with app icons, window previews, and labels.
- Activates the selected window on `Alt` release.
- Tracks foreground-window changes to keep MRU ordering current outside switching sessions.
- Runs as a single instance and exposes control actions through a tray icon.

## Current Features

- Native Win32 implementation in pure Go.
- Fast overlay window created up front at startup.
- App icons cached in memory for faster repainting.
- Optional live window thumbnails in the switcher overlay.
- Native settings window.
- TOML config file stored in your user profile.
- Tray actions for `Settings`, `Open Config File`, and `Close`.
- Optional launch on Windows startup via the current user's `Run` registry key.
- Current-virtual-desktop filtering when the Windows virtual desktop API is available.
- Explorer/taskbar restart handling for the tray icon.

## Settings And Config

Quick App Switcher stores its config at:

`%USERPROFILE%\.config\quick-app-switcher\config.toml`

If the file does not exist yet, the app creates it with defaults on first launch.

Current config options:

```toml
show_thumbnails = true
launch_on_startup = false
instant_switch_preview = true
```

What they do:

- `show_thumbnails`: enables live window thumbnails in the overlay. If disabled, the app still shows icons and labels.
- `launch_on_startup`: adds or removes the app from `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`.
- `instant_switch_preview`: when `true`, the selected window is activated as you cycle. When `false`, the current window keeps focus until you release `Alt`.

Settings changed through the built-in settings window are applied immediately and saved back to `config.toml`; no restart is required.

The tray menu also includes `Open Config File`, which opens the config path directly in Windows so you can edit it manually.

## Usage

1. Start `quick-app-switcher.exe`.
2. The app stays in the background and adds a tray icon.
3. Hold `Alt` and press `Tab` to start switching.
4. Keep pressing `Tab` to move forward through the MRU list.
5. Release `Alt` to commit the current selection.
6. Press `Esc` during a switching session to cancel it.

Tray behavior:

- Left-click the tray icon to open Settings.
- Right-click the tray icon to open the tray menu.

## Limitations

- Windows only.
- The switcher currently cycles forward only; reverse cycling with `Shift+Alt+Tab` is not implemented.
- The candidate list is limited to windows on the current virtual desktop when that information is available.
- The overlay is keyboard-driven only.

## Requirements

- Windows
- Go 1.26 or newer if you are building from source

## Build

Build the executable from the repo root:

```powershell
go build -o quick-app-switcher.exe ./cmd/quick-app-switcher
```

For a build-and-run loop during development, use:

```powershell
.\build-run.ps1
```

That script builds the app into `.gotmp\quick-app-switcher.exe` and launches it.

## Project Layout

- `cmd/quick-app-switcher`: program entry point
- `internal/app`: application wiring, message loop, session flow, tray commands, and settings handling
- `internal/config`: config loading, validation, normalization, and saving
- `internal/events`: foreground-window event watcher
- `internal/input`: low-level keyboard hook for `Alt`, `Tab`, and cancel handling
- `internal/mru`: MRU ordering logic
- `internal/session`: switch-session state machine
- `internal/startup`: Windows startup registration
- `internal/ui`: overlay, layout, tray, and settings window UI
- `internal/win32`: Win32 bindings and helpers
- `internal/windows`: window discovery, filtering, activation, desktop checks, icon loading, and thumbnails
