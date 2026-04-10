# Performance Analysis

## Goal

This document captures the current investigation into the noticeable delay between releasing `Alt` and the actual window switch. It prioritizes the likely causes, explains why they matter, and outlines the most practical ways to make switching feel instant.

The repo spec is explicit:

- UI display on `Alt+Tab` should be instant.
- Activation on `Alt` release should be instant.
- `Alt` release activation is hot-path code.

The current implementation does not fully meet that requirement.

## Summary

The main delay does not appear to come from an inherent limitation of the low-level keyboard hook approach.

The strongest likely cause is current app behavior:

- the hook suppresses `Tab`
- but does not suppress the final owned `Alt` key-up
- then the app immediately tries to activate another window with `SetForegroundWindow`

That sequencing likely conflicts with Windows foreground activation rules and can introduce the visible delay after key release.

There are also several secondary inefficiencies in the app that make the hot path heavier and less deterministic than it should be, especially full inventory rescans during commit and routine foreground tracking.

## Findings

### 1. Highest Priority: owned `Alt` release is not swallowed

Relevant code:

- [internal/input/hook.go](/C:/Users/matej/dev/quick_app_switcher/internal/input/hook.go)
- [internal/app/app.go](/C:/Users/matej/dev/quick_app_switcher/internal/app/app.go)
- [internal/windows/activate.go](/C:/Users/matej/dev/quick_app_switcher/internal/windows/activate.go)

What happens now:

- `Tab` is swallowed once the app takes ownership of the `Alt+Tab` sequence.
- On `Alt` key-up, the hook posts the app-level commit message.
- But the hook then falls through to `CallNextHook`, so the foreground app still receives the `Alt` key-up.

Why this matters:

- A naked `Alt` release can put the current foreground app into Alt/menu mode.
- The app then immediately calls `SetForegroundWindow` on the target window.
- Microsoft documents that `SetForegroundWindow` has restrictions, including that no menus are active.
- This makes the current sequence hostile to instant activation.

Why this is likely the main cause:

- The delay is observed specifically after letting go of `Alt`.
- The current code changes system/input state at that exact moment in a way that can interfere with foreground activation.
- This lines up better with the observed symptom than the measured cost of the window snapshot itself.

Recommended fix:

- When the switcher owns the session, swallow the final `Alt` key-up.
- Post the app-level commit message, but do not let that `Alt` key-up reach the previously focused app.

Expected effect:

- Avoids putting the old app into menu/Alt mode.
- Makes `SetForegroundWindow` more likely to succeed immediately.
- Should be the single highest-value latency fix.

## 2. High Priority: full window inventory rebuild on `Alt` release

Relevant code:

- [internal/app/app.go](/C:/Users/matej/dev/quick_app_switcher/internal/app/app.go)
- [internal/windows/inventory.go](/C:/Users/matej/dev/quick_app_switcher/internal/windows/inventory.go)
- [internal/win32/ui.go](/C:/Users/matej/dev/quick_app_switcher/internal/win32/ui.go)

What happens now:

- `onAltReleased` calls `commitSelection`.
- `commitSelection` calls `inventory.IsValidSwitchable`.
- `IsValidSwitchable` does not perform a cheap targeted validation.
- Instead, it rebuilds a full fresh snapshot by enumerating all top-level windows.

What that snapshot currently does per window:

- checks validity
- reads title
- reads PID
- reads executable path
- reads style/ex-style
- resolves owner/root-owner/popup relationships
- checks cloaking
- checks current virtual desktop membership
- checks class name

Why this matters:

- This is explicitly hot-path work.
- Even if the full snapshot is not the main source of the 200 ms delay on this machine, it is still avoidable and adds jitter.
- It also makes latency sensitive to the user’s current desktop/window count.

Measured data:

- A one-iteration local benchmark of `Snapshot()` was about `1.23 ms/op` on this machine.

Interpretation:

- This measurement suggests the snapshot is not the primary explanation for a stable, clearly visible `~200 ms` delay here.
- It is still a design problem because it makes the release path heavier than necessary and can combine badly with other delays.

Recommended fix:

- Do not rebuild global inventory on commit.
- Use the frozen session candidate list and the snapshot captured at session start.
- Revalidate only the chosen target with cheap direct checks.

Cheap commit-time validation should prefer:

- `IsWindow`
- maybe `IsWindowVisible` if needed
- maybe style/ex-style checks if needed
- maybe minimized state
- maybe desktop/cloak check only if there is a known risk of drift

Expected effect:

- Keeps `Alt` release commit deterministic.
- Removes avoidable global work from the critical path.

## 3. High Priority: full inventory rebuild on every foreground event

Relevant code:

- [internal/app/app.go](/C:/Users/matej/dev/quick_app_switcher/internal/app/app.go)
- [internal/events/foreground.go](/C:/Users/matej/dev/quick_app_switcher/internal/events/foreground.go)
- [internal/windows/inventory.go](/C:/Users/matej/dev/quick_app_switcher/internal/windows/inventory.go)

What happens now:

- The foreground watcher posts the new foreground HWND.
- `onForegroundChanged` calls `inventory.IsValidSwitchable`.
- `IsValidSwitchable` again does a full snapshot.

Why this matters:

- Foreground changes are common.
- This means the controller thread is repeatedly doing heavier work than needed even while idle.
- That can create queue backlog or timing variability right before `Alt` release arrives.

Why it is not ideal:

- The foreground event already gives the candidate HWND.
- Most of the time, this only needs a cheap targeted eligibility check or a trust-but-filter policy.
- Full global recomputation here is unnecessary for MRU maintenance.

Recommended fix:

- Replace full snapshot validation in foreground handling with a cheap targeted check.
- Keep occasional repair rescans outside the hot path.
- Treat foreground events primarily as MRU updates, not as a trigger for whole-desktop recomputation.

Expected effect:

- Lower steady-state controller load.
- Less chance that a release message lands behind avoidable work.

## 4. Medium Priority: hot-path allocations and avoidable data churn

Relevant code:

- [internal/app/app.go](/C:/Users/matej/dev/quick_app_switcher/internal/app/app.go)
- [internal/mru/store.go](/C:/Users/matej/dev/quick_app_switcher/internal/mru/store.go)
- [internal/ui/overlay.go](/C:/Users/matej/dev/quick_app_switcher/internal/ui/overlay.go)

Current issues:

- `renderOverlay` allocates a fresh `items` slice each render.
- `MoveToFront` prepends using `append([]windows.WindowID{id}, s.order...)`, which allocates and copies.
- The session and overlay paths are otherwise simple, but there is still avoidable churn.

Why this matters:

- These are not the likely source of the large visible delay.
- They still make the app less crisp under rapid repeated use.
- Reducing them improves consistency once the main blocker is fixed.

Recommended fix:

- Reuse overlay item buffers.
- Use more efficient MRU front-move logic.
- Keep repeated selection changes allocation-light.

Expected effect:

- Smaller but worthwhile improvement in responsiveness consistency.

## 5. Medium Priority: `EnumWindows` callback allocation pattern is not reusable

Relevant code:

- [internal/win32/core.go](/C:/Users/matej/dev/quick_app_switcher/internal/win32/core.go)

What happens now:

- `EnumWindows` creates a new `syscall.NewCallback` every time it runs.

Observed issue:

- Repeated synthetic benchmarking eventually hit Go’s callback limit with `fatal error: too many callback functions`.

Why this matters:

- This is not the visible interactive switching delay by itself.
- It is still an implementation smell.
- It confirms that the current enumeration mechanism is not suitable for being called frequently in latency-sensitive paths.

Recommended fix:

- Rework `EnumWindows` to use a reusable callback bridge instead of creating a new callback on every call.

Expected effect:

- Improves robustness.
- Removes one more reason to avoid frequent full scans.

## 6. Lower Priority: current activation fallback is simple but still synchronous

Relevant code:

- [internal/windows/activate.go](/C:/Users/matej/dev/quick_app_switcher/internal/windows/activate.go)
- [internal/app/app.go](/C:/Users/matej/dev/quick_app_switcher/internal/app/app.go)

What happens now:

- activation may retry the selected target twice
- replacement target logic is still synchronous
- verification is immediate

Assessment:

- The activation helper itself is small and reasonable.
- It is not obviously bloated.
- The bigger issue is the input/menu-state sequencing before activation and the rescan-driven validation around it.

Recommended fix:

- Keep activation logic simple.
- Fix the sequencing and validation path first.
- Only add more complex fallback logic if real failures remain after the main fixes.

## Is this a limitation of the low-level hook approach?

Current assessment: no, not primarily.

Reasons:

- The hook callback is already kept fairly small.
- It runs on its own dedicated thread.
- It posts async messages to the controller thread.
- This matches the intended architecture from `spec.md`.

The larger problem is not using `WH_KEYBOARD_LL`.

The larger problem is what the app does around the owned `Alt` release and what it asks the controller thread to do synchronously during commit.

## Best path to make switching feel instant

The fastest practical path is:

1. Swallow the final owned `Alt` key-up.
2. Remove full inventory rescans from `Alt` release commit.
3. Remove full inventory rescans from routine foreground tracking.
4. Keep candidate construction at session start, then freeze that list for the hold duration.
5. Revalidate targets cheaply at commit.
6. Clean up smaller allocations and reusable callback issues.

## Proposed implementation approach

### Phase 1: fix the biggest perceived delay

- Update the keyboard hook so an owned session consumes the final `Alt` key-up.
- Keep posting the commit message, but do not forward that `Alt` release to the old foreground app.

Success criteria:

- Releasing `Alt` no longer leaves the old app in menu/Alt state.
- Switch commit feels materially more immediate.

### Phase 2: thin the release path

- Replace `inventory.IsValidSwitchable(selected)` in commit with a targeted validation helper.
- Use session-frozen candidates and session-start snapshot data as the source of truth for the current hold.
- Avoid any whole-desktop enumeration in `onAltReleased` and candidate fallback walking.

Success criteria:

- `Alt` release path does no global enumeration.
- Commit work is bounded and predictable.

### Phase 3: thin steady-state controller work

- Replace full-snapshot validation in `onForegroundChanged` with a cheaper targeted eligibility check.
- Keep full rescans for startup, session start, and explicit repair only.

Success criteria:

- Idle foreground churn no longer drives repeated full snapshot rebuilds.
- Controller thread stays more available for hook-driven events.

### Phase 4: clean up smaller overheads

- Reuse slices/buffers in overlay rendering.
- Improve MRU move-to-front mechanics.
- Rework `EnumWindows` callback allocation strategy.

Success criteria:

- Less allocation churn during rapid repeated switching.
- Better long-run robustness.

## Suggested target design for commit-time validation

Commit-time validation should be narrow and explicit.

Suggested checks:

- target HWND is non-zero
- `IsWindow(target)`
- if minimized, restore it
- if known-disqualifying style/visibility conditions changed, reject it
- if desktop membership drift matters, do one direct desktop check for that HWND only
- try activation
- verify foreground changed

What commit-time validation should avoid:

- enumerating every top-level window
- rebuilding MRU inputs
- resolving process paths for unrelated windows
- reconstructing owner chains for the whole desktop

## Recommended fallback strategy

Keep fallback simple:

1. try selected target
2. retry selected target once if the failure looks transient
3. if target became invalid, walk forward through the frozen candidate list
4. stop after one replacement success/failure path

Do not build a large activation ladder unless real post-fix evidence shows it is needed.

## Risks and notes

- Some activation failures may still exist for elevated/admin windows when the switcher is not elevated. That is a likely Windows limitation and should be treated separately from the current delay issue.
- Manual verification on real Windows behavior remains important even after code cleanup, because foreground rules are timing-sensitive.

## Validation performed during this investigation

- Reviewed `spec.md` and the app hot-path implementation.
- Compared the current activation flow to the OSS reference included in the repo.
- Confirmed that the commit path currently performs avoidable global validation work.
- Measured a one-iteration local `Snapshot()` benchmark at roughly `1.23 ms/op`.
- Confirmed that the benchmark surface also exposed the current `EnumWindows` callback allocation issue.
- Re-ran `go test ./...` successfully after the investigation.

## Bottom line

The switch delay appears to be primarily self-inflicted, not an unavoidable limitation of the low-level input approach.

The top fix is to consume the final owned `Alt` release.

After that, the next most important work is removing full inventory rescans from both:

- `Alt` release commit
- routine foreground tracking

Those changes together are the most credible path toward making switching feel instant.

## Reference material

- Microsoft `SetForegroundWindow`: https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setforegroundwindow
- Microsoft `AllowSetForegroundWindow`: https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-allowsetforegroundwindow
- Microsoft `LockSetForegroundWindow`: https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-locksetforegroundwindow
- Microsoft `LowLevelKeyboardProc`: https://learn.microsoft.com/en-us/windows/win32/winmsg/lowlevelkeyboardproc
