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

### Nil safety requires a shared reflect-based helper, not receiver hardening

Receiver hardening alone **cannot** cover the types that matter, and this was discovered by test, not assumed. Generated Impl types embed `WrappedImpl` by value (`NodeImpl` → `ObjectImpl` → `WrappedImpl`). Calling a promoted method on a typed nil faults while Go computes `&p.WrappedImpl`, before the hardened method body runs. Measured against every call form:

| Call form on a typed-nil generated type | Result |
|---|---|
| `p.AsGDExtensionObjectPtr()` direct | faults |
| `if w != nil { w.AsGDExtensionObjectPtr() }` | faults — typed nil satisfies `!= nil` |
| reflect-based helper | returns null correctly |

So the template's nil-interface check is necessary but not sufficient.

**Adopted:** one non-generated helper, `ObjectArgPtr(Wrapped) GDExtensionObjectPtr` in `pkg/builtin`, which returns null for a nil interface, uses `reflect.ValueOf(...).IsNil()` to catch a typed nil, and otherwise delegates to the hardened accessor. The template calls this helper at every object-argument site.

Measured cost on this machine: **6.694 ns/op** for the helper versus **0.466 ns/op** for the plain accessor — **+6.2 ns per object argument**. Against a cgo boundary crossing measured at roughly 50–110 µs, that is about **0.1%**, and it buys the spec's no-panic guarantee for every shape a caller can produce.

Receiver hardening is retained as the inner layer: it makes a nil `*WrappedImpl`, a nil `Owner`, and a nil `WrappedClassInstance.Instance` return null rather than panicking, and the helper delegates into it for live objects.

**Rejected — reflection inlined into 794 wrappers:** the original rejection still stands for that form. A single shared helper is the opposite: one place to read, test, and cost-account.

**Rejected — `recover()` in the helper:** swallows every panic inside the helper rather than distinguishing a nil reference from a genuine bug, and is not measurably cheaper than the reflect check.

### `Ref` arguments need `IsValid()` before `ToObject()`

`RefBase.ToObject()` returns `r.m_ref` directly, so it dereferences and panics on a nil `*RefBase`. The generated `Ref`-argument path must call `IsValid()` first and emit a null slot when it reports false. `IsValid()` is nil-receiver-safe by construction (`r != nil && r.m_ref != zero`) and is declared directly on `*RefBase[T]` rather than promoted, so it survives a nil receiver.

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
