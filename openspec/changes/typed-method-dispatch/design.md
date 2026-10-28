# Design

## Context

See proposal.md "Why". Current mechanics: every bind stores `Func reflect.Value`; `Ptrcall`, the virtual callback path (`classdb_callback.go`), and varcall (`CallWithError`) bridge the fixed engine trampoline to arbitrary Go signatures via `reflectFuncCallArgsFrom*` (per-arg reflect values, per-call boxing), `Func.Call`/`CallSlice`, reflect-based return enumeration, and a `reflect.DeepEqual` borrow-echo check (see `borrow-echo-detection`). Registration resolves handlers by name-string through `t.MethodByName` / qualified-virtual chain walk (`classdb.go`), then validates shape via reflect.Types. Go→engine generated bindings (`classes.gen.go`) are already reflection-free and are untouched here.

Constraint that shapes everything: cgo forbids converting Go func *values* into C function pointers. All engine→Go entry already funnels through exported `GoCallback_*` symbols plus a metadata lookup, so replacing reflection means storing a **compiled closure on the bind metadata**, not inventing new C boundaries. That per-method closure must be built without runtime signature discovery — i.e., from type information the compiler itself verified.

## Goals / Non-Goals

**Goals:**

- Zero per-call reflection on ptrcall, virtual/callback, and varcall dispatch when the bind was created through the typed API.
- A typed registration API (`method`-expression based) whose signature mismatches are compile errors; the name-string API preserved as reflective fallback with unchanged semantics.
- Exact behavioral parity with the reflective paths for decode, refcount/container lifecycle, borrow-echo outcomes, default-argument fill, variadic raw-Variant handling, and error reporting; existing spec capabilities pin these and are not modified.
- Quantified, deterministic proof of the allocation reduction (per-call allocation caps asserted with `testing.AllocsPerRun`).

**Non-Goals:** (design-level; proposal lists the product-level ones)

- Pinning-churn reduction in the generated Go→engine shims; per-call `argPtrSlice` weaving in varcall; eager log-field fixes — separate changes.
- Removing the `reflect.Value` field from bind metadata (fallback + registry introspection keep it).
- Speeding the reflective fallback path; it exists for exotics and must simply stay correct.

## Decisions

### D1: Typed closures stored on bind metadata; no user-project codegen

Each bind gains a `dispatch` closure able to serve its call style, selected **at bind time**:

- `ptrcallFn  func(inst GDClass, args []GDExtensionConstTypePtr, retTypePtr unsafe.Pointer)`
- `varcallFn  func(inst GDClass, args []Variant) (Variant, error)` — carries the existing decode-failure reporting seam in its signature

The pointed-to `args []GDExtensionConstTypePtr` slice and pointer-arithmetic reinterpretation live inside the closure, compiled once.

*Alternatives considered:* (a) godot-cpp/gdext-style per-method `//export` trampolines generated over user code — rejected: forces a codegen step into every user project and multiplies cgo exports; the single-dispatcher shape already in place makes them unnecessary. (b) Keep `reflect` but erase its allocations (`runtime.MakeFuncContext`/frame pooling) — rejected: there is no supported way to call an abstract `func` value without building frames; type erasure is the cost we are removing.

### D2: Arity family via existing `.tmpl` generation, not hand-written maps

Go generics abstract types but not arity, so the repo's template machinery emits generic wrappers per arity (0..16) × return style (`None`, `Value`, `ValueAndBool`) × static/instance × ptrcall/varcall. Example core shape:

```go
// generated:
func wrapPtrcall2[T GDClass, A1 argDecode, A2 argDecode, R returnEncode](
    fn func(*T, A1, A2) R,
) ptrcallFn {
    return func(inst GDClass, args []GDExtensionConstTypePtr, ret unsafe.Pointer) {
        self := inst.(*T)
        out := fn(self, decodeArg[A1](args[0]), decodeArg[A2](args[1]))
        encodeReturn(out, ret)
    }
}
```

Arity ceiling 16, mirroring the widest signatures the engine's own method metadata declares; above the ceiling or for any parameter type lacking a decoder implementation, the bind silently stores no closure and the reflective fallback serves it (spec: "Untypeable signatures fall back without behavior change"). The generated wrappers are plain `.gen.go` files — the same convention as all other generated surfaces in this repo.

### D3: Typed decode through per-type interfaces implemented by generated code

`argDecode` / `returnEncode` / and a varcall-only Variant decode path are interfaces the generated builtin types (`pkg/builtin`) implement — native-layout loads for vector/POD builtins; copy-constructor construction for returned refcounted types; borrow wrapper without copy for refcounted *arguments*; object args via the existing object-argument decode semantics.

Semantics are copied from the existing decode functions parameter-for-parameter; the interfaces are the carrier, not a redesign. Sensitive corner preserved as-is: the class shape requirement (first embedded field = parent `Impl`) means `inst.(*T)` receiver access and `GDClass`-typed parameters both work through the existing object decode.

*Alternative considered:* decode everything generically via `unsafe` casts on `any` — rejected: it papers over the copy-construct vs borrow distinction for refcounted types and would reintroduce boxing for objects; the per-type interface is where that chemistry is already correct.

### D4: Varcall defaults and variadic as first-class closures

The variadic branch compiles to `varcallFn` directly (the collected `[]Variant` is the variadic slice; no `CallSlice`). Non-variadic varcall wrappers know `declared` and `defArgsCount` from bind metadata and reproduce the `fillCallArgs` trailing-default convention exactly (defaults array maps to the *last* defArgsCount parameters), including the loud panic on under-supply from direct Go callers; the callback-side `classifyVarcallArity` pre-validation remains the gate for engine calls.

### D5: Borrow-echo classification re-implemented on typed values, semantics bit-identical

The typed return encoder performs echo classification against the decoded arguments using per-type byte-compare helpers that reproduce what `reflect.DeepEqual` compares on the wrapper structs today (shared-storage pointer equality for COW containers, intern equality for `StringName`). This preserves every `borrow-echo-detection` scenario, including the accepted same-content-clone leak, and keeps variadic dispatch free of echo suppression. A differential fixture (same call through both APIs) guards the parity.

### D6: Notification and property callbacks get fixed-shape closures

Notification, property-list, property-get/set, and revert dispatch are per-class shapes already catalogued by the virtual-surface registration; they become named closure types (e.g. `func(GDClass, int32)` for notifications) attached at registration, removing the last generic `Func.Call` sites in `classdb_callback.go`. Untyped handlers through name-string registration fall back.

## Risks / Trade-offs

- [Two dispatch paths share semantics but not code] → Parity is enforced by differential fixtures (same calls both APIs), `AllocsPerRun` caps pinning the allocation budget, and the existing varcall/echo/decode suites must stay green unchanged.
- [Arity/style fan-out blows up generated code] → Hard ceiling (arity 16) with reflective fallback beyond; fan-out is 2 styles × ~3 return shapes × 2 arity families — bounded, generated once.
- [Echo-compare parity depends on reproducing `DeepEqual`'s exact field semantics] → Per-type compare helpers are written against the wrapper structs' fields, with fixtures covering echo, non-echo, COW clone, and intern-equality cases before the reflect path is demoted.
- [Registration complexity grows: two APIs to support indefinitely] → Name-string API is the fallback contract and stays; typed API is additive; no engine ABI or registration-flag changes.
- [cgo dispatch shape unchanged, but closure capture extends object lifetimes] → Closures are per-bind, created at registration and dropped at unregister (same lifetime as metadata); no per-call captures that could pin engine memory.

## Migration Plan

1. Generate decoder interfaces + wrapper family; nothing consumes them yet (pure addition, suite green).
2. Add typed bind overloads; wire ptrcall + virtual dispatch to prefer the closure when present, fallback otherwise. Name-string binds unaffected.
3. Varcall wrappers (defaults, variadic) + callback fixed shapes; differential parity fixtures run.
4. Quantified gates: `AllocsPerRun` caps on a representative arity matrix; `make test` green including the existing mirrored suites; test/demo manual pass in the demo suite.
5. Rollback is trivial per phase: binds not created through the typed API keep dispatching reflectively; removing the typed API call sites restores the prior state with no structural change. No repo-wide flag or ABI flip is involved.

## Open Questions

- Arity ceiling value (16 proposed) — confirm against the widest virtual surfaces actually catalogued for user classes before generation; narrowing requires only regenerating.
- Final names for the typed overloads (e.g. `ClassDBBindTypeCheckedMethod`) — cosmetic, decided during implementation.
