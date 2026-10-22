# Design

## Context

See `proposal.md` for motivation. The constraints that shape the approach:

- Three registration paths in `pkg/core/classdb.go` diverge from the reference
  implementation and the engine: `ClassDBAddSignal` checks duplicates against
  `ClassInfo.SignalNameSet` only; `ClassDBBindBitfieldFlag` passes an empty enum name to
  the shared integer-constant helper; the signal duplicate path reports
  "Constant already registered."
- The engine's own guards are the authority on what gets registered. `GDType::add_signal`
  rejects a name already in its member set, which includes inherited names, and the
  rejection is an error print plus a dropped registration — the binding learns nothing.
  `GDType::bind_integer_constant` stores bitfield-ness on the enum member, so an empty
  enum name discards the flag's bitfield-ness entirely.
- The binding holds no signal metadata for engine classes: no generated file enumerates
  signals. Only the engine knows that a `Node`-derived class already carries `ready`.
- Two usable engine introspection primitives are already generated and unused:
  `ClassDB.class_has_signal(class, signal)` (inheritance-inclusive, no flag) and
  `ClassDB.class_get_signal_list(class, no_inheritance)`. Reachable through the existing
  generated singleton accessor.
- The run-output gate treats engine `ERROR` lines as evidence only when the driver
  summary is missing, so an engine-side rejection cannot fail a run by itself.

## Goals / Non-Goals

**Goals:**

- Make the Go-side duplicate gate agree with the engine's own acceptance set, so a
  registration that Go reports as successful is one the engine kept.
- Make the enum and bitfield arguments of constant binding mean what they say.
- Give each requirement in both new specs an executable witness, using metadata
  introspection rather than process output where possible.
- Keep the boundary with the four adjacent capabilities named in the proposal.

**Non-Goals:**

- No Go-side table of engine-class signals, and no new generated surface.
- No new engine calls at method-dispatch or per-frame hot paths; registration is
  start-up work.
- No change to how logging is configured beyond what is already observable, and no
  engine-routed output.
- No change to signal parameter `class_name` or to the property-group failure style.

## Decisions

### D1 — Ask the engine whether the class already carries the signal name

Use `ClassDB.class_has_signal(className, signalName)` as the duplicate gate, replacing
the `SignalNameSet` lookup, and fail through the library's fatal path naming class and
signal.

The engine's answer is derived from the same member set its registration guard uses, so
the check cannot disagree with the call it protects, and inherited names are covered
without the binding knowing them. godot-cpp walks its own parent chain only because it
registers engine classes itself; that option does not exist here.

Alternatives rejected:

- *Go-side parent walk.* `ClassInfo` knows its parent's name but nothing about the
  parent's signals, so the walk stops at the first engine class — precisely the case
  that matters.
- *A generated signal table from `extension_api.json`.* Correct but adds a generated
  surface with a freshness obligation, for a check that runs a handful of times at
  start-up.
- *`class_get_signal_list` plus a linear scan.* Same answer as `class_has_signal`, more
  allocation and more code. Kept as the fallback if the spike finds `class_has_signal`
  unusable at registration time.

Ordering matters: the Go-side `SignalNameSet` is still the record of names this class
registered, so keep the existing write-through after a successful engine answer, and
keep the rejection *before* any name is retained so a rejected registration leaves no
trace.

Spike outcome (task 1.1, confirmed — no fallback needed): from inside
`RegisterClassExample` at scene-initialization time, `class_has_signal("Example", …)`
reported `true` for `ready` (carried only by inheritance from `Node`), `true` for
`custom_signal` (registered moments earlier in the same pass), and `false` for an
unused name (`test-output.log`, three `ratify-1.1 class_has_signal` WARN lines). The
boolean query answers correctly during class registration, so the gate uses
`class_has_signal` and does not switch to `class_get_signal_list(class, false)`.

Hand-written boundary and its drift sentinel: `pkg/core` deliberately keeps no
production edge to the generated `pkg/gdclassimpl` layer — the substrate sits below
codegen — so the probe hand-writes the method-bind call and duplicates the
`class_has_signal` hash literal from the generated `ClassDBImpl.ClassHasSignal` body
(`pkg/core`'s tests import the generated layer; its production code does not). To keep
that copy from going stale unnoticed, the
run-output gate fails the suite on the degradation warning (`signal-gate-degraded` in
`test/check_test_output.sh`): the demo registers signals on every run, so a hash that
drifts after an extension-api refresh turns CI red instead of quietly reverting to the
engine-side silent drop.

### D2 — Fail loudly rather than decline

An inherited-name collision aborts with a diagnostic naming class and signal, matching
godot-cpp's message shape and the sibling registration paths. Declining and continuing
was rejected: the signal would not exist, and the author would hold a registration that
silently does nothing, which is the current bug with a nicer log line.

### D3 — Pass the enum name through, and adopt the engine's qualified-name shorthand

`ClassDBBindBitfieldFlag` forwards `p_enum_name` instead of `""`, so the enum member is
created and records bitfield-ness. The existing demo convention of qualifying an enum
name as `Class.EnumName` is kept and documented rather than normalised in Go: the engine
already reduces the name to the portion after its first separator. Normalising in Go
would create a second rule to keep in step with the first.

The consequence is accepted: an enum name now becomes a class member, so it can collide
with an existing member name or with an enum owned by a parent class, and the engine
rejects the constant in that case. The spec's uniqueness requirements plus witnesses
cover what Go can see; the engine's rejection of a parent-owned enum name stays an
engine error.

### D4 — Witness by introspection, and by a recovered panic

- Enum and bitfield behaviour is asserted from GDScript against `ClassDB` metadata
  queries: enum present, enum reported as bitfield, constant's owning enum, and
  unqualified constant values unchanged. These assertions fail before the fix and pass
  after it, which is the point.
- The inherited-signal rejection cannot be asserted by a normal demo assertion, because
  a fatal aborts the run. Use the existing recovered-probe pattern: a Go fixture that
  registers a colliding signal inside `recover()` and reports what it caught, driven from
  the demo, so the run survives and the assertion is real.
- Start-up-order and no-run-fatal requirements are verified by inspection, following the
  precedent set by the substrate specs.

### D5 — Log the ratified surface, leave the unused knobs alone

The logging spec describes the surface in use: environment-selected threshold applied at
start, standard output with serialized writes, stack at warning and above, the
record-then-unwind fault path, the single recovered boundary, teardown availability. The
unused level and writer setters and the unreferenced fatal helper are deliberately left
unspecified, so ratifying today's surface does not forbid or freeze them. Engine
print-interface routing stays unloaded-by-use and is specified as not used.

## Risks / Trade-offs

- **The query may not answer correctly at registration time** → Make this task 1: prove
  `class_has_signal` reports inherited names for a freshly registered extension class
  from inside an initializer, before writing the gate. If it does not, fall back to
  `class_get_signal_list(class, false)`, and if that also fails, stop and revisit rather
  than shipping a gate that consults an unready database.
- **A working extension now aborts at start-up** (its signal name collided and was
  silently dropped before) → Intended, and the diagnostic names class and signal; state
  it in the release note as a behaviour change.
- **Engine call during registration could fail if the singleton is unavailable** → The
  singleton accessor yields nil rather than faulting, so check for nil and fall back to
  the existing Go-side set, logging a warning that inheritance was not checked. A
  degraded check is no worse than today's, and the run-output gate fails the suite when
  that warning appears, so the degraded path cannot persist unnoticed (D1's sentinel).
- **Passing the enum name through surfaces new engine-side enum collisions** → Add a
  witness binding two distinct flag enums and one intentionally distinct constant name,
  and confirm the run is clean; enum-name collisions remain engine-reported. The release
  note should state that an enum name now becomes a class member, so a name colliding
  with an existing or parent-owned member is rejected where it once registered.
- **The gate cannot see engine errors** → Metadata witnesses assert metadata positively
  from GDScript instead of relying on output absence.
- **Ratifying the parameter class-name field cements a divergence from the reference** →
  Recorded as a ratified divergence; a later change can align it, at which point the
  requirement is modified deliberately rather than silently.

## Migration Plan

No data or API migration. Sequence: prove the query's behaviour, land the duplicate gate
with its diagnostic and probe witness, land the enum pass-through with its introspection
assertions, then the spec witnesses that need no code change, then documentation.
Rollback is reverting the commits; the engine-visible metadata reverts with them, and no
committed artifact depends on the new behaviour.

## Open Questions

- Should signal parameter `class_name` eventually report the parameter's own class for
  object-typed parameters, matching the reference? Deferred with the ratified divergence.
- Should property groups be tracked per class for de-duplication and teardown? Deferred;
  today they are write-only passthroughs and no requirement depends on the record.
