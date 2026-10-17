# Proposal

## Why

Every Go method declaring a native slice (`[]byte`, `[]string`, `[]int32`, `[]Vector2`, …) aborts the process: `log.Panic("slice not implemented")` sits in the varcall argument decoder (`pkg/core/method_bind_reflect.go:214`), the ptrcall argument decoder (`:599`), and the return encoder. Worse, registration silently mis-maps `[]T` to its *element's* variant type (`[]int32` registers as scalar `INT`, not `PACKED_INT32_ARRAY`). Packed arrays are GDScript's workhorse data types; until slices cross the boundary, Go code cannot consume or produce bulk data (asset bytes, point buffers, string lists). This is roadmap item A1.

## What Changes

- Go native slices `[]T` become first-class method signature types across both call paths, mapped to the engine's ten `Packed*K` types: `[]byte`↔`PackedByteArray`, `[]int32`↔`PackedInt32Array`, `[]int64` and `[]int`↔`PackedInt64Array`, `[]float32`/`[]float64`↔`PackedFloat32/64Array`, `[]string`↔`PackedStringArray`, `[]Vector2`/`[]Vector3`/`[]Vector4`↔`PackedVector2/3/4Array`, `[]Color`↔`PackedColorArray`.
- Deep-copy semantics: a decoded slice is Go-owned memory (GC frees it; safe to retain after the call); a returned slice builds a fresh engine container. No borrowed views, no write-back.
- Data plane uses `operator_index`/`operator_index_const` from the GDExtension interface (already resolved in `pkg/ffi`) plus generated cell methods: bulk copy for layout-identical plain elements, per-element `String` protocol for `PackedStringArray`.
- Registration maps slices to the correct `PACKED_*` variant type; element types with no packed counterpart (`[]bool`, `[]Vector2i`, `[]Object`, nested slices) are rejected at bind time with a clear error instead of being mis-registered as scalars.
- Fixtures pin round-trip fidelity for all 10 types, retention safety, cgocheck-safe direction, and empty/nil conventions; `docs/roadmap.md` A1 and the `docs/overview.md` "Packed Arrays" limitation are updated.

## Capabilities

### New Capabilities

- `native-slice-marshalling`: Go native slices ↔ engine `Packed*Array` across ptrcall and varcall — registration mapping, deep-copy data plane, per-element string handling, ownership regimes, and rejection of unmappable element types.

### Modified Capabilities

- None. The opaque `Packed*K` cell behavior, container release idioms, and return-ownership rules (`packed-vector4-array-decode`, `varcall-argument-decode`, `variant-copy-ownership`, `container-lifecycle-management`) keep their existing requirements; slice marshalling reuses them.

## Impact

- `pkg/core/variant_reflect_type.go` (Slice registration mapping), `pkg/core/method_bind_reflect.go` (:214, :599 decoders), `pkg/core/variant_refect_value.go` (return encoder), plus a new hand-written `pkg/core` slice-marshalling helper file.
- Fixtures: `test/pkg/example.go` + `test/demo/main.gd`.
- Docs: `docs/roadmap.md` (A1), `docs/overview.md`.
- No generated files change; no new FFI surface (all `operator_index` interface functions are already resolved).

## Non-goals

- Typed `Array[T]` / `Dictionary[K,V]` generics (godot-cpp `TypedArray`/`TypedDictionary` equivalents) — a follow-on change; this change's rejection errors are the seam it grows into.
- Zero-copy or borrowed-slice views across the boundary, and in-place write-back of mutated arguments.
- `[]bool` (the engine has no packed bool array) and other elements with no packed type; no silent fallback to `Array`.
- Slices of refcounted engine cells (`[]String`, `[]Array`) — GC cannot release their elements; use `[]string` or the cell type directly.
- Performance benchmarking beyond round-trip fixtures.
