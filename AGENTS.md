# AGENTS.md

Repo-wide guidance for autonomous coding agents.

## General

- Read `spec.md` before making architectural or behavioral changes.
- Keep changes scoped. Do not widen scope without a concrete reason.
- Prefer the smallest correct change over a broad refactor.
- Preserve the existing architecture unless there is a clear technical reason to improve it.

## Coding Style

- Follow KISS. Prefer simple, obvious code over clever code.
- Prefer small, focused functions with one clear responsibility.
- Keep separation of concerns sharp.
- Use idiomatic Go.
- Prefer concrete types internally. Use interfaces mainly at system boundaries where they materially help testing or substitution.
- Keep exported surface area small.
- Add comments only for non-obvious intent, invariants, or Win32 quirks.

## Structure

- Keep pure logic separate from Win32 integration.
- Keep syscall/unsafe details near the boundary layer.
- Keep state transitions explicit.
- Prefer one clear owner for mutable state.
- Avoid speculative abstractions, generic helpers, and premature extension points.

## Hot Path

Treat the following as hot-path code:

- keyboard hook handling
- session start
- first overlay display
- activation on `Alt` release

For hot-path code:

- avoid blocking calls
- avoid disk I/O
- avoid unnecessary allocations
- avoid synchronous cross-thread calls
- avoid expensive work unless it is clearly required

If a change risks making the first visible frame slower, reconsider it.

## Error Handling

- Fail cleanly.
- Return contextual errors from setup and infrastructure code.
- In user-facing hot paths, prefer graceful degradation over crashes.
- Do not silently swallow errors without a deliberate reason.
- Do not assume important Win32 operations succeeded; verify critical postconditions when needed.

## Testing

- Use Go’s standard `testing` package by default.
- Test pure logic aggressively.
- Prefer table-driven tests for logic-heavy behavior.
- Add tests when changing state transitions, ordering logic, filtering logic, or bug-prone edge cases.
- Mock or wrap Win32 boundaries only when it materially improves testing of real logic.
- Do not write brittle tests for OS behavior that cannot be made deterministic.
- If a change depends on real Windows behavior, note the need for manual verification explicitly.

## When Tests Are Not Necessary

Tests are usually unnecessary for:

- obvious wiring changes with no meaningful branching
- thin Win32 pass-through wrappers with no real logic
- comments, naming, or no-behavior refactors

## Workflow

Before coding:

- identify what is pure logic vs OS-bound logic
- identify whether the change touches hot-path behavior

While coding:

- keep concerns separated
- avoid unnecessary abstractions
- update tests when logic changes

Before finishing:

- verify the change still matches `spec.md`
- check affected edge cases
- call out any untested Windows-specific behavior

## If In Doubt

- choose the simpler design
- choose the more explicit design
- keep the hot path thinner
- document limitations instead of hiding them behind fragile behavior
