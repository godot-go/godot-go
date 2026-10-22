# Design

## Context

See proposal.md — Why. The relevant constraint is that the substrate is not a
blank slate: `pkg/core` already implements it, the demo suite already exercises
most of it, and 41 specs already govern adjacent behaviour. Six areas have no
owner, yet other specs cite their symbols as if they were defined:
`stringname-lifecycle-management` cites the reflect encoders,
`class-instance-lifecycle` cites construction helpers, and
`method-call-error-reporting` assumes bind-time signature gates exist.

The specs describe observable behaviour, so this change extracts the contract the
code already honours rather than inventing a new design. Where the code and a
plausible contract diverge, the divergence is recorded rather than silently
resolved (see Open Questions).

## Goals / Non-Goals

**Goals:**

- Give each unowned substrate area one capability spec with testable requirements.
- State the boundary between a new capability and every existing spec that
  already touches the same call path, so authority is unambiguous.
- Give each requirement an executable witness where the current demo or unit
  suite can provide one, and add one where it cannot.

**Non-Goals:**

- No refactor of `pkg/core` or `pkg/builtin`; the change is additive specs plus
  demo witnesses.
- No class-metadata semantics (signals, property groups, constants, enums,
  bitfields) and no logging contract; those belong to the follow-on change.
- No generated-surface or codegen contract.
- No re-specification of ownership, refcount, pinning, or virtual-dispatch rules;
  those specs stay authoritative.

## Decisions

### D1 — Ratify observed behaviour instead of designing replacement architecture

The specs are written from the code, not from an idealised model. Rewriting the
substrate to a cleaner design would make this a high-risk refactor with no user
value, and writing only what tests already assert would leave the cited symbols
undefined.

Alternative rejected: spec a target design first, then change the code to match.
That inverts the purpose of retrofitting a contract and turns every incidental
difference into a migration.

### D2 — Six capabilities, not one `binding-substrate` catch-all

Each capability owns one call path: bootstrap, registration, type mapping,
property access, vararg binding, wrapper/cast. A single catch-all would become
the place every future requirement is appended, defeating discovery.

Alternative rejected: one spec. It would exceed the 500-character requirement
guidance on nearly every requirement and blur the boundary table below.

### D3 — Boundary ownership is explicit

| Concern | Owning spec | This change's stance |
| --- | --- | --- |
| Instance creation, free guard, construction info | `class-instance-lifecycle` | Not restated; registration only states what it installs |
| Binding-slot resolution and handle release | `instance-binding-lookup` | Cast defers resolution to the binding lookup |
| Ownership of decoded args/returns, borrow echoes | `borrow-echo-detection`, `container-lifecycle-management`, `object-return-ownership` | Type mapping only selects the variant type |
| Variant copy vs view helpers | `variant-copy-ownership` | Property set only states the delivered value is borrowed |
| Pins and process-lifetime retention | `native-call-pinning` | Bootstrap only states registries persist |
| Virtual naming, dispatch depth, defaults | `hierarchical-virtual-methods`, `gdvirtual-unimplemented-defaults` | Registration owns only namespace collisions |
| Arity/type rejection at call time | `method-call-error-reporting`, `varcall-argument-decode` | Vararg binding owns only bind-time gates and dispatch shape |
| String/StringName lifetime at the boundary | `string-boundary-marshalling`, `stringname-lifecycle-management` | Not restated |
| Demo harness gates | `test-harness-integrity` | Demo witnesses are added, harness rules unchanged |

### D4 — Requirements are phrased behaviourally, not structurally

Requirements name engine-facing concepts (`_get`, vararg flag, class tag,
creation-info callbacks) and public Go entry points, but not internal helpers.
That keeps a future implementation free to rename or split helpers without a
spec change, while still pinning what a caller and the engine can observe.

### D5 — Demo witnesses are added only where a requirement has none

Existing witnesses already cover most requirements: `test/demo/main.gd` exercises
zero-argument varargs (`varargs_func()`), `_get`/`_set`-only properties
(`dproperty_*`), signal registration, scene-level init/terminator, and the
32-type marshalling matrix. Unit tests already pin slice rejection
(`pkg/core/variant_reflect_type_test.go`). `ObjectCastTo` has no caller and no
witness, so the change adds a demo assertion for cast identity, cast rejection,
and the inheritance check; the binding-creation path a cast can take is stated in
the requirement instead of being asserted, because reaching it needs an object
this library has never wrapped.

## Risks / Trade-offs

- **Specs freeze incidental implementation choices** → Phrase each requirement
  as an observable outcome; leave symbol names out except engine-facing ones and
  public entry points.
- **A new spec contradicts an existing one at archive time** → The boundary table
  in D3 is checked against every existing spec's requirements before archiving;
  the archive step merges deltas and will surface a conflicting requirement name.
- **Some behaviours are hard to witness without aborting the process** (bind-time
  panics, no-initializer startup) → Those requirements are stated as fatal
  outcomes and verified by inspection; no test is added that would kill the demo
  run.
- **Ratification can cement a defect** → Requirements avoid defect-specific text:
  the property callback requirement says "fails loudly" rather than naming the
  panic, and signal diagnostics are out of scope for this change.
- **Requirement descriptions must stay short** → Each requirement states one
  behaviour; examples live in scenarios.

## Migration Plan

Archive merges the six delta specs into `openspec/specs/`, taking the capability
count from 41 to 47. No code migration, no API change, no dependency change.
Rollback is reverting the archive commit; the delta files and new main specs are
both reversible.

## Divergences Recorded During Apply

The draft deltas described a plausible contract in three places where the code
does something narrower. Each was corrected to ratify the code, not the draft:

- A Variant **parameter** advertises the nil type without the "nil is variant"
  usage flag; only a Variant **return** carries the flag. godot-cpp sets the flag
  on both, so a later change could align the parameter spelling.
- A pointer to Variant is **not** mapped to the untyped marker; the pointer branch
  only recognises object carriers, so `*Variant` fails the bind. The dead
  Variant check in that branch is therefore unreachable.
- A successful cast resolves through the instance binding, which the engine
  **creates** from the registered callbacks when nothing is bound yet; the draft
  claimed it never produces a new wrapper.

## Verification Baseline

`go test ./...` is not green on `main` and this change does not repair it. Two
pre-existing failures are unrelated to the substrate specs:

- `pkg/builtin`: `TestRefAcquireUnrefReleasesExactlyOnce` and
  `TestRefAcquireFinalizerReleasesWhenUnreachable` segfault, because
  `RefBase.Unref` calls the engine's object-destroy slot once the refcount drops
  and that slot is nil outside a running extension.
- `test/concepts/array_type_equality` and
  `test/concepts/generic_struct_memory_layout`: `go vet` printf diagnostics.

Both reproduce at the base commit. The change's own gate is therefore
`go test ./pkg/core ./pkg/builtin` (with those two engine-dependent tests
skipped), `go build ./...`, and the `make test` runtime and editor gates.

## Open Questions

- Should property get/set on an instance whose class registration has already
  been torn down become a silent no-op, as notification and revert already are?
  Today it is fatal; changing that is a separate behaviour change.
- Should signal duplicate detection walk parent classes like the reference
  implementation, and should the diagnostic name the class and signal? Deferred
  to the class-metadata change.
- Should a cast to a class name the engine does not publish fail loudly or yield
  nil? Today it is loud; either is defensible, so the spec records the current
  behaviour and a later change can revisit it.
