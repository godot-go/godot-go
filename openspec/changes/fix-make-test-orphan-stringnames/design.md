## Context

Three orphan StringName warnings appear at `make test` exit: "returned_name", "root", "child". See proposal.md for motivation.

The root cause is systemic: godot-go's encoding/decoding layer does raw byte copies (memcpy) when StringName/NodePath values cross the C/Go boundary. Godot's `StringName` is a refcounted interned-string type — every copy must go through the GDExtension copy constructor (`constructor_1`) to increment the internal refcount, and every destruction must go through the GDExtension destructor to decrement it. Raw byte copies bypass this entirely.

godot-cpp (reference implementation) never does raw byte copies for refcounted types. Its `PtrToArg<StringName>::encode` uses copy-assignment (`operator=`), which internally calls destructor-then-copy-constructor, producing correct refcounts.

A prior fix (`451b065` on `fix-vector-built-in-types` branch, not merged to main) addressed the return path via `StringNameCopyConstructor`/`NodePathCopyConstructor` helpers.

## Goals / Non-Goals

**Goals:**
- Every StringName/NodePath that crosses the C/Go boundary uses the GDExtension copy constructor, matching godot-cpp exactly
- Eliminate "returned_name", "root", and "child" orphan StringName warnings from `make test`
- Fix `NewStringNameWithGDExtensionConstStringNamePtr` to use copy constructor instead of raw byte copy
- Fix argument decoding from ptrcall for StringName/NodePath
- Remove or fix the leaking `NewSimpleGDExtensionPropertyInfo`

**Non-Goals:**
- Fixing Godot-internal orphans (e.g., Godot-side Image orphan) — these are out of our control
- Any user-facing API changes

## Decisions

### Pattern: copy constructor + Destroy

The core pattern matches godot-cpp's refcount semantics exactly. Compare:

```
godot-cpp:  return value  ──▶ copy-assign into r_ret ──▶ temp goes out of scope
              ptrcall          (destruct+copy-construct)  (destructor runs)
              refcount=1       refcount: 1 → 2            refcount: 2 → 1

godot-go:   return value  ──▶ copy-construct into r_ret ──▶ inst.Destroy()
              ptrcall          StringNameCopyConstructor    calls destructor
              refcount=1       refcount: 1 → 2              refcount: 2 → 1
```

Net refcount on `r_ret` is 1 in both cases. The proposed pattern is verified against godot-cpp's `PtrToArg<StringName>::encode` in `godot-cpp/include/godot_cpp/core/method_ptrcall.hpp`.

### Sites requiring the fix

```
┌──────────────────────────────────────────────────────────────────┐
│                  ALL C/Go BOUNDARY CROSSINGS                     │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Return encoding (ptrcall, Struct branch)                        │
│    variant_refect_value.go:108 — StringName                      │
│    variant_refect_value.go:110 — NodePath                        │
│                                                                  │
│  Return encoding (ptrcall, Array branch)                         │
│    variant_refect_value.go:194 — StringName                      │
│    variant_refect_value.go:196 — NodePath                        │
│                                                                  │
│  Return encoding (varcall, Variant branch)                       │
│    variant_refect_value.go:334 — StringName                      │
│    variant_refect_value.go:336 — NodePath                        │
│                                                                  │
│  Argument decode (ptrcall)                                       │
│    method_bind_reflect.go:610-612 — StringName                   │
│    (uses NewStringNameWithGDExtensionConstStringNamePtr)         │
│                                                                  │
│  NewStringNameWithGDExtensionConstStringNamePtr                  │
│    char_string.go:18-26 — raw byte copy, no refcount incr       │
│                                                                  │
│  NewSimpleGDExtensionPropertyInfo                                │
│    object.go:18-40 — DEPRECATED, StringNames never destroyed    │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### Return encoding (Struct + Array branches)

Replace `StringNameEncoder.EncodeTypePtrArg(inst, rOut)` with:
1. `StringNameCopyConstructor(rOut, inst.NativeConstPtr())` — calls Godot's copy constructor, refcount++
2. `inst.Destroy()` — calls Godot's destructor, refcount--

Same for NodePath with `NodePathCopyConstructor`.

### Return variant encoding (Variant branch)

The `encodeVariantPtrArg` function already properly constructs a Variant via `GDExtensionVariantFromTypeConstructorFunc`, which creates a new refcounted copy. The fix needed is simply `inst.Destroy()` after encoding, so the original ref is released.

### Argument decode from ptrcall

`NewStringNameWithGDExtensionConstStringNamePtr` currently does a raw byte copy from a Godot-owned `GDExtensionConstTypePtr`. Replace with `NewStringNameWithStringName()` (the GDExtension copy constructor) to create a properly refcounted independent copy. This is safe because Godot's ptrcall argument buffer is guaranteed to live for the duration of the call.

### NewSimpleGDExtensionPropertyInfo (DEPRECATED)

This function creates `StringName` and `String` objects internally, extracts their native pointers into the `GDExtensionPropertyInfo`, pins the struct, but never calls `Destroy()`. Remove the function entirely or replace its callers with `NewGDExtensionPropertyInfoFromNames`.

### Helper functions

Add to `pkg/builtin/char_string.go`:

```go
func StringNameCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
    pnr.Pin(unsafe.Pointer(src))
    CallBuiltinConstructor(globalStringNameMethodBindings.constructor_1, out, src)
}

func NodePathCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
    pnr.Pin(unsafe.Pointer(src))
    CallBuiltinConstructor(globalNodePathMethodBindings.constructor_1, out, src)
}
```

### godot-cpp cross-reference mandate

AGENTS.md must be updated to require checking godot-cpp at `../godot-cpp` before implementing any GDExtension boundary crossing — this ensures refcounted types are always handled via the GDExtension constructor/destructor API, never by raw byte copy.

## Risks / Trade-offs

- **CallBuiltinConstructor availability:** `globalStringNameMethodBindings.constructor_1` and `globalNodePathMethodBindings.constructor_1` must exist. They were present in the prior fix (`451b065`).
- **Godot side effect:** If a future Godot version changes copy-constructor semantics, this approach may need revisiting.
- **Encoder generality:** The current `variant_builtinclass_encoder.go` does raw byte copy for ALL builtin types (Vector2, Color, etc.). For non-refcounted types this is correct. We're special-casing StringName/NodePath rather than rewriting the encoder — that's the right tradeoff for now.
