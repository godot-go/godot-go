# Proposal

## Why

Every object-returning generated method leaks one Godot reference per call. A real engine run of 200 `CollisionShape2D.get_shape()` calls through Go ended with `Leaked instance: CircleShape2D ... Reference count: 200` and `1 RID allocations of type 'GodotShape2D' were leaked at exit` — the count equals the call count exactly.

Root cause (`pkg/builtin/ref_generic.go`): the engine's ptrcall return path constructs a `Ref<T>` into the Go return slot, which adds a reference and transfers ownership of it to the receiver. The generated factory `NewRefXGDExtensionIternalConstructor` wraps that transferred reference with `NewRef`, which changes no refcount and sets **no finalizer** — so the transferred +1 is never released. The sibling constructors are correct by contrast: the copy `Ref()` calls `Reference()` and installs a finalizer, and `NewRefInit` calls `InitRef()` and installs one. The scale is 661 generated owning factories (1,323 `NewRef[...]` call sites total).

This surfaced while completing the sibling change `fix-generated-call-pin-leak` (3fe924d). That change made return wrappers collectable — but collectable is not released: with no finalizer, GC reclaims the Go wrapper and leaves the Godot reference orphaned. The pin was never the cause; do not conflate the two.

## What Changes

- A deliberate decision task first: classify every `NewRef` call site as owning vs borrowing, then choose between (a) adding a finalizer to `NewRef` itself — simple, but it changes the contract of 1,323 call sites, and any caller that only borrows would over-release into a premature free — or (b) a distinct owning constructor used only by the generated return path, leaving `NewRef`'s borrowing contract intact. The design justifies (b): the generated `*GDExtensionIternalConstructor` factories are the specific sites that receive an engine-transferred reference.
- Implement the choice in hand-written `pkg/builtin/ref_generic.go` and template `cmd/generate/gdclassimpl/classes.refs.go.tmpl`; propagate with `make generate`.
- Preserve idempotence with the existing `Unref()` (which already clears the finalizer), so a caller that unrefs manually is never double-released by a later finalizer.
- Add regression tests for refcount stability across repeated returns, drop-then-free, manual-unref idempotence, borrow safety, and a clean engine exit leak check.

## Capabilities

### New Capabilities
- `object-return-ownership`: Ownership semantics for Godot references transferred to Go through generated object-return paths — each transferred reference is owned by the returned wrapper and released exactly once on drop or explicit unref, while borrowing wrappers never release.

### Modified Capabilities
(none)

## Impact

- `pkg/builtin/ref_generic.go` (owning constructor) and `cmd/generate/gdclassimpl/classes.refs.go.tmpl` → regenerated `pkg/gdclassimpl/classes.refs.gen.go`.
- `test/pkg/` and `test/demo/main.gd` regression tests.
- No public API signature changes.

## Non-goals

- Not changing `NewRefInit` or copy `Ref()` semantics (already correct), and not changing the borrowing contract of `NewRef`/`NewRefXAsRef`.
- Not editing `*.gen.*` files directly; all changes flow through templates and `make generate`.
- Not redesigning the Ref ownership model (no dispose-only discipline, no pools), and not touching argument-decode or variant-return paths covered by sibling changes.
