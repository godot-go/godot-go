## Context

Every Godot built-in type is generated as a fixed-size byte array (`String`, `StringName`, `NodePath` are all `[8]uint8` — `pkg/builtin/builtinclasses.gen.go`). Method argument/return marshalling is dispatched **by `reflect.Kind`** at the call boundary (see `pkg/core/method_bind_reflect.go` and `pkg/core/variant_refect_value.go`), so these types land in `case reflect.Array`.

The audit in the proposal's motivation found three concrete defects for the basic types:

1. **Argument decode panics** — `convertVariantToGoTypeReflectValue` (varcall) and `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs` (ptrcall) both fall through to `log.Panic` for `String`, `StringName`, `NodePath` args (`method_bind_reflect.go:286-289`, `:598-600`). The varcall array switch handles 19 types but not these; the ptrcall switch handles only 4.
2. **Varcall narrow-number return OOB read** — `createNumberEncoder.encodeReflectVariantPtrArg` (`pkg/builtin/variant_number_encoder.go:121-128`) passes a pointer to the narrow Go value (`T`, e.g. `int8`) to Godot's INT/FLOAT `variant_from_type_constructor`, which reads 8 bytes.
3. **Go-string varcall return leak** — `createGoStringEncoder.encodeVariantPtrArg` / `encodeReflectVariantPtrArg` (`pkg/builtin/variant_string_encoder.go:76-114`) build a temporary Godot `String` but the `Destroy()` calls are commented out (`:68`, `:81`).
4. **Unsigned decode sign-wrap** — the varcall path derives every unsigned kind from signed `arg.ToInt64()` (`method_bind_reflect.go:103-137`), wrapping `uint64` > `MaxInt64`.

See proposal.md for motivation; the spec (`basic-built-in-types`) pins the required behavior.

## Goals / Non-Goals

**Goals:**
- Decode Godot `String`, `StringName`, `NodePath` method arguments on both the ptrcall and varcall paths.
- Fix the varcall narrow-numeric return OOB read and the Go-string return leak.
- Preserve `uint64` fidelity on the varcall path.
- Round-trip tests for the basic types in both call styles.

**Non-Goals:**
- Vector/engine/container built-in types (e.g. `Vector3` via ptrcall, `Packed*Array`) — separate gap, listed in the proposal's non-goals.
- `Packed*Array` → Go slice conversion.
- Introducing owned-copy semantics for decoded args (see Decision 2); existing args are all borrowed.

## Decisions

### Decision 1: Decode `String`/`StringName`/`NodePath` with the existing Variant accessors (varcall) and byte-copy helpers (ptrcall)

- **varcall path** — add cases to the `case reflect.Array:` type-switch in `convertVariantToGoTypeReflectValue` that call the already-generated `arg.ToString()`, `arg.ToStringName()`, `arg.ToNodePath()` (`variant.gen.go:485/995/1029`). This mirrors the existing `arg.ToVector2()` pattern exactly.
- **ptrcall path** — add cases to `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`:
  - `String`: `v := *(*String)(arg)` (raw byte copy, SSO-safe).
  - `StringName`: `v := NewStringNameWithGDExtensionConstStringNamePtr((GDExtensionConstStringNamePtr)(arg))` (existing helper, `char_string.go:18`).
  - `NodePath`: `v := *(*NodePath)(arg)`.
- **Alternatives considered:** using Godot copy constructors (`NewStringNameWithStringName`, `NewNodePathWithNodePath`) to produce refcounted owned copies. Rejected because no decoded argument in the codebase is owned/destroyed — every other arg type is borrowed — so owned copies would introduce an inconsistent and leak-prone contract (callers would have to `Destroy` every arg).

### Decision 2: Keep decoded args borrowed

Decoded args reference Godot-owned data for the duration of the call, matching every existing arg type and the documented `StringName` lifetime guidance. Receivers must not `Destroy` argument values. Trade-off recorded in Risks.

### Decision 3: Fix narrow-numeric varcall returns by widening through an `E` temp

In `createNumberEncoder.encodeReflectVariantPtrArg`, mirror the already-correct `encodeVariantPtrArg` (`variant_number_encoder.go:57-66`): convert the reflect value to `T`, write it into an `E` (`int64`/`float64`) temp via `encodeTypePtrArg`, and pass `&enc` to the constructor. This guarantees an 8-byte, correctly widened value for `int8/16/32`, `uint*`, and `float32`.

**Alternative considered:** type-assert directly to `E`. Rejected — `T` and `E` differ (e.g. `T=int8`, `E=int64`), so a direct cast requires the same reflect `Int()/Float()` branching the fix avoids by reusing `encodeTypePtrArg`.

### Decision 4: Restore the intermediate `String` destruction

Re-enable `defer enc.Destroy()` at `variant_string_encoder.go:68` (decode) and `:81` (encode), and add the matching `Destroy()` in `encodeReflectVariantPtrArg` (`:108-114`). `StringNewWithUtf8Chars` may heap-allocate for non-SSO strings; `Destroy()` frees that. The `defer` ordering is correct: the variant is built from `pEnc` before the deferred destroy runs. Safe for SSO strings (no-op).

### Decision 5: Use typed unsigned accessors in the varcall decode

Replace the signed `arg.ToInt64()` derivation for `reflect.Uint*` kinds with the generated `ToUint*` accessors (`variant.gen.go:64-344`), so `uint64` values above `MaxInt64` are preserved. Signed kinds keep `ToInt*`.

## Risks / Trade-offs

- [Borrowed args alias Godot-owned data] → A receiver that `Destroy()`s a `StringName`/`NodePath` arg double-frees. Mitigation: this matches existing arg semantics; document the borrowed contract and cover only the read path in tests.
- [String leak for non-SSO decode remains in the pre-existing `case reflect.String` ptrcall path] → Out of scope (that path predates this change and is unchanged); noted for a follow-up.
- [Narrow-numeric fix only touches the reflect/varcall path] → The ptrcall path already widens correctly; verified in the audit.
- [`uint64` > `MaxInt64` on ptrcall] → Already correct (`*(*uint64)(arg)`); only the varcall path changes.

## Migration Plan

No data migration. Backwards compatible: this only adds previously-panicking decode cases and fixes leaks/OOB on existing code paths. If a regression slips through, the per-path changes are isolated to `method_bind_reflect.go` (decode) and two encoder files; each is independently revertible.
