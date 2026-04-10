# Quick App Switcher

A fast, lightweight Windows application switcher written in Go. Quick App Switcher provides a Most Recently Used (MRU) based overlay to quickly cycle through open windows, serving as an efficient alternative to the built-in Alt+Tab switcher.

## Features

- **MRU-Based Switching:** Intelligently tracks window focus and orders your switch targets based on your most recently used applications.
- **Native Performance:** Built using pure Go and native Windows (Win32) APIs without heavy UI frameworks, ensuring minimal memory footprint and instant responsiveness.
- **System Tray Integration:** Runs quietly in the background with a system tray icon for easy management and exit.
- **Custom Overlay:** Displays a lightweight, non-intrusive on-screen overlay showing the available applications during a switch session.
- **Single Instance:** Prevents multiple instances from running concurrently.

## Requirements

- **Operating System:** Windows
- **Go Version:** Go 1.26 or higher (for building from source)

## Building from Source

To build the application, ensure you have Go installed, then run the following command from the root of the repository:

```bash
go build -o quick-app-switcher.exe ./cmd/quick-app-switcher
```

Alternatively, you can use the provided PowerShell build script if available:

```powershell
.\build-run.ps1
```

## Usage

1. Run the compiled `quick-app-switcher.exe`.
2. The application will start in the background, and a new icon will appear in your Windows System Tray.
3. Use the keyboard shortcut (typically `Alt` + `Tab`) to invoke the overlay and cycle through your open windows. Release the modifier key (`Alt`) to switch to the selected application.
4. To exit, right-click the Quick App Switcher icon in the system tray and select **Exit**.

## Architecture

The application is structured into several internal packages for clear separation of concerns:

- `cmd/quick-app-switcher`: The main entry point.
- `internal/app`: Core application logic, window message loop, and state management.
- `internal/events`: Foreground window watchers and event tracking.
- `internal/input`: Global keyboard hooks (e.g., intercepting Tab/Alt).
- `internal/mru`: Most Recently Used list logic.
- `internal/ui`: Rendering of the system tray and on-screen overlay.
- `internal/win32` & `internal/windows`: Low-level Win32 API bindings and window management.
