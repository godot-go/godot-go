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

Go generics abstract types but not arity, so the repo's template machinery emits generic wrappers per arity (0..16) × return family × static/instance × ptrcall/varcall. Decode and encode are package-level generics parameterized by one constraint union per slot (D3); the wrappers carry no per-type logic. Example core shape (instance, value-return family):

```go
// generated:
func wrapPtrcallValue2[TImpl any, S selfTyped[TImpl], A1 nativeSlot, A2 nativeSlot, R retValueSlot](
    fn func(*TImpl, A1, A2) R,
) ptrcallFn {
    return func(inst GDClass, args []GDExtensionConstTypePtr, retPtr GDExtensionUninitializedTypePtr) {
        self := inst.(S).Self()
        out := fn(self, loadTypedSlot[A1](args[0]), loadTypedSlot[A2](args[1]))
        storeTypedRet[R](retPtr, out)
    }
}
```

The instance handoff needs one non-obvious Go-generics fact, pinned empirically: `inst.(*T)` with `T` interface-constrained is a compile-time impossible type assertion ("type *T is pointer to type parameter, not type parameter"), so the receiver is resolved through the `selfTyped` contract instead —

```go
// pkg/core/method_bind_typed.go
type selfTyped[T any] interface {
    GDClass
    Self() *T
}
```

— asserting `inst.(S)` (an interface target, always legal) and taking `Self()`'s concrete `*T`. The method is exported because a bound class lives in the user's package and a Go method with an unexported name does not satisfy an interface declared in another package. A bound class carries one one-liner, exactly alongside the `GetClassName`/`GetParentClassName` pair the API already requires per class and outside the "pure library" boundary (users write their own class code; no user-project codegen step exists):

```go
func (c *Example) Self() *Example { return c }
```

The assertion is a single itab check and `Self()` a return; S is not inferable from the handler alone, so the typed bind instantiates with the class pointer and its self-typed twin (`wrapPtrcallValue2[Example, *Example]`, how the spot test drives it) — the bind-overload shape in task 3.1 supplies both parameters to keep user call sites one-argument-deep. A misbound dynamic type fails loudly on the itab check.

Return encoding has exactly two families, chosen at bind time by the parameter type: `retValueSlot` (native-width scalars, the vector/geometry builtins, RID, and String — raw store, generic, no fan-out; membership pinned to the audited matrix below) and the copy-constructor returns (StringName, NodePath, Callable, Signal, Array, Dictionary, every Packed* — engine copy constructor through per-type methods — D3). A distinct generated wrapper family serves the copy-ret shape. Arity ceiling 16, mirroring the widest signatures the engine's own method metadata declares; above the ceiling or for any parameter type outside the constraint unions, the bind stores no closure and the reflective fallback serves it (spec: "Untypeable signatures fall back without behavior change"). The generated wrappers are plain `.gen.go` files — the same convention as all other generated surfaces in this repo.

### D3: Typed decode is a constraint-union raw load, byte-identical to the reflective decoder

Investigation against the live reflective decoder (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`) showed every engine ptrcall cell is already raw-loaded in Go today: scalar slots are direct `*(*A)(arg)` dereferences; Color, Array, Dictionary, Callable, Signal, RID, String, StringName, NodePath, and every Packed*Array load raw with pure borrow semantics ("nothing was acquired, released nowhere"); only the vector/geometry family (Vector2..Vector4i, Rect2, Transform2D, Basis, AABB, Plane, Quaternion, Transform3D, Projection) funnels through engine copy constructors that juggle call-scoped `runtime.Pinner` records — pins whose sole purpose is protecting a Go destination during a synchronous copy of POD bytes. All generated builtins are `[N]uint8` native-layout byte arrays, so a single constraint union plus a raw load compiles per-instantiation to the byte-identical operation for the whole family:

```go
// constraint unions (pkg/core/method_bind_typed.go):
type nativeSlot interface {
    bool | ~int | ~int8 | ~int16 | ~int32 | ~int64 |
        ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float64 |
        Vector2 | .. | Color | RID | Callable | Signal |
        Array | Dictionary | PackedByteArray | .. | PackedColorArray
}

func loadTypedSlot[A nativeSlot](ptr GDExtensionConstTypePtr) A {
    return *(*A)(unsafe.Pointer(ptr))
}
```

Notes pinned by this shape:

- No per-type decode methods are emitted for arguments; the interface-method shape originally sketched here would have boxed every loaded struct back into an `any` (the allocation this change removes), and scalar types cannot carry methods at all.
- Small integers (int8..int32, uint8..uint32) load byte-truncating exactly like the reflective decoder's own direct loads (both read the low bytes of the int64 cell on little-endian), so value parity is code-equality, not accidental.
- float32 is excluded: the reflective decoder double-loads and truncates, while a raw `*(*float32)` read would break parity; it stays reflective.
- The vector-family raw load deliberately drops the reflective copy-constructor's engine call and pins. Copy constructors of POD values are byte copies with no reference effects, so observable semantics are unchanged while the per-arg pinner records and engine round-trips disappear — this is part of the change's allocation goal, and the engine tier's fixtures assert value equality.
- Go `string` (UTF-8 conversion), `[]T` slices (deep copy), `Object`/`Ref` parameters (binding resolution / ref acquisition), `Variant` arguments and `Variant` returns (view/construct semantics differ from a raw load/store) all stay on the reflective fallback; they are the exact boundary of the two constraint unions.
- Refcounted *returns* split by the audited return-encoder matrix (`variant_refect_value.go`, backed by `test/pkg/variant_matrix_audit_tests.go`): String returns through `StringEncoder.EncodeTypePtrArg`, which is exactly a raw store (`*(*R)(pOut) = value`) — String therefore *joins* `retValueSlot`. StringName, NodePath, Callable, Signal, Array, Dictionary, and every Packed*Array return through their engine copy constructor plus a conditional source `Destroy()` gated on `destroySource` (echo classification decides it upstream) — those 16 types are the per-type emission family: generated value-receiver encode methods (`RetEncodeInto`, pkg/builtin/typed_ret_encoders.gen.go) calling `XCopyConstructor` with call-scoped pinning, invoked through the exported method constraint `RetCopyEncodable` (declared beside the methods in pkg/builtin — a dot-import only brings exported identifiers into scope) so no fan-out over refcounted types is needed and no raw byte store ever transfers refcount ownership (AGENTS.md's copy-constructor rule).

*Alternative considered:* decode everything generically via `unsafe` casts on `any` — rejected: every load would box into an interface. Per-type interface decode methods returning the concrete type — rejected: an interface-returning call boxes; only a pointer-receiver method-value call avoids that, and the union needs no methods at all.

### D4: Varcall defaults and variadic as first-class closures

The variadic branch compiles to `varcallFn` directly (the collected `[]Variant` is the variadic slice; no `CallSlice`). Non-variadic varcall wrappers know `declared` and `defArgsCount` from bind metadata and reproduce the `fillCallArgs` trailing-default convention exactly (defaults array maps to the *last* defArgsCount parameters), including the loud panic on under-supply from direct Go callers; the callback-side `classifyVarcallArity` pre-validation remains the gate for engine calls.

Arity wrappers for varcall resolve their per-arg Variant decoding through a bind-time cached decoder list (the existing per-type Variant conversion helpers captured as concrete function values); each decoded value crosses into the typed handler as a plain typed value by type assertion, which is unboxing, not reflection. Per-argument decode allocations remain at parity with the reflective path (one boxed value per Variant argument); this eliminates reflect frames, `[]reflect.Value` slices, `Func.Call` machinery, and the DeepEqual-heavy echo check on the hot path.

The trailing-default fill is not re-implemented in the wrappers: `fillCallArgs` and every typed varcall closure call the same `fillVarcallCallArgs(gdArgs, declared, defaultStart, defaults, name)`, so the convention and the loud under-supply panic are literally one code path.

### D5: Borrow-echo classification re-implemented on typed values, semantics bit-identical

The typed return encoder performs echo classification against the decoded arguments using per-type byte-compare helpers that reproduce what `reflect.DeepEqual` compares on the wrapper structs today (shared-storage pointer equality for COW containers, intern equality for `StringName`). This preserves every `borrow-echo-detection` scenario, including the accepted same-content-clone leak, and keeps variadic dispatch free of echo suppression. A differential fixture (same call through both APIs) guards the parity.

Resolved during implementation:

- The classification helpers are a generated factory trio in `pkg/core/ret_release.gen.go`: `newEchoCompare[R]()` returns the concrete `func(T, T) bool` boxed as `any` and the wrapper type-asserts it to `func(A, R)` at bind time (succeeding only on exact type equality — the reflective classifier's `args[i].Type() == ret.Type()` requirement, resolved without per-call reflection); `newVarcallRetDestroyer[R]()` encodes the varcall release matrix (StringName/NodePath destroy, everything else nil); `newVarCopyRetEncoder[R]()` resolves the per-type `XEncoder.EncodeVariantPtrArg` that the reflective Variant return matrix performs. The compare itself is `a == b` on the fixed-size byte arrays these builtins are — byte-for-byte the fields `reflect.DeepEqual` walks, at memequal cost with zero allocation.
- Release wiring follows the audited matrices exactly: ptrcall copy-ret wrappers hand every non-echo return to `RetDestroy` (the generated value-receiver counterpart of the matrix's `inst.Destroy()`), all 16 family members including `PackedVector4Array` — the main `borrow-echo-detection` spec's "except PackedVector4Array" clause predates PR #158, which upgraded that ptrcall return from a raw store to copy-construct plus gated destroy; the code matrix is the parity contract and the spec text is corrected alongside this change. Varcall copy-ret wrappers release only StringName/NodePath, after encoding and before the owned-argument release, matching the reflective order (encode → source release → `destroyOwnedContainerArgs`).
- Variadic copy-ret returns (`ClassDBBind{,Static}VariadicCopyRetMethod`) use the varcall destroyer unconditionally — no echo check exists, mirroring the reflective variadic branch's hard-coded `destroySource=true` (raw Variants hold no borrows).
- The receiver never joins the comparison on either path: the reflective `hasReceiver` skip of `args[0]` corresponds structurally to the typed wrappers keeping the receiver in `self`, outside the parameter slots.

### D6: Notification and property callbacks get fixed-shape closures

Fixed handler shapes, taken from the actual dispatch sites in `classdb_callback.go`, become named closure types attached at registration, removing the last generic `Func.Call` sites there:

```go
notificationFn     func(inst GDClass, what int32, reversed bool)  // GoCallback_ClassCreationInfoNotification
propGetFn          func(inst GDClass, name string) (Variant, bool) // GoCallback_ClassCreationInfoGet
propSetFn          func(inst GDClass, name string, value Variant) bool
propCanRevertFn    func(inst GDClass, name StringName) bool        // _property_can_revert
propGetRevertFn    func(inst GDClass, name StringName) (Variant, bool)
```

Property-list and to-string have no `Func.Call` sites (they already run through typed direct fields such as `ValidateProperty`) and are out of scope. Untyped handlers through name-string registration fall back.

## Risks / Trade-offs

- [Two dispatch paths share semantics but not code] → Parity is enforced by differential fixtures (same calls both APIs), `AllocsPerRun` caps pinning the allocation budget, and the existing varcall/echo/decode suites must stay green unchanged.
- [Arity/style fan-out blows up generated code] → Hard ceiling (arity 16) with reflective fallback beyond; fan-out is 2 styles × ~3 return shapes × 2 arity families — bounded, generated once.
- [Echo-compare parity depends on reproducing `DeepEqual`'s exact field semantics] → Per-type compare helpers are written against the wrapper structs' fields, with fixtures covering echo, non-echo, COW clone, and intern-equality cases before the reflect path is demoted.
- [Registration complexity grows: two APIs to support indefinitely] → Name-string API is the fallback contract and stays; typed API is additive; no engine ABI or registration-flag changes.
- [Typed coverage excludes common-but-decode-heavy signatures: Go string, []T slices, Object/Ref parameters, float32, Variant] → These binds keep the reflective path (identical semantics); a generated arity×family fan-out can extend coverage later without changing the constraint unions or the fallback behaviors.
- [cgo dispatch shape unchanged, but closure capture extends object lifetimes] → Closures are per-bind, created at registration and dropped at unregister (same lifetime as metadata); no per-call captures that could pin engine memory.

## Migration Plan

1. Generate decoder interfaces + wrapper family; nothing consumes them yet (pure addition, suite green).
2. Add typed bind overloads; wire ptrcall + virtual dispatch to prefer the closure when present, fallback otherwise. Name-string binds unaffected.
3. Varcall wrappers (defaults, variadic) + callback fixed shapes; differential parity fixtures run.
4. Quantified gates: `AllocsPerRun` caps on a representative arity matrix; `make test` green including the existing mirrored suites; test/demo manual pass in the demo suite.
5. Rollback is trivial per phase: binds not created through the typed API keep dispatching reflectively; removing the typed API call sites restores the prior state with no structural change. No repo-wide flag or ABI flip is involved.

## Open Questions

- Arity ceiling value (16 proposed) — confirm against the widest virtual surfaces actually catalogued for user classes before generation; narrowing requires only regenerating.

Resolved during implementation:

- Typed overload names are `ClassDBBind{Value,CopyRet}MethodN` for instance binds, with `Static`/`Virtual` inserted after `ClassDBBind` for the static and virtual shapes (arity `N` in the name, matching the generated wrapper families). `Self()` (not `self()`) is the `selfTyped` method — a bound class lives in the user's package and an unexported method cannot satisfy a core-declared interface across packages.
- The bind-time decoder list is a generated factory trio in `pkg/core/var_args.gen.go` (`newVarArgDecoder`, `newVarArgReleaser`, `newVarRetEncoder`): each resolves `reflect.TypeFor[T]()` once at bind time and returns the concrete closure the wrapper captures — scalar families by `Kind` (so named types decode exactly as the reflective kind switch does), array-kind builtins by exact type (the reflective decoder's inner switch). Per-type *method* families are impossible: Go forbids methods on predeclared types (`bool`, `int`, …), and a tilde member (`~int64`) cannot be type-switched. Bind-time reflection is allowed; per-call reflection is not.
- Varcall argument coverage is a strict superset of the ptrcall `nativeSlot` union (the shared bind overloads constrain one parameter list for both shapes) and additionally admits `float32` and Go `string`, which the varcall decoder serves natively. `Variant`, slices, and `Object`/`Ref`/`GDClass` parameters remain reflective-only. Copy-constructor varcall *returns* ride their own echo-aware closure family (`wrap{Static,}VarcallCopyRet{N}` plus the variadic pair, landed with task 5.2): encode through the type's Variant encoder, then release per the varcall matrix only when the return is not an echo.
- Variadic binds are varcall-only: `ClassDBBind{,Static}Variadic{Value,None}Method` bake `METHOD_FLAG_VARARG`, register no ptrcall closure (the engine varcalls vararg binds), and spread the collected `[]Variant` straight into the handler — no fill, no per-parameter decode, so no borrows are held and the always-destroy return discipline holds trivially.
