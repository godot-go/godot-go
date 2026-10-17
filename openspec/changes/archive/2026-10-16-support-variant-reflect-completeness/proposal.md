# Proposal: support-variant-reflect-completeness

## Why

The roadmap's A3 bucket — "some real signatures still can't cross" — resolved in exploration into three concrete, code-verified gaps:

1. **The Ref family is half-dead.** `RefXxx` generated interfaces decode on both paths, but every other Ref spelling fails: bare `builtin.Ref` parameters hit `log.Fatal` (name-key hack yields `""`), `RefXxxImpl` struct parameters reach an assert that can never succeed (`method_bind_reflect.go:457` — `ToObject()` returns `*XImpl` wrappers which don't implement `Ref`), ptrcall `*RefXxxImpl`/`*RefBase[T]` panic at `:1144`, and the ptrcall struct Ref branch is guarded by `strings.HasPrefix(t.String(), "gdextension.Ref")` — a package name the codebase no longer has, so 45 lines below it are unreachable. Ref parameters also **borrow** (`NewRef`) with no `+1`, so a Go method that stores the argument dangles where godot-cpp would hold a reference; non-refcounted or wrong-class arguments produce naked panics that abort the host.

2. **ptrcall lacks the object instance-binding resolution varcall gained.** A user-defined class passed to an object-typed parameter works through `call()` but panics (`:659`) on a direct typed call, because ptrcall only consults `GDNativeConstructors` (engine classes). Two dispatch paths, two behaviors — the asymmetry A1 pinned away for packed types.

3. **Completeness claims are unbacked.** `variant_refect_value.go:80` "TODO: add all types supported by variant" is stale by inspection (both return encoders cover all 36 variant types) but unpinned by tests.

## What Changes

Adopt the Option-A surface ratified in exploration: `RefXxx` interfaces and bare `builtin.Ref` are the supported Ref parameter/return types; decode keys off the declared class (bare `Ref` → `RefCounted`); decoded argument wrappers acquire their own reference and release deterministically at call end (a method that stores the argument must copy it via `NewRefAcquire`); `nil` → zero `Ref` on both paths; wrong-class/non-refcounted arguments fail as contained call errors (the A2 stance, stricter than godot-cpp's silent null, matching gdext); `RefXxxImpl`/`*RefBase[T]` signatures are rejected at bind time, deleting the dead branches rather than fixing them. ptrcall mirrors varcall for object and Ref parameters: instance-binding lookup first, `GDNativeConstructors` fallback, class guard, contained errors. A fixture audit pins every variant type across argument/return × varcall/ptrcall and retires the stale TODO.

## Capabilities

- **New:** `ref-method-boundary` — Ref-typed parameters/returns of reflectively bound methods on both dispatch paths: supported types, class keying, ownership, nil and mismatch behavior, bind-time rejections.
- **Modified:** `ptrcall-object-argument-decode` — object parameters resolve through the instance binding (user classes) before the engine-class fallback; decode failures become contained call errors.

## Non-goals

- The D1 fatal-panic policy sweep (remaining panics are enumerated here, not re-litigated).
- Static methods/variables (B2), signals/await (C1/C2).
- Generics-based typed `Array[T]`/`Dictionary[K,V]` (already split out).
- Silent-null mismatch semantics à la godot-cpp; acquire is decided, contained-error wins.
- Generated-class call sites (they use typed ptrcall paths, not the reflect binder).
