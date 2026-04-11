# Quick App Switcher

An alternative `Alt+Tab` app switcher for Windows, built in Go, with a hard focus on instant UI display and instant window switching.

## Feature Specification

- Replace the native Windows `Alt+Tab` behavior for switching between open application windows.
- Maintain switch order by most recently used window, matching the expected `Alt+Tab` mental model.
- On initial `Alt+Tab` press, preselect the second-most-recently-used window.
- While `Alt` remains held, each additional `Tab` press advances selection through the MRU window list.
- Releasing `Alt` immediately activates the currently selected window.
- Repeated `Alt+Tab` presses should rapidly toggle between the two most recently used windows.
- Display an on-screen switcher UI containing the icons of currently switchable open apps/windows.
- Keep the switcher UI visible only while `Alt` is held.
- Only include windows from the current Windows virtual desktop.
- Follow native Windows `Alt+Tab` behavior for which windows are considered switchable.
- If an app has multiple top-level windows, include each switchable window separately.
- The switcher is keyboard-only.
- `Shift+Alt+Tab` backward cycling is not supported.
- If the selected window disappears or becomes invalid before activation, activate the next valid window instead.
- Provide a system tray icon with a menu containing a `Close` action.
- Do not provide a fallback shortcut to the native Windows `Alt+Tab` UI.
- Do not implement auto-start at login in the initial version.
- The UI must appear instantly on `Alt+Tab`, with no perceptible startup or animation delay.
- Window activation on `Alt` release must happen instantly, with no perceptible delay.
- The implementation must prioritize low latency and responsiveness over visual complexity.
- The application will be implemented in Go.

## Notes To Make Explicit In Implementation

- The switch target is individual windows, not app processes.
- The set of included windows should match native Windows `Alt+Tab` behavior as closely as practical.
- Minimized windows are included if Windows can normally switch to them.
- Virtual desktop handling is intentionally fixed to the current desktop only for robustness; mirroring the native Windows cross-desktop `Alt+Tab` setting is out of scope for the initial version.
- The currently focused window is the first MRU item before `Alt+Tab` starts.
- If fewer than two switchable windows exist, no switch occurs and the UI may remain hidden.

## Implementation Outline

This outline should borrow the overall native-Windows shape from [_oss_window_switcher_example](./_oss_window_switcher_example): low-level keyboard hook, WinEvent-based foreground tracking, a pre-created layered overlay window, icon caching, and a tray icon. The main difference is that this project should be built around a per-window MRU cache rather than app grouping.

### Overall Approach

- Build this as a native Win32 background process in Go, not with a cross-platform GUI toolkit.
- Use raw Win32 APIs from Go so the hot path stays as small as possible.
- Lock the controller/UI goroutine to one OS thread with `runtime.LockOSThread`.
- Run the low-level keyboard hook on its own dedicated locked OS thread with its own message pump so UI work cannot stall key interception.
- Initialize COM on the controller/UI thread because virtual desktop APIs are COM-based.
- Run as a single resident process with:
  - one hidden top-level controller window for the message loop and tray icon
  - one pre-created overlay window for the switcher UI
- Enforce single-instance behavior with a named mutex so multiple processes do not compete for the same global hook.

### Input Interception

- Install a global `WH_KEYBOARD_LL` hook.
- Keep the hook callback extremely small:
  - track `Alt` pressed/released state
  - detect `Tab` presses while `Alt` is held
  - suppress the native `Alt+Tab` keystroke by returning a non-zero result when our switcher takes ownership of the combo
  - hand work off asynchronously to the controller thread with `PostMessage` or an equivalent non-blocking queue wake-up
- Do not enumerate windows, resolve icons, allocate heavily, or render UI inside the hook callback.
- Do not make synchronous cross-thread/UI calls from the hook callback.
- Model the interaction as a small state machine:
  - `idle`
  - `cycling`
  - `commit`
  - `cancel`
- The first `Tab` press while `Alt` is held starts a switching session.
- The initial selected item is index `1` in the MRU list.
- Each further `Tab` press advances the selected item modulo the candidate list length.
- Releasing `Alt` commits the current selection and ends the session.

### Window Inventory And MRU Tracking

- Keep a warm in-memory cache of switchable windows instead of discovering everything from scratch on every `Alt+Tab`.
- Seed the cache once at startup by enumerating top-level windows.
- Maintain per-window MRU order continuously from `EVENT_SYSTEM_FOREGROUND` notifications.
- Treat the foreground event stream as the source of truth for ordering, but not as the complete source of window inventory.
- On session start, build the candidate list as:
  - currently valid windows that already exist in MRU order
  - followed by other currently valid windows not yet in MRU, appended in a deterministic fallback order from a fresh inventory snapshot
- Remove stale handles lazily when a cached window no longer exists or no longer passes filtering.
- If the cache becomes inconsistent, repair it outside the hook path with a full rescan.

### Switchable Window Filtering

- Start from top-level desktop windows.
- Filter them to match native `Alt+Tab` behavior as closely as practical.
- The initial filtering rules should be:
  - visible window
  - not one of this app's own windows
  - not a tool window (`WS_EX_TOOLWINDOW`)
  - not an inert helper/shell surface
  - not cloaked for the current desktop
  - has meaningful top-level presence rather than being a tiny helper or hidden owner window
- The intended eligibility algorithm should explicitly account for:
  - root-owner and last-active-popup relationships
  - owned popup suppression
  - shell/UWP host exclusions such as `ApplicationFrameHost`-style surfaces when they are not the real switch target
- Follow native behavior for multi-window apps by keeping each switchable top-level window as its own MRU item.
- Keep minimized windows in the candidate list if they are otherwise switchable.
- Resolve owner/root-owner relationships carefully so transient helper windows are not surfaced as normal switch targets.
- Keep the filtering code separate from ordering logic so the heuristics can be refined independently.

### Virtual Desktop Handling

- Filter to the current virtual desktop only.
- Use the official `IVirtualDesktopManager::IsWindowOnCurrentVirtualDesktop` API as the main desktop-membership check.
- Keep a `DWMWA_CLOAKED` check as a secondary signal because Windows may cloak windows that should not be presented.
- Do not mirror the system `VirtualDesktopAltTabFilter` setting in v1 even though the example project supports that behavior.

### UI Rendering

- Create the overlay window at startup and keep it hidden until a switch session begins.
- The overlay window should:
  - be topmost
  - not appear in `Alt+Tab`
  - not steal focus from the currently active app while visible
  - be explicitly non-activating (`WS_EX_NOACTIVATE`)
  - support transparent composited drawing
- Use a layered window and paint the final frame into an offscreen bitmap, then present it with `UpdateLayeredWindow`.
- Keep the UI intentionally simple:
  - icon strip only
  - selected item highlight
  - no thumbnails
  - no animation requirement
- Do not call `SetFocus`, `SetActiveWindow`, or `SetForegroundWindow` on the overlay.
- Cache icons ahead of time and reuse them between sessions.
- Resolve icons in this order:
  - window-provided icon (`WM_GETICON` / class icon)
  - executable or shell icon fallback
  - generic fallback icon
- Place the overlay on the monitor that contains the currently focused window, centered on that monitor.
- While the user is holding `Alt`, repaint only when the selected index changes or the session is cancelled/committed.
- Keep the overlay keyboard-only; do not add mouse hover or click interaction in v1.

### Switching And Activation

- When `Alt` is released, validate the currently selected window before activating it.
- If the selected window disappeared or became invalid, walk forward in MRU order until the next valid candidate is found.
- If no valid candidate remains, cancel the switch cleanly and hide the UI.
- If the chosen window is minimized, restore it before activation.
- Use the same foreground-activation pattern proven in the OSS example:
  - restore minimized windows with `ShowWindow(..., SW_RESTORE)` when needed
  - send a minimal synthetic input event
  - call `SetForegroundWindow`
- Verify that the chosen window actually became foreground after activation and handle failure explicitly rather than assuming success.
- Once activation succeeds, update the in-memory MRU ordering immediately so rapid repeated `Alt+Tab` presses toggle between the most recent two windows.

### Tray Icon And Lifecycle

- Register a tray icon with `Shell_NotifyIconW`.
- The tray menu only needs one action in v1: `Close`.
- Re-register the tray icon if Explorer/taskbar restarts (`TaskbarCreated`).
- On exit:
  - unhook the keyboard hook
  - unregister WinEvent hooks
  - remove the tray icon
  - destroy the overlay/controller windows
  - release icon and GDI resources

### Performance Rules

- No expensive work inside the low-level keyboard hook.
- Pre-create windows, graphics resources, and caches at startup so the first `Alt+Tab` after launch is already warm.
- Reuse slices, maps, icon handles, and drawing buffers rather than rebuilding them every time.
- Avoid disk I/O, process inspection, and icon extraction in the session hot path unless a cache miss forces it.
- Prefer message passing between components over locks on the hot path.
- Keep the overlay hidden rather than destroying and recreating it between sessions.

### Early Validation Points

- Verify that suppressing native `Alt+Tab` with `WH_KEYBOARD_LL` is reliable across normal desktop usage.
- Verify that foreground activation remains reliable for minimized windows and for rapid repeated `Alt+Tab` presses.
- Verify behavior with elevated apps; Windows integrity boundaries may require special handling or a documented limitation.
- Verify that the overlay window never appears in the switch list and never steals focus while `Alt` is still held.

## Detailed Implementation Spec

### Suggested Code Structure

The initial codebase should be organized around one executable and a small set of focused internal packages:

- `cmd/better-alt-tab/main.go`
  - process entrypoint
  - single-instance guard
  - startup logging
  - application bootstrap and shutdown
- `internal/app/app.go`
  - owns the full application lifecycle
  - wires together hooks, caches, controller window, overlay, tray, and shutdown
  - contains the top-level event loop and dispatch
- `internal/win32/`
  - low-level Win32 wrappers, constants, structs, and helper functions
  - custom message constants used by the controller window
  - keep unsafe/syscall details here rather than leaking them into higher-level packages
- `internal/input/hook.go`
  - installs and removes `WH_KEYBOARD_LL`
  - runs on its own locked OS thread
  - translates raw key events into app-level events such as `tab_pressed`, `alt_released`, `cancel`
- `internal/events/foreground.go`
  - installs WinEvent hooks for foreground tracking
  - pushes foreground changes into the MRU tracker
- `internal/windows/inventory.go`
  - enumerates top-level windows
  - validates whether a window is still alive and switchable
  - produces fresh inventory snapshots on demand
- `internal/windows/filter.go`
  - implements the switchable-window eligibility algorithm
  - keeps all `Alt+Tab`-parity heuristics in one place
- `internal/windows/activate.go`
  - restore + activate logic
  - activation verification
  - failure fallback to next valid candidate
- `internal/windows/icons.go`
  - extracts and caches icons for windows/processes
  - separates icon lookup from inventory logic
- `internal/windows/desktop.go`
  - virtual desktop checks
  - COM initialization assumptions
  - `IsWindowOnCurrentVirtualDesktop` wrapper
- `internal/mru/store.go`
  - per-window MRU ordering
  - stale-handle removal
  - candidate list construction from MRU + fresh inventory
- `internal/session/session.go`
  - switching session state machine
  - selected index management
  - commit/cancel behavior
- `internal/ui/controller_window.go`
  - hidden top-level controller window
  - receives hook/tray/system messages
- `internal/ui/overlay.go`
  - non-activating topmost layered overlay window
  - draw/update/hide behavior
- `internal/ui/layout.go`
  - icon strip sizing and monitor placement
- `internal/ui/tray.go`
  - tray icon registration
  - context menu with `Close`
  - `TaskbarCreated` re-registration handling
  - `NIM_SETVERSION` setup for modern shell behavior
- `internal/runtime/single_instance.go`
  - named mutex ownership
- `internal/runtime/threading.go`
  - helpers for locked-thread startup and cleanup where useful

### Preferred Dependency Approach

- Use native Win32 from Go directly rather than a GUI toolkit.
- Prefer `golang.org/x/sys/windows` plus local syscall wrappers in `internal/win32`.
- Avoid bringing in a heavy UI framework; the overlay is simple enough to own directly.
- Keep the hook path free of reflection-heavy or allocation-heavy abstractions.

### Runtime Model

At runtime the program should have three main execution domains:

- Hook thread
  - locked OS thread
  - owns `WH_KEYBOARD_LL`
  - performs only minimal key-state tracking and asynchronous dispatch
- Controller/UI thread
  - locked OS thread
  - owns COM initialization
  - owns the controller window, overlay window, tray icon, and message pump
  - serializes all session state transitions
- Background worker paths
  - optional goroutines for slow cache fill work such as icon extraction or periodic repair scans
  - must never block hook handling or activation

The controller/UI thread is the authoritative owner of session state, MRU state mutation, and visible UI state.

### Core Data Model

The implementation should revolve around a few explicit structs:

- `WindowID`
  - wraps `HWND`
  - used as the stable identity key throughout the app
- `WindowInfo`
  - `HWND`
  - title
  - process ID
  - executable path
  - root owner / owner
  - style / ex-style flags needed for filtering
  - minimized / cloaked / visible state
  - current-desktop membership
  - cached icon handle reference
- `InventorySnapshot`
  - list or map of all currently valid switchable windows discovered in a fresh scan
  - contains a deterministic fallback order derived from the fresh enumeration order
- `MRUStore`
  - ordered list of `WindowID`
  - fast membership lookup
  - helpers to move a window to front, remove stale entries, and build a candidate list
- `SwitchSession`
  - active flag
  - candidate list
  - selected index
  - started-from foreground window
  - whether the overlay is currently shown
- `AppState`
  - owns `MRUStore`
  - icon cache
  - current `SwitchSession`
  - references to hooks, controller window, overlay, and tray

### Controller Messages

The controller window should receive a small fixed set of custom messages. The exact numeric values can be assigned later, but the message categories should be:

- `msgHookTabPressed`
- `msgHookAltReleased`
- `msgHookCancel`
- `msgForegroundChanged`
- `msgRescanInventory`
- `msgTrayOpen`
- `msgTrayExit`
- `msgTaskbarCreated`

The hook thread should only post these messages; it should not directly call session logic.

### Startup Sequence

The startup sequence should be:

1. Acquire single-instance mutex.
2. Start logging if enabled.
3. Start controller/UI thread.
4. Initialize COM on the controller/UI thread.
5. Register controller window class and create controller window.
6. Create overlay window up front and keep it hidden.
7. Register tray icon.
8. Perform initial inventory scan.
9. Seed MRU order from the current foreground window plus any fresh inventory fallback order.
10. Install foreground WinEvent hook.
11. Start hook thread and install `WH_KEYBOARD_LL`.
12. Enter steady-state message loop.

If any required startup step fails after partially initializing the process, unwind in reverse order and exit cleanly.

### Keyboard Hook Behavior

The hook should track only the minimal state needed to model `Alt+Tab`:

- whether left or right `Alt` is currently pressed
- whether `Tab` has been seen while `Alt` is held during the current session
- whether `Tab` is currently down, so held-key auto-repeat does not advance selection repeatedly
- whether the current key sequence belongs to this app and must suppress native behavior

Detailed behavior:

- `Alt` down alone does nothing.
- First `Tab` down while `Alt` is held:
  - suppress native `Alt+Tab`
  - post `msgHookTabPressed`
  - mark the session as owned by this app
- Additional `Tab` down while `Alt` remains held:
  - suppress native `Alt+Tab`
  - post `msgHookTabPressed`
- Only a fresh `Tab` down after a `Tab` up should advance selection; held-key auto-repeat should be ignored.
- `Alt` up after at least one handled `Tab`:
  - post `msgHookAltReleased`
- `Esc` while a session is active:
  - suppress the key
  - post `msgHookCancel`
- Keys unrelated to the active switching sequence should normally pass through.

The hook must not call into rendering, window enumeration, icon resolution, or activation code.

### Session State Machine

The controller thread should manage a simple explicit state machine:

- `Idle`
  - no visible overlay
  - no active candidate list
- `Cycling`
  - overlay visible
  - candidate list frozen for the current session
  - selected index advances on each `Tab`
- `CommitPending`
  - `Alt` released
  - overlay hidden
  - selected target being validated and activated
- `Cancelled`
  - transient cleanup state before returning to `Idle`

State transitions:

- `Idle -> Cycling`
  - on first handled `Tab`
  - build session candidate list
  - if fewer than two valid candidates exist, stay in `Idle`
- `Cycling -> Cycling`
  - on additional handled `Tab`
  - increment selected index modulo list length
- `Cycling -> CommitPending`
  - on `Alt` release
- `Cycling -> Cancelled`
  - on explicit cancel or unrecoverable session inconsistency
- `CommitPending -> Idle`
  - after activation success or clean failure
- `Cancelled -> Idle`
  - after overlay hide and session reset

### Candidate List Construction

When a session starts, build the candidate list in this exact order:

1. Take the current `MRUStore`.
2. Remove any entries whose windows are no longer valid or no longer switchable.
3. Take a fresh `InventorySnapshot`.
4. Append all valid MRU windows that are present in the fresh snapshot, preserving MRU order.
5. Append all remaining valid windows from the fresh snapshot that are not already present, preserving the snapshot's deterministic fallback order based on consistent native top-level enumeration order.

This prevents stale-cache omission and ensures newly opened but never-focused windows can still appear.

Selection rules:

- Candidate `0` is the currently active/most-recent window.
- Initial selected index is `1`.
- If candidate count is `< 2`, no switching session is shown.

The candidate list should remain stable for the duration of a single `Alt` hold. Do not reorder it mid-session from external foreground noise.

### Switchable-Window Eligibility Algorithm

The window filter should be implemented as an ordered predicate chain. A window is eligible only if all of the following are true:

1. The handle is still valid.
2. The window is a top-level desktop window candidate from enumeration.
3. The window is not one of this app's own windows.
4. The window is visible in the sense required for `Alt+Tab` participation.
5. The window is not a tool window.
6. The window is on the current virtual desktop.
7. The window is not cloaked in a way that disqualifies it from current-desktop switching.
8. The window is not a known shell/helper surface that should never be surfaced directly.
9. The window has an appropriate owner/root-owner/last-active-popup relationship for `Alt+Tab` visibility.
10. The final chosen representative for that owner chain is the visible switch target, not a hidden helper.
11. The window remains activatable.

Implementation notes:

- Minimized windows are still eligible if they otherwise pass.
- Owned popup suppression must avoid surfacing transient dialogs or invisible owners as primary switch targets.
- The algorithm should be written so the special-case exclusions are centralized and auditable.
- The first implementation may still be heuristic-based, but the heuristics must be explicit, ordered, and easy to adjust.

### Inventory Refresh Policy

Inventory freshness should come from several sources rather than only foreground changes:

- initial full scan at startup
- the fresh snapshot order should come from one explicit enumeration strategy and stay consistent; v1 should use the native top-level enumeration order returned by the chosen Win32 enumeration path
- fresh snapshot at session start
- foreground events to keep MRU order warm
- lazy removal when a cached handle fails validation
- explicit repair rescan when activation or validation detects inconsistency
- optional low-frequency background repair if needed later

The default v1 behavior should avoid constant full rescans while idle, but it should always be allowed to take one fresh snapshot at session start.

### MRU Update Rules

Update MRU ordering in the following situations:

- on `EVENT_SYSTEM_FOREGROUND`, move the new foreground window to the front if it is switchable
- after a successful activation initiated by this app, move the activated window to the front immediately
- when a window becomes invalid, remove it lazily

Do not reorder MRU from non-foreground windows becoming visible in the background.

### Icon Cache Policy

The icon cache should be keyed primarily by executable path, with optional per-window override when needed.

Resolution order:

1. cached icon
2. `WM_GETICON`
3. class icon
4. executable or shell icon
5. generic fallback icon

Rules:

- do not block the first visible frame on expensive icon extraction when a cached or fallback icon can be shown immediately
- allow asynchronous refinement later if a better icon is discovered
- keep icon handles alive for the process lifetime unless resource pressure forces a cleanup policy later

### Overlay Window Behavior

The overlay window should be:

- `WS_EX_TOPMOST`
- `WS_EX_TOOLWINDOW`
- `WS_EX_LAYERED`
- `WS_EX_NOACTIVATE`

Behavior rules:

- create once, hide/show as needed
- never call focus APIs on it
- keep it out of the switchable window set
- position it based on the monitor containing the current foreground window
- render the full frame offscreen, then push one composed update
- show and position it through a no-activate path such as `SetWindowPos(..., SWP_NOACTIVATE | SWP_SHOWWINDOW)` or equivalent
- only repaint when the session starts, selection changes, or the session ends

The first implementation can use a simple centered horizontal icon strip with a selected-item background.

### Activation Flow

Commit flow on `Alt` release:

1. Hide the overlay immediately.
2. Read the selected candidate.
3. Validate it again.
4. If invalid, walk forward through the frozen candidate list until a valid target is found.
5. If none is found, reset to `Idle` with no activation.
6. If minimized, restore the target.
7. Send the minimal synthetic input event needed to satisfy foreground rules.
8. Call `SetForegroundWindow`.
9. Verify that the target actually became foreground.
10. If successful, update MRU and end session.
11. If activation failed because the target became invalid during commit, try the next valid candidate once.
12. If activation failed for a still-valid target, retry that same target at most once, then fail cleanly and return to `Idle`.

For v1, keep fallback simple: advance only when the original target became invalid, and otherwise retry only that same target once. Do not build a complex activation fallback ladder yet.

### Failure And Limitation Policy

The implementation should explicitly tolerate the following without crashing or hanging:

- a window disappears mid-session
- the candidate list becomes partially stale
- Explorer restarts and tray state is lost
- icon extraction fails
- a target cannot be activated

Known likely limitation for v1:

- switching to elevated/admin windows may fail when the switcher itself is not elevated

This limitation should be documented rather than papered over with brittle workarounds.

### Shutdown Sequence

Shutdown should happen in reverse dependency order:

1. mark app as shutting down so new sessions are ignored
2. hide overlay
3. remove keyboard hook
4. remove WinEvent hooks
5. remove tray icon
6. destroy overlay and controller windows
7. release icon/GDI resources
8. uninitialize COM on the controller thread
9. release single-instance mutex

### Implementation Priorities

Build the feature in this order:

1. application bootstrap, controller window, and clean shutdown
2. keyboard hook with native `Alt+Tab` suppression
3. top-level window enumeration and eligibility filtering
4. MRU tracking from foreground events
5. session state machine with no UI
6. activation path
7. overlay rendering with placeholder icons
8. icon cache and tray icon
9. edge-case hardening and latency cleanup
