# Design

## Context

See `proposal.md` — Why, for the defect and its blast radius.

The constraint that shapes the fix is that the encoding happens in a Go text template (`cmd/generate/gdclassimpl/classes.go.tmpl`), so whatever is decided here is replicated across every generated wrapper. The template already distinguishes the two cases it needs to — `{{ if $view.ContainsClassName ... }}` selects engine-class arguments — and currently ignores the distinction. The varcall path is not involved: it marshals through `Variant`, whose object encoding is already correct.

Two wrapper families exist and they disagree about pointer accessors, which is the main trap:

| Accessor | `WrappedImpl` | `WrappedClassInstance` |
|---|---|---|
| `AsGDExtensionObjectPtr()` | `unsafe.Pointer(w.Owner)` — the object pointer | `unsafe.Pointer(w.Instance.GetGodotObjectOwner())` — the object pointer |
| `AsGDExtensionTypePtr()` | `unsafe.Pointer(&w.Owner)` — **address of** the field | the object pointer **itself** |

`AsGDExtensionObjectPtr()` is consistent across both families. `AsGDExtensionTypePtr()` is not.

## Goals / Non-Goals

**Goals:**

- One encoding rule for engine-class arguments, generated uniformly, correct for both plain classes and `Ref[T]`.
- Nil-safety without a `recover()` in generated code.
- A regenerated diff that is reviewable: only object-argument bodies should move.

**Non-Goals:**

- Making the template prettier. The duplicated branch is the bug's fingerprint; it stays, with a correct body.
- Touching varcall, return paths, or value-type arguments.
- Reducing per-call overhead. The added work is one interface dispatch per argument against a boundary that already costs tens of microseconds; it is not a performance change.

## Decisions

### Encode through `AsGDExtensionObjectPtr()`, into a local variable

The generated body becomes:

```go
var argObj_0 GDExtensionObjectPtr = node.AsGDExtensionObjectPtr()
argPtrSlice[0] = (GDExtensionConstTypePtr)(unsafe.Pointer(&argObj_0))
```

The slot must point at a cell whose *contents* are the object pointer. Binding a named local makes that explicit and gives the pointer a stable, pinnable address.

**Rejected — `arg.AsGDExtensionTypePtr()`:** one call, no local, and wrong for half the wrapper families. `WrappedImpl` returns `&w.Owner` while `WrappedClassInstance` returns the pointer value, so the generated code would mean different things depending on which family implements the argument. This is exactly the class of subtle boundary bug this change exists to remove.

**Rejected — type-switching on concrete class types:** 164 argument types, and the interface already exposes the accessor uniformly. It would also add a second instance of the event-type-discrimination pattern that is already a known limitation elsewhere in the bindings.

### Reach `Ref[T]` arguments through `ToObject()`

`Ref` exposes `ToObject() RefCounted`, which returns the held value with no refcount side effect, and `RefCounted` reaches `Wrapped` and so `AsGDExtensionObjectPtr()`. `Ptr()` is deliberately not used: it is absent from the `Ref` interface, so the template cannot rely on it for a generically-typed argument.

### Harden `AsGDExtensionObjectPtr()` against a nil receiver rather than guarding in the template

A nil interface value cannot have a method called on it at all, and a *typed* nil — an interface holding a nil `*WrappedImpl` — passes an `== nil` check and then panics on field access inside the accessor. Guarding both cases in the template would mean emitting reflection or `recover()` into every one of 794 wrappers.

Instead, make the accessor itself nil-receiver-safe (`if w == nil { return nil }`) and have the template emit a single nil-interface check. Invalid references are already covered: `RefBase.IsValid()` reports false for a zero-valued held reference, and the resulting owner pointer is nil, which is the null Godot wants.

**Rejected — `recover()` in generated code:** hides real nil dereferences behind a catch-all, in generated code nobody reads.

### Do not add reference counting

Godot's C++ receiving code refcounts on assignment; a call argument is borrowed, not owned. The commented-out `.Reference()` in the template reflects a misreading of that and must stay commented. Adding a reference here would leak one per call — the same failure mode `container-lifecycle-management` had to unwind.

## Risks / Trade-offs

- **[Large regenerated diff — 794 methods change shape] → Mitigation:** regenerate and diff filtered to object-argument lines; confirm no value-type or return-path body moved. Full test suite before merge.
- **[Typed-nil slipping past a nil-interface check] → Mitigation:** nil-receiver hardening in `WrappedImpl`/`WrappedClassInstance` plus explicit typed-nil and invalid-`Ref` test cases, not just a plain-nil test.
- **[`Ref.ToObject()` assumed refcount-free] → Mitigation:** verified against `RefBase.ToObject()`, which returns the held value directly; add a regression test asserting the reference count is unchanged across repeated argument passing, so a future change to `ToObject()` breaks loudly.
- **[A class-typed argument that Godot actually expects as a `Variant`] → Mitigation:** the template's existing `ContainsClassName` predicate already separates class-typed from built-in-typed arguments; the fix stays inside that predicate and does not reclassify anything.
- **[Trade-off: one extra local and interface dispatch per object argument] →** negligible against the cgo crossing; accepted for correctness and reviewability.

## Migration Plan

1. Change the template and the two accessor methods.
2. `make generate` to regenerate bindings.
3. Run the full test suite; inspect the regenerated diff for unintended changes.
4. No consumer migration: no exported signature changes, so existing Go code that compiles today still compiles. Code that previously crashed on an object argument simply starts working.
5. Rollback is reverting the template and regenerating.

## Open Questions

- Whether the same nil-receiver hardening should be applied to the sibling accessors (`AsGDExtensionConstObjectPtr()`, `AsGDExtensionTypePtr()`) for consistency, or left to the accessors this change actually depends on. Deferrable; does not affect the spec or the task breakdown.
