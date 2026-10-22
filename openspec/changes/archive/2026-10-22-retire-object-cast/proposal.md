# Proposal

## Why

`builtin.ObjectCastTo` is the library's only cast entry point, but nothing in the
binding substrate calls it. The decode paths that genuinely need a cast use the
pieces instead — `ObjectInheritsClass` plus `ObjectFromInstanceBinding`
(`pkg/core/method_bind_reflect.go:161` and `:180`) — so they can report a typed
error naming both classes. The function's only callers are the demo witness that
was written for it and one demo input handler. Because it is stringly typed,
returns nil on every failure, panics on an unknown class name, and spends two
engine calls on `log.Info` per cast, its shape cannot be improved without
rewording the requirement that ratifies it. Retiring it makes the spec describe
what callers actually do.

## What Changes

- Remove `builtin.ObjectCastTo`. Narrowing a wrapper becomes a Go type assertion
  on the wrapper the binding resolved, which works because generated wrappers
  embed their ancestor chain; asking whether an object is an instance of a class
  or subclass stays `builtin.ObjectInheritsClass`.
- Rewrite the demo witness `Example.TestObjectCast` against that pair: identity
  through `ObjectFromInstanceBinding`, rejection through a failed assertion, and
  the inheritance predicate for ancestors and unrelated classes.
- Replace the cast in `Example.V_Example_Input` with a type assertion on the
  decoded ref.
- Drop the `ObjectCastTo` entry from the pin-scoping allowlist, whose documented
  reason (a retained callbacks pointer) no longer exists in the implementation.
- **BREAKING**: downstream code that called `builtin.ObjectCastTo` must use
  `ObjectFromInstanceBinding` (or the wrapper it already holds) plus a type
  assertion. The project's API is experimental and unversioned.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `object-wrapper-cast`: remove the "Casting uses the engine's class tag and
  resolves through the instance binding" requirement and its five scenarios, and
  reword the capability purpose so it no longer promises a cast entry point. The
  inheritance-check requirement and the four wrapper requirements stay unchanged.

## Impact

- `pkg/builtin/wrapped.go` — delete `ObjectCastTo`.
- `test/pkg/example.go` — rewrite `TestObjectCast`, rewire `V_Example_Input`.
- `test/demo/main.gd` — keep the `test_object_cast` assertion, update its comment.
- `pkg/gdclassimpl/pin_scoping_test.go` — drop the stale allowlist entry.
- `docs/overview.md` — remove `builtin.ObjectCastTo` from the capability map.
- `openspec/specs/object-wrapper-cast/spec.md` — one removed requirement.

## Non-goals

- No replacement cast API: no generic `CastTo[T]` and no error-returning
  pointer-taking variant in this change.
- No change to `ObjectInheritsClass`, `ObjectFromInstanceBinding`,
  `ObjectFromVariant`, or `ObjectClassName` behaviour.
- No change to Ref decode, object-argument resolution, or binding creation.
- No new demo coverage beyond rewiring the existing witness.
