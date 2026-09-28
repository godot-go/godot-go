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

### Detect rejection from the return value, not from control flow

**Chosen: `last_call_was_rejected = obj.callv(method, args) == null`.**

The first draft of this design said to set the flag *before* the call, on the
reasoning that a rejected call never reaches a `return`, so only a
pre-set value survives. That reasoning came from watching `example.call(...)`
abandon its frame, and it does not survive contact with `callv`.

Measured, the two dynamic call forms behave differently on the same rejection:

| Form | On rejection |
|---|---|
| `obj.call("m", bad)` | `SCRIPT ERROR`, frame abandoned, everything after it is dead code |
| `obj.callv("m", [bad])` | engine `ERROR` logged, **returns `<null>`, caller continues** |

`callv` is the better foundation: it removes the stranding problem instead of
containing it. And because a rejected call yields `null` where an accepted one
yields the method's value, rejection is directly observable after the fact --
no before-the-fact flag required.

**Constraint this introduces:** a void method also yields `null`, so for a void
target the flag cannot distinguish rejection from normal completion. The
primitive is for value-returning targets; a void target must assert an
observable side effect instead, such as a counter the body increments.

**Rejected: set the flag before the call.** Built on the wrong premise, and it
failed exactly the way a wrong premise does -- quietly. The first run reported
`last_call_was_rejected == false` on a call that *was* rejected, because
`callv` returned normally and cleared the flag.

**Rejected: `call()` with the primitive containing the abandonment.** It would
work, but `callv` is strictly better: there is nothing to contain.

### The marker line still earns its place

The original reason for the marker was a broken containment assumption. That
reason is gone, but the line still pays for itself: the engine's rejection
surfaces as an `ERROR` naming the method and argument conversion, and the marker
identifies which test induced it rather than leaving the error to be matched
back by hand.

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
