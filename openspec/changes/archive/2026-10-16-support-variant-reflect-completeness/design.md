# Design: support-variant-reflect-completeness

## Context

Both dispatch paths decode Ref-typed parameters, but each is defective: varcall keys `GDClassRefConstructors` off `t.Name()[3:]` and `log.Fatal`s on a miss (bare `builtin.Ref` yields `""`), asserts `obj.(Ref)` on struct forms that can never hold (wrappers are `*XImpl`, not `Ref`), and **borrows** (`NewRefXxxAsRef` → `NewRef[T]`) so a stored argument dangles. ptrcall's Ref branch sits behind `strings.HasPrefix(t.String(), "gdextension.Ref")` — a package name the codebase no longer has — so it is dead, and its object branch skips the instance-binding lookup varcall gained, panicking on user-class objects.

Verified anchors:

- `pkg/core/method_bind_reflect.go`: varcall Ref branch 228-246 (`log.Fatal` 237, unchecked `obj.(RefCounted)` 245); varcall object branch 247-296 (binding → name fallback → `Implements` guard); ptrcall object branch ~618-692 (`log.Panic` ~659); dead ptrcall Ref block ~1090-1134.
- `pkg/core/variant_reflect_type.go:210` and `pkg/core/object_class_name.go:49`: `GDClassRefConstructors.Get(tn[3:])` membership probes for "is this a Ref type" — a bare `Ref` name produces `""` and misses.
- `pkg/builtin/ref_generic.go`: three ratified constructors — `NewRefInit` (+1 adopt), `NewRefTransfer` (finalizer, no bump), `NewRef` (borrow, no ownership). No acquire flavor.
- `pkg/gdclassimpl/classes.refs.gen.go:12554`: `NewRefXxxAsRef` is borrow; the generated `*RefBase[Class]` structurally satisfies `RefXxx` (`{Ref; Ptr() Class}`) because `RefBase` has `Ptr() T`. Ref interfaces do **not** embed parent Ref interfaces, so a Go-level `Implements` check between `RefCircleShape2D` and `RefShape2D` cannot express inheritance.
- `GDClassRefConstructors` consumers (grep): the three binder decode sites above plus the two membership probes and tests — the registry's only behavioral consumer is the decode path.
- Engine parity primitives are already loaded in `pkg/ffi`: `ClassdbGetClassTag` + `ObjectCastTo` (`ffi.gen.go:164,185`). godot-cpp's `PtrToArg<Ref<T>>::convert` acquires a +1 released by the C++ destructor; `Object::cast_to` is the engine-side inheritance check (`ClassDB::is_parent_class` under the hood, which handles extension classes registered in ClassDB).

## Goals / Non-Goals

**Goals:** Option-A Ref surface (`RefXxx` + bare `builtin.Ref`) working identically on varcall and ptrcall; decoded arguments own a reference; nil → zero Ref; mismatch → contained error (varcall) / loud typed failure (ptrcall); bind-time rejection of concrete Ref forms with the dead branches deleted; binding-first object decode on ptrcall; a fixture audit retiring the stale completeness TODO.

**Non-Goals:** re-litigating the panic→error policy (D1 lane); anything for generated-class call sites; signals/await/statics.

## Decisions

### D1 — Supported forms enforced at bind time, dead code deleted

In the reflect-type→registration-type mapper (`variant_reflect_type.go`): interface kinds implementing the `Ref` protocol register as `OBJECT` — both generated `RefXxx` and bare `builtin.Ref`. Because `tn[3:]` on the name `Ref` yields `""`, the mapper special-cases the protocol interface itself instead of probing the registry. Struct kinds and pointer kinds whose type set implements `Ref` (`RefXxxImpl`, `*RefXxxImpl`, `*RefBase[T]`) return a registration error naming the offending Go type; registration aborts before the method is callable (extends the A1 bind-time-rejection stance). Consequence: the runtime branches that handled those forms become unreachable and are **deleted** — the varcall Pointer/Struct `Ref` asserts (~430-473), the `gdextension.Ref`-guarded ptrcall block (~1090-1134), and the `log.Panic`s at ~1144/~1152 they fell through to. The name-key hack is replaced at every site; only bare `Ref` keeps a name-independent path.

Alternative rejected: fixing the struct branches to work — ratifies two spellings of one concept and keeps 90 lines of near-duplicate decode.

### D2 — Decode keys off the declared class; bare `Ref` is generic

Ref argument decode (both paths): declared `RefXxx` → class name `tn[3:]` → `GDClassRefConstructors` (membership is now guaranteed by bind time; a miss degrades to a contained/typed error, never `log.Fatal`). The registry's `NewRefXxxAsRef*` yields `*RefBase[Class]`, so a subclass object is sliced to the declared class, matching the spec. Bare `builtin.Ref` → `NewRefAcquire[RefCounted](obj)` directly — no registry key, satisfies the `Ref` protocol interface. Null/nil argument → `reflect.Zero(t)` before any constructor (varcall: `arg.IsNil()`; ptrcall: null object pointer in the cell).

### D3 — Inheritance guard via engine class tags

The generated Ref interfaces share no method identity (see Context), so the wrong-class check goes engine-side: `ffi.ClassdbGetClassTag(declaredClass)` once (cacheable per registration) + `ffi.ObjectCastTo(objPtr, tag)`; null result means the runtime class does not inherit the declared class. This is godot-cpp's `Object::cast_to` mechanism and it accepts user-defined subclasses of engine classes. Bare `Ref` guards with the `"RefCounted"` tag (this also replaces the unchecked `obj.(RefCounted)` type assert at `:245`, which currently aborts the host on a `Node` argument). Varcall mismatch → returned decode error → the A2-contained call-error machinery (method body never runs). Ptrcall mismatch → loud typed failure naming argument index and both class names — ptrcall has no `CallError` slot in the ABI; its reporting shape is D1-policy territory, pinned here only as "no naked reflect panic".

Alternative rejected: walking parent class names via `object_get_class_name` in Go — needs a hierarchy table we don't generate and re-implements ClassDB.

### D4 — Acquire ownership: `NewRefAcquire` + registry swap

Add to `pkg/builtin/ref_generic.go`:

```go
// NewRefAcquire wraps an object the caller merely observes, taking its own
// +1 via Reference(); the finalizer releases it, mirroring godot-cpp's
// PtrToArg<Ref<T>>::convert acquire + destructor release.
func NewRefAcquire[T RefCountedT](obj T) *RefBase[T]
```

Null object (`ObjectArgPtr(obj) == nil`) → zero wrapper, no finalizer (same discipline as `NewRefTransfer`). Extend `cmd/generate/gdclassimpl/classes.refs.go.tmpl` to emit `New{{X}}AsRefAcquired(RefCounted) Ref` beside the existing `AsRef`, and `classes.init.go.tmpl:48` to register the acquired flavor into `GDClassRefConstructors`. The user-facing `NewRefXxx` (Init) and `NewRefXxxAsRef` (borrow) stay unchanged; grep evidence shows the registry's only behavioral consumers are the binder decode paths (the two other readers test key presence only). Decode = acquire, exactly once per decoded argument, released deterministically at call end by `releaseDecodedRefArgs` (both dispatch paths, every return style); the finalizer stays as a safety net for wrappers the callee itself creates. A method that stores the argument must copy it (`NewRefAcquire`) — the decode's `+1` is gone before the callee regains control.

### D5 — ptrcall object decode mirrors varcall

Extract varcall's resolution sequence (binding getter → `Implements` check → `GDNativeConstructors` by runtime class → `Implements` guard, `method_bind_reflect.go:247-296`) into a helper keyed on a raw `GDExtensionObjectPtr`; varcall adapts by passing `arg.ToObject()`, ptrcall by reading the pointer from the argument cell. The ptrcall Interface object branch (~618-692) calls the helper instead of doing class-name lookup only: user-class instances now resolve to their registered Go wrapper (identity parity across dispatch paths), engine classes keep canonical wrapping, subclass params keep working, and the `log.Panic` at ~659 disappears behind the same typed-failure contract as D3. Shared helper, not copy-paste — the A1 lesson was that duplicated decode drifts.

### D6 — Coverage audit and TODO retirement

A synchronous (no `await`, zero frame budget) fixture set in `test/demo/` pins the matrix: every variant type × {varcall, ptrcall} × {argument, return}, plus the Ref scenarios from the spec delta (bind-time rejections asserted via Go tests, crossing asserted via GDScript calls both paths). Rejections in GDScript fixtures use `expect_rejected_call` so an accidentally-uncontained failure can't silently skip the rest of the function. `variant_refect_value.go:80`'s stale "add all types supported by variant" TODO is deleted once the matrix exists — the encoders already cover all 36 types; the audit turns that claim into evidence.

### D7 — Ownership corrections surfaced by the audit

Running the audit in-engine exposed four refcount defects the D4 finalizer model could not paper over, fixed as part of this change:

- **Deterministic decode release.** A finalizer is too lazy to bound the decode's `+1`: acquired refs pile up and release in a cascade at the first explicit `runtime.GC()`. Both dispatch paths now release decoded Ref arguments at call end through `releaseDecodedRefArgs` (every return style, after the return value is encoded), and a method that stores an argument copies it (`NewRefAcquire`) rather than aliasing the decode.
- **`Unref()` memdeletes.** Godot 4.8's `RefCounted::unreference()` returns whether the object should be destroyed but no longer self-deletes; the caller must `memdelete` (godot-cpp's `Ref` destructor does). `Unref()` calls `object_destroy` when `Unreference()` reports the last reference.
- **ptrcall return `reference()`.** `validated_call` initializes an empty OBJECT return slot and never references it, but the return-Variant destructor unreferences; the encoder now references before writing the slot so the destructor balances.
- **Copy-constructor returns.** Dictionary/Callable/Signal/PackedVector4Array return through engine copy constructors, not the generic encoder's raw struct copy, and `isPtrcallBorrowEcho` covers that whole destroy-gated set.

## Risks / Trade-offs

- **Refcount arithmetic changes for existing Ref-parameter methods** (+1 decode, −1 call-end release). Net zero, with release deterministic rather than GC-driven. Retention fixtures must copy (not alias) a stored argument and release it before the engine's leak snapshot, same discipline as the A2 string-retention fixtures.
- **Class tags are engine-class identifiers** — fine here, because declared `RefXxx` parameter types exist only for engine classes (codegen), and bare `Ref` checks `"RefCounted"`; engine-side `is_parent_class` accepts extension classes inheriting them.
- **Divergence from godot-cpp**: godot-cpp silently nulls a mismatched `Ref`; we produce a contained error (gdext-strict, ratified posture from A2).
- **Finalizer thread affinity**: `Unref()` now calls `object_destroy` when it drops the last reference (Godot 4.8 `unreference()` no longer self-deletes; the caller memdeletes, as godot-cpp's `Ref` destructor does). Deterministic decode release runs on the caller's thread, but a wrapper with a live finalizer that is GC'd while holding the last reference memdeletes off-thread; the deterministic release keeps this out of the common path.
