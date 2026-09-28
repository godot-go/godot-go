# Design

## Context

See proposal.md - Why.

Observed behaviour, from the `fix-varcall-arg-error-reporting` runs: a varcall
the engine rejects makes GDScript print a script error and abandon the frame that
made the call. The abandonment is function-local -- `_ready` carried on and the
driver still printed its summary -- but everything after the call in the same
function is dead. Two consequences were hit live: four assertions were skipped
while the suite reported green, and a later layout leaked two nodes whose `free()`
calls landed after the rejection.

The current workaround is structural and undocumented: the rejection is the last
statement in one function and the surviving assertion lives in the next one.
That works, but it reads as an odd split rather than a constraint, and the next
person to touch it will "tidy" it back into the broken shape.

## Goals / Non-Goals

**Goals:**

- One primitive that makes a call expected to be rejected, without stranding the
  caller.
- Detect the opposite regression: a rejection that quietly stops happening.
- Make the split-function workaround unnecessary, and remove it.

**Non-Goals:**

- Catching arbitrary script errors. This is for expected engine call rejections,
  not a general try/catch, which GDScript does not have.
- Changing the gate's failure modes. The existing gate stays as is.

## Decisions

### Set the flag before the call, not after

```gdscript
var last_call_was_rejected := false

func expect_rejected_call(obj: Object, method: String, args: Array) -> void:
    print("expect_rejected_call: %s" % method)
    last_call_was_rejected = true
    obj.callv(method, args)
    last_call_was_rejected = false   # reached only if the call completed
```

A return value cannot express "rejected", because a rejected call never reaches a
`return` -- the caller would receive null and have to interpret it, and null is
also what a method returning nothing would give. Setting the flag *before* the
call means the value that survives the abandonment is the correct one.

The caller then asserts normally:

```gdscript
expect_rejected_call(example, "test_varcall_reject_probe", [label])
assert_true(last_call_was_rejected)
assert_equal(example.test_varcall_reject_probe_count(), probe_count_after_control)
```

**Rejected: return a bool.** Cannot distinguish "rejected" from "returned
nothing", and invites callers to write `assert_true(expect_rejected_call(...))`
against a null.

**Rejected: return a tri-state enum.** Same reachability problem; the function
does not get to return anything when the call is rejected.

### Print a marker before the call

If the containment assumption ever breaks and the whole stack unwinds, the last
line in the log identifies the primitive and the method, rather than the run
just stopping. The existing gate already fails on a missing summary banner; the
marker makes that failure diagnosable instead of mysterious.

### Verify containment rather than trust it

The assumption that abandonment is function-local is observed, not documented by
Godot. The self-check is that the caller's `assert_true(last_call_was_rejected)`
and everything after it execute at all. If a future Godot unwinds the whole
stack, the driver never reaches its summary and the gate fails with
`no-driver-summary` -- which is the loud failure the spec asks for, and the
marker line points at the cause.

### Reshape the existing rejection group

`test_varcall_arg_rejection` in `main.gd` currently ends with the rejection and
relies on a following function to assert survival. With the primitive, the
cleanup and assertions move back next to the call where a reader expects them,
and the split disappears.

## Risks / Trade-offs

- **[Member-var state is global; nested or interleaved use clobbers it] →**
  Document that the flag is read immediately after the call and is not a record.
  A save/restore wrapper is available if nesting ever turns up, but adding it now
  would be speculative.
- **[The flag lies if `callv` fails for an unrelated reason] →** A method-not-found
  error also abandons the frame and leaves the flag at `true`, reading as "was
  rejected" when the real cause is a typo in the method name. Mitigate by
  asserting the method exists before the call inside the primitive, so a typo
  fails as a typo.
- **[Relies on undocumented engine behaviour] →** Accepted and pinned by the
  self-check above. The failure mode is a loud harness failure, not a false pass.
- **[A helper that prints could be noisy] →** One line per expected rejection;
  there are a handful. Cheap against the diagnostic value.
