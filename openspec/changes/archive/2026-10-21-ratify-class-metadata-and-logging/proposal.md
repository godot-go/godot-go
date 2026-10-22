# Proposal

## Why

Class metadata registration — signals, constants and enums, bitfield flags, property
groups — has no capability spec, and three of its paths publish metadata the engine
then discards or rejects: a duplicate signal name that a parent class owns passes the
Go check and is silently dropped by the engine, a bitfield flag's enum name is thrown
away so the flag's bitfield-ness is never recorded, and the duplicate-signal failure
reports itself as a constant error. Logging is equally unspecified even though the
six substrate specs ratified last change all defer to "the library's fatal-error
channel" that nothing defines.

## What Changes

- New `class-metadata-registration` spec: signal parameter metadata as delivered to
  the engine, duplicate-name gating, enum grouping and bitfield-ness, constant-name
  uniqueness, property-group registration and its diagnostics.
- Fix: duplicate signal detection consults inherited signal names via engine
  introspection (`ClassDB.class_has_signal`) instead of only this class's own
  set, and fails loudly naming the class and the signal. **BREAKING** a registration
  that today silently registers nothing now aborts startup with a clear diagnostic.
- Fix: `ClassDBBindBitfieldFlag` passes its `p_enum_name` through to the engine, so
  the enum member exists and records bitfield-ness. That argument stops being inert;
  its signature is unchanged. The enum name becomes a class member, so a name colliding
  with an existing or parent-owned member is now rejected by the engine where it
  previously registered anonymously — a release-note item.
- Fix: the duplicate-signal diagnostic reports a signal collision, not a constant one.
- New `diagnostic-output-channel` spec: level selection and its single configuration
  point, output destination, stack-trace policy, the fatal path binding faults use,
  the one boundary that recovers a user closure panic, and logging during engine
  teardown.
- Witnesses: GDScript metadata-introspection assertions for enum, bitfield-ness and
  inherited-signal rejection; a recovered-panic probe for the new fatal; a run-output
  gate sentinel that fails the suite when the duplicate gate degrades.

## Capabilities

### New Capabilities

- `class-metadata-registration`: how a Go-defined class publishes signals, constants,
  enum members, bitfield flags and property groups to the engine class database, and
  how name collisions are gated before the engine call.
- `diagnostic-output-channel`: the binding's log levels, destination and
  configuration, the fatal path used for binding faults, and which callback boundary
  converts a panic into a reported failure.

### Modified Capabilities

None. `class-registration` keeps authority over class-level invariants and method or
virtual name collisions; `method-call-error-reporting` keeps authority over call
argument rejection; `native-stacktrace-capture` keeps authority over C-side capture;
`test-harness-integrity` keeps authority over the run-output gate. All four are
cross-referenced, not restated.

## Impact

`pkg/core/classdb.go` (signal registration, integer-constant binding), `pkg/log`,
`test/pkg/*.go` fixtures, `test/demo/main.gd` assertions. No exported signature
changes. Engine-visible class metadata changes for bitfield flags: an enum member
appears where none existed. Startup behaviour changes for inherited-name signal
collisions.

## Non-goals

- Not changing signal parameter `class_name`, which reports the emitting class.
- Not changing property-group failure style (a raw panic) or adding group
  de-duplication.
- Not routing library logs to the engine's print interface, and not wiring or removing
  the unused level and writer setters.
- Not adding a Go-side registry of engine-class signals; the engine is the source of
  truth for inherited names.
- Not touching the call-argument rejection channel or the codegen pipeline.
