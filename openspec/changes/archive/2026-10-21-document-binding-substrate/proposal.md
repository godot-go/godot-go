# Proposal

## Why

Of the project's 41 specs, none owns the substrate every bind path is built on:
extension bootstrap and init levels, class registration, the reflect-type to
Variant mapping, property access dispatch, vararg method binding, and the object
wrapper/cast base. Those behaviours exist and are exercised by the demo suite,
but they are only *referenced* by other specs, so an intentional change to them
would violate no requirement and a regression would surface only as a demo
failure.

## What Changes

- Add six new capability specs that ratify the binding substrate as it exists
  today, with requirements and scenarios derived from `pkg/core` and
  `pkg/builtin`.
- Pin the bootstrap contract: registry construction, interface loading, Godot
  version read, initialisation-level selection, per-level callback dispatch, and
  deinitialisation ordering.
- Pin the class-registration contract: naming and embedding invariants, the
  creation-info wiring and flags, constructor/encoder registration, init-level
  capture, and unregistration teardown order.
- Pin the reflect-type mapping table and encoder pairing, including the
  bind-time rejection of unsupported Go types.
- Pin property access dispatch (`_get` / `_set` / `_validate_property`) and the
  `(value, bool)` return convention.
- Pin vararg method binding: flag/type alignment, zero named arguments,
  `CallSlice` dispatch, and nil return for void binds.
- Pin the object wrapper base and cast semantics, including nil-safety.
- Add demo assertions where a newly specified behaviour has no fixture, so each
  requirement has an executable witness rather than only a prose contract.
- **Non-goals**: no marshalling, ownership, pinning, virtual-dispatch, or
  instance-lifecycle behaviour changes. Those are already governed by existing
  specs and are only cross-referenced here. Class metadata registration
  (signals, groups, constants) and logging belong to a separate change. No
  public API changes.

## Capabilities

### New Capabilities

- `extension-bootstrap`: init-level selection, callback registration and
  dispatch, process-lifetime registry construction, engine-class table seeding,
  deinitialisation ordering.
- `class-registration`: `ClassDBRegisterClass`/`UnregisterClass` invariants,
  creation-info flags and callback wiring, registry shape, teardown order.
- `variant-type-mapping`: Go `reflect.Type` to `GDExtensionVariantType` table,
  encoder pairing across ptrcall and varcall, unsupported-type rejection.
- `property-access-dispatch`: engine `_get`/`_set`/`_validate_property` routing
  into Go, borrowed Variant argument contract, bool-guarded returns.
- `vararg-method-binding`: binding and dispatch of Go variadic methods,
  flag/type alignment, zero named arguments, void-return handling.
- `object-wrapper-cast`: `Wrapped`/`WrappedClassInstance` accessor contract,
  nil-safety, and `classdb`-based casting.

### Modified Capabilities

None. The two diagnostic fixes correct behaviour that no existing requirement
lets them violate, so no existing spec needs a delta.

## Impact

- Code: `pkg/core/godot.go`, `classdb.go`, `classdb_callback.go`,
  `method_bind.go`, `variant_reflect_type.go`, `variant_refect_value.go`,
  `pkg/builtin/wrapped.go`, plus the demo fixtures that prove them.
- Specs: six new capability directories under `openspec/specs/`.
- No public API, dependency, or generated-code changes.
