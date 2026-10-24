# Proposal

## Why

Generated methods returning a reference-counted object hand back a `Ref[Declared]`
whose pointee is always a private wrapper typed to the method's declared class.
The caller cannot narrow through `Ptr()`: `cs.GetShape().Ptr().(CircleShape2D)`
reports false even when the engine returned a `CircleShape2D`. The argument-decode
path already resolves through the instance binding (giving the runtime-class
wrapper); returns should match. The non-refcounted return resolution proves the
mechanism; this extends the same resolution to the pointee inside the `Ref`.

## What Changes

- Generated refcounted object returns pass `&ret` through
  `resolveGoStructObjectReturn` before wrapping it in
  `NewRef{Declared}GDExtensionReturnOwner`, so the `Ref` holds the
  runtime-class binding wrapper when one resolves.
- The `Ref` interface type is unchanged (`RefShape2D` stays `RefShape2D`); Go
  interface invariance prevents an automatic `RefCircleShape2D` conversion. The
  caller narrows through `Ptr().(CircleShape2D)` instead.
- Two calls returning the same engine object yield `Ref` values whose `Ptr()`
  are identical (the canonical binding wrapper), while each `Ref` still owns its
  own transferred +1.
- The existing scenario "Repeated returns yield the same wrapper" is scoped to
  non-refcounted returns, since each refcounted return must own its own
  reference and therefore allocates a distinct `RefBase`.
- A stale comment at `test/pkg/ref_method_boundary_tests.go:76-78` claiming
  `Ptr()` yields the declared class is corrected (the argument path already
  returns the runtime wrapper).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `instance-binding-lookup`: extend "Generated object returns resolve to the
  runtime-class wrapper" to cover refcounted returns' pointee; tighten the
  identity scenario wording.
- `object-return-ownership`: add "Pointee resolution does not move the
  transferred reference count" — a new guarantee that the binding lookup inside
  a refcounted return is refcount-neutral.

## Impact

- `cmd/generate/gdclassimpl/classes.go.tmpl` — the refcounted-return branch
  (one-line change wrapping `&ret` with the existing resolver).
- `pkg/gdclassimpl/classes.gen.go` — regenerated (~574 refcounted-return sites
  across 176 classes).
- `pkg/gdclassimpl/object_return.go` — existing `resolveGoStructObjectReturn`
  gains a doc note covering the ref branch; no new helper.
- `test/pkg/return_ownership_tests.go` — new witness: `Ptr()` narrowing after
  a generated ref return, plus `Ptr()` identity across repeated calls.
- `test/demo/main.gd` — assertion calls in `test_return_ownership`.
- `test/pkg/ref_method_boundary_tests.go` — stale comment correction.
- `docs/overview.md` — narrowing section updated for `Ptr()` assertion.

## Non-goals

- No change to the `Ref[T]` interface signature or invariance rules.
- No exported `AsRef[T]` conversion helper.
- No resolution of the `Ref` interface type itself (only the pointee inside it).
- No change to the vararg return path (it cannot produce a Ref).
