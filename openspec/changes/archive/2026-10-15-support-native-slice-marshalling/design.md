# Design

## Context

The opaque `Packed*K` cells already cross the boundary on all three wire points (varcall decode `method_bind_reflect.go:374-390`, ptrcall decode `:885-902`, return encode `variant_refect_value.go:155-180`); only Go-native `[]T` panics. Registration maps slices to their **element's scalar** variant type (`variant_reflect_type.go` Slice case) — latent, since every slice call died in the decoders anyway. Everything needed for the data plane exists today:

- `gdextension_interface.h` exposes mutable and const `operator_index` for all 10 packed types; `pkg/ffi/*.gen.go` already resolves them (`CallFunc_GDExtensionInterfacePacked…OperatorIndex[Const]`).
- Generated cells (`pkg/builtin/builtinclasses.gen.go`) provide `Size()`, `Append()`, `Destroy()`, `NativePtr()/NativeConstPtr()`, and `NewPacked*K()` / `NewPackedStringArrayWithArray(Array)` constructors.
- Cell types are native-layout byte mirrors (`Vector2 [8]uint8`, `Color [16]uint8`, …), so engine element buffers are byte-compatible with Go slice element storage.
- A2's ratified lessons apply, especially design **D5a** in the archived `ratify-string-marshalling`: the suite runs `GODEBUG=cgocheck=1`; the safe direction is Go code writing into engine-owned memory, never handing the engine a pointer to Go memory it could retain.

See `proposal.md → Why` for motivation and the delta spec for required behavior.

## Goals / Non-Goals

**Goals:**
- One marshalling mechanism covering all 10 packed types with a plain-data fast path and a string per-element path.
- Slice ownership simpler than A2's string regime: decoded slices need no engine release call at all.
- Rejection of unmappable element types at bind time, with an error a human can act on.

**Non-Goals (design-level):**
- No element conversion beyond the fixed table (no `uint32`→`int64` widening, no `int64`→`int32` narrowing).
- No container-level sharing between Go and the engine (no CoW handle exposed to Go).
- No changes to generated code, FFI surface, or the existing cell marshalling.

## Decisions

### D1 — Deep copy in both directions; no views
Engine packed arrays are refcounted CoW buffers; Go slices are GC-owned. A borrowed view is unsound in both directions (a GDScript write reallocs the buffer under a Go reader; the engine could retain Go memory past a call). So decode allocates fresh Go memory and copies; encode builds a fresh engine container and copies. Rejected alternative: ptrcall decode aliasing engine memory — saves one copy but turns every call-scoped borrow into a dangling-pointer trap, and the A2 borrow regime showed how expensive pinning such semantics is. Consequence: decoded `[]T` is Go-owned, GC-released, **no `Destroy()` ceremony**; the ownership table gains a trivial row and the retention fixture proves it under `cgocheck=1`.

### D2 — Closed element-mapping table, exact matches only

| Go element | Packed type | Variant type |
|---|---|---|
| `byte` (`uint8`) | PackedByteArray | PACKED_BYTE_ARRAY |
| `int32` | PackedInt32Array | PACKED_INT32_ARRAY |
| `int`, `int64` | PackedInt64Array | PACKED_INT64_ARRAY |
| `float32` | PackedFloat32Array | PACKED_FLOAT32_ARRAY |
| `float64` | PackedFloat64Array | PACKED_FLOAT64_ARRAY |
| `string` | PackedStringArray | PACKED_STRING_ARRAY |
| `Vector2` / `Vector3` / `Vector4` | PackedVector2/3/4Array | PACKED_VECTOR2/3/4_ARRAY |
| `Color` | PackedColorArray | PACKED_COLOR_ARRAY |

`[]int` follows the repo's scalar `int→int64` rule. **Exception:** a trailing `[]Variant` slot keeps the predating variadic convention and registers as the scalar VARIANT type; the variadic call machinery assembles that slice itself and never enters the packed-array decoders. Everything else — `bool`, `uint16+`, `Vector2i`/`Vector3i`, objects, nested slices, slices of refcounted cells (`[]String`) — fails at bind time. Rejected alternatives: silent widening to a nearby packed type (surprise marshalling) and falling back to `Array` (loses the packed guarantee; the future typed-generics change owns that). Rejected-slice tests run against the mapping helper at unit level, because a bind-time panic inside `Example` binds would abort the whole suite (the A2 lesson about naive rejections silently killing the rest of a GDScript driver).

### D3 — Data plane: bulk copy for plain data, per-element `String` protocol for strings
- **Decode plain** (`byte`/`int32`/`int64`/floats/vectors/`Color`): `Size()` → `operator_index_const(cell, 0)` → `unsafe.Slice` → `copy` into a fresh Go slice. Guard `len == 0` before touching the pointer (engine may hand back null for empty buffers).
- **Encode plain**: `NewPacked*K()` → `resize(n)` → `operator_index(cell, 0)` → write engine memory from Go. This is exactly D5a's safe direction: the engine never receives a Go pointer; Go code stores through a C pointer.
- **Decode strings**: per element, `operator_index_const(cell, i)` borrows a native `String`; `ToUtf8()` to Go; nothing to release (the borrow dies with the call).
- **Encode strings**: `Append(value String)` per element on a fresh cell. Rejected: `resize` + `operator_index(i)` slot-write — post-`resize` slots are already constructed and there is no operator-assign binding in the FFI; rejected: the `Array` bridge (`NewArray` + `Append(Variant)` + `NewPackedStringArrayWithArray`) — works but churns 2× the allocations and a Variant per element.

### D4 — Ownership at the three wire points reuses ratified idioms
- **varcall decode**: `variant.ToPacked*K()` gives an owned CoW copy → copy elements out → `Destroy()` the temp cell, including on decode-failure paths (the `varcall-argument-decode` release discipline).
- **ptrcall decode**: the argument cell is a call-scoped borrow → read elements directly through `operator_index_const`; no container copy, no release.
- **return**: build the cell per D3 → generated `Packed*KCopyConstructor(rOut, tmp.NativeConstPtr())` (the existing `variant_refect_value.go` pattern) → `Destroy()` the temp, matching `variant-copy-ownership`.

### D5 — Rewrite the Slice case of `ReflectTypeToGDExtensionVariantType`
Map the slice's **element** type to the packed variant type (replacing today's element-scalar mis-mapping) for parameters and returns alike; unmappable elements `log.Panic` naming the element type at bind time. This is where D2's table lives; decoders/encoders consume the same table so registration and marshalling cannot drift.

### D6 — Nil conventions
Decode: an empty engine array → zero-length slice. A Nil argument is **not convertible to a packed parameter** (corrected during implementation after fixture evidence: "Cannot convert argument 2 from Nil to PackedByteArray"). Note where this rejection lives: the engine forwards raw Variants to extension methodbinds unvalidated (`GDExtensionMethodBind::call`, core/extension/gdextension.cpp), so it is **our** varcall strict gate (`varcallArgumentConvertible`, A2) that refuses before the Go body runs, and the engine merely renders the resulting CallError. This matches the reference bindings: godot-cpp checks the same `variant_can_convert_strict` predicate in `VariantCasterAndValidate` (binder_common.hpp; debug builds only, and its body still executes with converted args), and gdext short-circuits `from_varcall_args` before invoking the Rust body (param_tuple/impls.rs) — gdext is the tighter of the two, which we follow. All set `GDEXTENSION_CALL_ERROR_INVALID_ARGUMENT` + argument index + expected type. So null is rejected on either path before any Go decode, pinned contained via `expect_rejected_call` — the same strict-conversion stance A2 ratified. Encode: nil or empty Go slice → empty `Packed*K` (never `null`, so GDScript static typing on the return stays satisfiable).

## Risks / Trade-offs

- `operator_index(0)` on an empty array may return null → length-0 guard on every pointer acquisition (decode and encode).
- The engine's call-layer Variant conversion may silently coerce one packed type into another (e.g. bytes→int32) before the binder sees it → fidelity fixtures assert exact element outcomes and the spec pins whatever the engine actually does; no Go-side assumption about conversions.
- Two copies per string element on encode (Go→native `String`→container) → accepted for correctness; the plain-data bulk path covers the hot `[]byte` workload, and benchmarking is a non-goal.
- Vector4/Color cell sizes are trusted from `extension_api.json` → round-trip fixtures with extreme values cover any layout surprise.
- Bind-time rejection panics are invisible to the GDScript suite → rejection is tested as Go unit tests on the mapping table, not as bound methods.

## Migration Plan

Library-internal; no data or API migration. Rollback is a clean revert — nothing else consumes the slice paths.

## Open Questions

- Whether the generated `PackedStringArray.Append(String)` wrapper has the exact expected signature — if not, fall back to the `Array` bridge from D3. Either choice leaves the spec, approach, and task breakdown unchanged.
