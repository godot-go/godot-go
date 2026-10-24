# Proposal

## Why

Generated methods that return an engine object hand back a wrapper typed to the
method's *declared* class: the template declares `var ret NodeImpl`, the engine
writes the object pointer into it, and the binding returns `&ret`. So
`node.GetChild(0, false).(Label)` reports false for a real `Label`, and every
consumer that wants the runtime class has to call
`ObjectFromInstanceBinding` first. The `ObjectCastTo` retirement documented that
as the migration, and downstream demos already carry a hand-rolled helper per
package to do it.

## What Changes

- Generated **non-refcounted** object returns resolve the returned owner through
  the instance binding and return the wrapper bound for the object's runtime
  class, so a plain Go type assertion to the runtime class succeeds. The
  resolution happens inside the generated method: callers never call a helper or
  re-resolve a binding themselves.
- A return whose binding cannot be resolved (no registered callbacks, e.g. a
  Godot-owned custom-class instance) falls back to today's declared-class
  wrapper, so no return path starts failing that used to work.
- Two returns of the same engine object yield the identical wrapper, aligning
  returns with the identity the argument-decode path already guarantees.
- Add a test witness: a Go method narrows a generated return with a plain type
  assertion and checks repeated returns are the same wrapper.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `instance-binding-lookup`: add a requirement that a generated object-return
  path resolves to the wrapper bound for the object's runtime class, and falls
  back to a declared-class wrapper when no binding can be resolved.

## Impact

- `cmd/generate/gdclassimpl/classes.go.tmpl` — the object-return branch.
- `pkg/gdclassimpl/classes.gen.go` — regenerated (~202 object-return sites).
- `pkg/gdclassimpl/object_return.go` — hand-written, unexported
  `resolveGoStructObjectReturn` helper; the generator does not own this file.
- `test/pkg/`, `test/demo/main.gd` — new witness method and assertion.
- `docs/overview.md` — capability-map row note for the return-resolution path.

## Non-goals

- No cast API (`ObjectCastTo`, `CastTo[T]`, or any replacement) comes back.
- No change to refcounted returns: `NewRef*GDExtensionReturnOwner` keeps owning
  the reference the return slot transfers. Resolving those through the binding
  would leak or dangle, so they stay untouched.
- No change to object argument decode, binding creation, or the inheritance
  check.
- No new capability: the behaviour belongs to the existing binding-lookup
  capability.
