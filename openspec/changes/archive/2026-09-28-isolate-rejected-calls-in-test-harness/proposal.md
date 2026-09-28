# Proposal

## Why

A deliberately-rejected varcall takes the rest of its calling function with it.
When the engine reports `INVALID_ARGUMENT`, GDScript prints a script error and
abandons the frame that made the call. Writing the rejection tests for
`fix-varcall-arg-error-reporting` hit this twice: the first version put the
rejection mid-function, so four assertions written after it never executed while
the run still reported green, and the second version stranded two `free()` calls
and leaked nodes into the harness's leak check.

Both failures were silent in the way that matters most -- the suite kept passing
while verifying less than it appeared to. A test file that cannot place a
rejection anywhere except dead-last has no safe way to express the thing it most
needs to test.

## What Changes

- Add a harness helper that makes a call expected to be rejected, containing the
  frame abandonment inside the helper so the caller keeps executing and its
  cleanup and follow-up assertions survive.
- The helper reports whether the call actually was rejected, so a rejection that
  silently stops happening -- a regression in the opposite direction -- fails
  the test instead of passing quietly.
- Document the rule the helper encodes: never place a deliberately-rejected call
  inline where anything follows it.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `test-harness-integrity`: adds the requirement that a deliberately-rejected
  call cannot strand the assertions or cleanup that follow it, and that a
  rejection which stops being a rejection is detected rather than passed over.

## Impact

- `test/demo/test_base.gd` -- one helper alongside the existing `assert_*`
  primitives.
- `test/demo/main.gd` -- the rejection group added in `fix-varcall-arg-error-reporting`
  can be reshaped to use it, removing the split-function workaround that is
  currently load-bearing but undocumented.
- No production code affected. This is test-harness surface only.
- Depends on the frame-abandonment behaviour being function-local, which is
  observed but must be pinned by a test rather than assumed -- if a future Godot
  version unwinds the whole stack, the helper must fail loudly rather than
  silently stop containing anything.
