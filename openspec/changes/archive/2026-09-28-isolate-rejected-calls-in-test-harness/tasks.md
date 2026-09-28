# Tasks

**Design correction during implementation:** the primitive does not *contain* a frame
abandonment -- `callv` never strands the caller in the first place. Rejection is
read from the return value. See design.md, "Detect rejection from the return
value". Task 1.1's original wording was built on a wrong premise and is recorded
as such below.

## 1. Build the primitive

- [x] 1.1 Add `expect_rejected_call(obj, method, args)` to `test/demo/test_base.gd` alongside the existing `assert_*` primitives, setting `last_call_was_rejected = true` *before* the `callv` and clearing it after, so the value that survives frame abandonment is the correct one. Verify by reading the function and confirming no path reports rejection via a return value.
  - Implemented, but **the stated mechanism was wrong and the first run proved it.** The pre-set flag reported `false` on a call that *was* rejected, because `callv` returns normally and the flag was cleared on the next line. Measured, the two dynamic call forms differ:
    - `obj.call("m", bad)` -- `SCRIPT ERROR`, frame abandoned, everything after it dead.
    - `obj.callv("m", [bad])` -- engine `ERROR` logged, returns `<null>`, caller continues.
  - The primitive now derives the flag from that return: `last_call_was_rejected = obj.callv(method, args) == null`. No before-the-fact state needed, and `callv` removes the stranding problem rather than containing it.
  - Constraint this introduces, documented at the definition: the target must return a value. A void method yields the same null as a rejection, so void targets cannot be judged this way.
- [x] 1.2 Print a marker line naming the target method before making the call, so a run that dies here points at the primitive in the log instead of just stopping. Verify the marker appears in a run's output for an induced rejection.
  - Verified: `expect_rejected_call: test_varcall_reject_probe (rejection expected)` appears immediately before the engine's rejection error.
- [x] 1.3 Guard against a mistyped method name masquerading as a rejection: check the target actually has the method before calling, and fail the test with a message identifying the typo if it does not. Verify by passing a nonexistent method name and confirming the failure says "no such method" rather than reading as a successful rejection.
  - Verified by experiment with `test_varcall_rejst_typo`:
    `expect_rejected_call: target has no method 'test_varcall_rejst_typo' -- that is a typo, not a rejection`.
    Without the guard this would have read as a rejection that worked.

## 2. Prove the containment assumption

- [x] 2.1 Verify by experiment that a rejection inside the primitive leaves the calling function executing: place an assertion immediately after the primitive call and confirm it is evaluated and counted. Verify the pass count increases by the expected amount rather than the assertion vanishing.
  - Confirmed. The assertion two lines after the rejection executed and was counted -- that is how the wrong premise surfaced at all. Total went 1103 → 1104 with the survivable group removed and two assertions added, exactly as predicted.
- [x] 2.2 Verify the inverse regression is caught: point the primitive at a call that is *not* rejected and confirm the test fails via `last_call_was_rejected` being cleared. Verify it fails rather than passing quietly.
  - Verified by pointing the primitive at the accepted `sprite` call instead of the rejected `label` call. Two failures fired:
    - `assert_true(last_call_was_rejected)` -- correctly reports not-rejected.
    - `assert_equal(probe_count, before)` -- the body ran, so the count moved.
  - Both detectors work. Injection reverted.
- [x] 2.3 Confirm the whole-stack-unwind failure mode is loud: temporarily reason about or simulate a non-local unwind and confirm the driver's missing summary banner trips the existing `no-driver-summary` gate with the marker line identifying the primitive. Record what was checked in the change notes -- this one is about knowing how it would fail, not forcing it.
  - **This task's premise dissolved.** There is no containment to fail: `callv` does not unwind anything, so the scenario the task guarded against cannot occur.
  - What was checked instead, and what the spec now requires: the *return-value inference* is the load-bearing assumption, and both sides of it are pinned by 2.1 (rejected → reports true) and 2.2 (accepted → reports false). If a future Godot changed `callv`'s return behaviour, one of those two flips and the suite fails. The spec requirement was rewritten from "containment shall be pinned" to "the rejection signal's basis shall be pinned".

## 3. Retire the workaround

- [x] 3.1 Reshape `test_varcall_arg_rejection` in `test/demo/main.gd` to use the primitive, moving the `free()` calls and the surviving probe-count assertion back next to the rejection instead of split across two functions. Verify cleanup runs and no engine objects leak.
  - Reshaped. `label.free()` and `sprite.free()` now follow the rejection and run; the run reports no leaked engine objects. The `doomed_label` / `add_child` trick that existed only to dodge the stranding is gone.
- [x] 3.2 Remove the now-redundant `test_varcall_rejection_survivable` group if the reshaped function covers it, or keep it and say why. Verify no assertion coverage was lost in the move by comparing the pass count before and after.
  - Removed, along with its `_ready` call and the `probe_count_after_control` script-level var (now a local). The reshaped function covers it: assertions after the rejection execute (survival) and the probe count is unchanged (body never ran).
  - Pass count 1103 → 1104: −1 for the removed group, +2 for the rejection report and probe-count assertions. Net gain, no coverage lost.
- [x] 3.3 Add a short comment at the reshaped rejection site recording why the primitive is required -- the inline shape silently skips everything after the rejection. Verify a reader who has never seen this change understands the constraint from the comment alone.
  - Comment at the call site names the failure mode explicitly: written inline with `call()`, the engine's rejection makes GDScript abandon the function and everything below -- including the frees -- silently stops while the suite still reports green.

## 4. Verify

- [x] 4.1 Run `make test` and confirm the suite is green: extension loaded, positive assertion count, zero failures, no leaked engine objects.
  - `PASS: extension loaded, 1104 assertions ran, 0 failures, no leaked engine objects`.
- [x] 4.2 Confirm the pass count did not drop from the reshaping -- a lower count means assertions were stranded or deleted rather than moved. Record the before and after counts in the change notes.
  - Before: 1103. After: 1104. Up by one, matching the arithmetic above.
