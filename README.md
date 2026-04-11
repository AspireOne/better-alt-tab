# Better Alt Tab

Better Alt Tab is a fast, lightweight replacement for the default Windows `Alt+Tab` experience.

It is built for people who switch windows constantly and want something that feels immediate, predictable, and quiet. Instead of the heavier native switcher flow, Better Alt Tab keeps a tight MRU-based list, shows a clean overlay, and gets you to the next window with as little friction as possible.

## Why Use It

- Faster back-and-forth switching between your most recent windows.
- A simpler, more focused overlay that stays out of the way.
- Optional live thumbnails when you want visual previews.
- Native Windows behavior where it matters, without a big framework footprint.
- Small background utility with tray access and simple settings.

## Highlights

- MRU-first switching.
  Repeated `Alt+Tab` presses quickly bounce between the two windows you were just using.

- Immediate switching flow.
  The overlay appears as soon as you start cycling, and selection commits on `Alt` release.

- Optional instant switching.
  You can make Better Alt Tab switch to each newly selected window immediately as you cycle, or keep the current window active until you release `Alt`.

- Optional window thumbnails.
  When enabled, the overlay shows live previews. When disabled, it stays lean with icons and labels only.

- Simple built-in settings.
  Open Settings from the tray icon and change behavior without digging through menus or restarting the app.

- Config file access.
  If you prefer editing settings directly, the tray menu can open the config file for you.

- Startup option.
  You can have the app launch automatically with Windows.

- Single-instance background app.
  It stays in the tray, avoids duplicate instances, and is meant to just run quietly.

## What It Feels Like

Better Alt Tab is designed around one main use case: frequent keyboard-driven app switching.

When you hold `Alt` and press `Tab`, it starts from the previous window first, which makes the common "jump to the last thing I was using" case very fast. Keep pressing `Tab` to move through the rest of your recent windows, then release `Alt` to commit.

If you enable instant switching, Better Alt Tab activates the newly selected window on each `Tab` press while you are still holding `Alt`. If you disable instant switching, it only commits the switch once you release `Alt`, which keeps the current window active while you cycle through choices.

## Settings

Current options:

- `show_thumbnails`
- `launch_on_startup`
- `instant_switch_preview`

These can be changed from the built-in settings window. Saved settings apply immediately.

## Usage

1. Start `better-alt-tab.exe`.
2. Hold `Alt` and press `Tab` to begin switching.
3. Press `Tab` again to move forward through recent windows.
4. Release `Alt` to activate the selected window.
5. Press `Esc` to cancel an in-progress switch session.

Tray actions:

- Left-click: open Settings
- Right-click: open the tray menu
- Tray menu: `Settings`, `Open Config File`, `Close`

## Configuration File

The config file lives at:

`%USERPROFILE%\.config\better-alt-tab\config.toml`

Default config:

```toml
show_thumbnails = true
launch_on_startup = false
instant_switch_preview = true
```

## Limitations

- Windows only
- Forward cycling only for now; `Shift+Alt+Tab` style reverse cycling is not implemented
- The switch list is limited to windows on the current virtual desktop when that information is available

## Build From Source

If you want to build it yourself:

```powershell
go build -o better-alt-tab.exe ./cmd/better-alt-tab
```

For a quick development build-and-run loop:

```powershell
.\build-run.ps1
```

## Technical Notes

- Written in Go with native Win32 APIs
- Uses a tray icon, a pre-created overlay window, and a keyboard hook
- Keeps MRU ordering based on foreground window changes
- Stores configuration as TOML in your user profile
