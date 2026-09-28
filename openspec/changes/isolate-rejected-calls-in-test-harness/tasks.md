# Tasks

## 1. Build the primitive

- [ ] 1.1 Add `expect_rejected_call(obj, method, args)` to `test/demo/test_base.gd` alongside the existing `assert_*` primitives, setting `last_call_was_rejected = true` *before* the `callv` and clearing it after, so the value that survives frame abandonment is the correct one. Verify by reading the function and confirming no path reports rejection via a return value.
- [ ] 1.2 Print a marker line naming the target method before making the call, so a run that dies here points at the primitive in the log instead of just stopping. Verify the marker appears in a run's output for an induced rejection.
- [ ] 1.3 Guard against a mistyped method name masquerading as a rejection: check the target actually has the method before calling, and fail the test with a message identifying the typo if it does not. Verify by passing a nonexistent method name and confirming the failure says "no such method" rather than reading as a successful rejection.

## 2. Prove the containment assumption

- [ ] 2.1 Verify by experiment that a rejection inside the primitive leaves the calling function executing: place an assertion immediately after the primitive call and confirm it is evaluated and counted. Verify the pass count increases by the expected amount rather than the assertion vanishing.
- [ ] 2.2 Verify the inverse regression is caught: point the primitive at a call that is *not* rejected and confirm the test fails via `last_call_was_rejected` being cleared. Verify it fails rather than passing quietly.
- [ ] 2.3 Confirm the whole-stack-unwind failure mode is loud: temporarily reason about or simulate a non-local unwind and confirm the driver's missing summary banner trips the existing `no-driver-summary` gate with the marker line identifying the primitive. Record what was checked in the change notes -- this one is about knowing how it would fail, not forcing it.

## 3. Retire the workaround

- [ ] 3.1 Reshape `test_varcall_arg_rejection` in `test/demo/main.gd` to use the primitive, moving the `free()` calls and the surviving probe-count assertion back next to the rejection instead of split across two functions. Verify cleanup runs and no engine objects leak.
- [ ] 3.2 Remove the now-redundant `test_varcall_rejection_survivable` group if the reshaped function covers it, or keep it and say why. Verify no assertion coverage was lost in the move by comparing the pass count before and after.
- [ ] 3.3 Add a short comment at the reshaped rejection site recording why the primitive is required -- the inline shape silently skips everything after the rejection. Verify a reader who has never seen this change understands the constraint from the comment alone.

## 4. Verify

- [ ] 4.1 Run `make test` and confirm the suite is green: extension loaded, positive assertion count, zero failures, no leaked engine objects.
- [ ] 4.2 Confirm the pass count did not drop from the reshaping -- a lower count means assertions were stranded or deleted rather than moved. Record the before and after counts in the change notes.
