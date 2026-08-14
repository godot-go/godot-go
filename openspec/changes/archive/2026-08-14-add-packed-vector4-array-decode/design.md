## Context

See proposal.md. `PackedVector4Array` is the only container built-in type where the Go→Godot encoding path works but the Godot→Go decode path has a gap. The root cause is twofold:

1. **Missing codegen entry**: `GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR4_ARRAY` is absent from the `encoderTypeMap` in `cmd/generate/builtin/templatefunctions.go`. Every other Packed type (`PackedVector2Array`, `PackedVector3Array`, etc.) has an entry, which causes `variant.go.tmpl` to generate the `ToPackedVector4Array()` and `NewVariantPackedVector4Array()` methods. Without it, `variant.gen.go` lacks these methods.

2. **Missing decode cases**: The varcall decode switch (`convertVariantToGoTypeReflectValue`, `method_bind_reflect.go:225-336`) has no `PackedVector4Array` case (would hit the `default: log.Panic("unsupported array type")`). The ptrcall decode switch (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`, `method_bind_reflect.go:606-785`) has no `PackedVector4Array` case (would hit `default: log.Panic(...)`).

The `PackedVector4ArrayEncoder` already exists in `builtinclasses.gen.go` and `NewPackedVector4ArrayWithPackedVector4Array` exists — so the encoding paths and the ptrcall copy constructor are already available. Only the codegen entry and decode cases need to be added.

## Goals / Non-Goals

**Goals:**
- Add `GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR4_ARRAY` to `encoderTypeMap` so `make generate` produces `ToPackedVector4Array()` and `NewVariantPackedVector4Array()`.
- Add `PackedVector4Array` cases to both varcall and ptrcall decode switches in `pkg/core/method_bind_reflect.go`.
- Round-trip tests covering ptrcall and varcall for `PackedVector4Array`.

**Non-Goals:**
- Changing the encoding path (Go→Godot), which already handles `PackedVector4Array` via `PackedVector4ArrayEncoder`.
- Adding other missing types (e.g. Go-native `[]Vector4` slice marshalling).
- Regenerating unrelated generated files — only `variant.gen.go` is expected to change.

## Decisions

### Decision 1: Fix the codegen by adding the map entry, not by hand-editing variant.gen.go

The AGENTS.md guardrail states: "Do not directly modify any files that have `.gen.` in the filename." The correct fix is to add `GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR4_ARRAY` to `encoderTypeMap` in `cmd/generate/builtin/templatefunctions.go`, mirroring the existing `PackedVector3Array` entry:

```go
"GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR4_ARRAY": {
    IsReference: true,
    Encodings: []encoding{
        {"PackedVector4Array", "PackedVector4Array", "PackedVector4Array"},
    },
},
```

Then run `make generate` to regenerate `variant.gen.go`. This produces both `ToPackedVector4Array()` (used by the varcall decode) and `NewVariantPackedVector4Array()` (used for debug logging in ptrcall decode).

**Alternative considered**: Hand-editing `variant.gen.go` to add `ToPackedVector4Array()`. Rejected — violates the codegen guardrail and would be overwritten on next regeneration.

### Decision 2: Varcall decode uses `arg.ToPackedVector4Array()`

The varcall decode switch already calls `arg.ToPackedVector3Array()` for `PackedVector3Array`. Adding `case PackedVector4Array: v := arg.ToPackedVector4Array(); return reflect.ValueOf(v), nil` follows the identical pattern.

### Decision 3: Ptrcall decode uses `NewPackedVector4ArrayWithPackedVector4Array(*pV)`

The ptrcall decode switch receives a `GDExtensionConstTypePtr` pointing to a borrowed buffer. Container types hold refcounted internal data, so a copy constructor is needed (matching the existing `PackedVector3Array`/`PackedInt64Array` pattern). `NewPackedVector4ArrayWithPackedVector4Array` already exists in `builtinclasses.gen.go` (`builtinclasses.gen.go:30539`).

## Risks / Trade-offs

- [Regenerating variant.gen.go may touch other methods] → `make generate` regenerates the entire `variant.gen.go`. If the extension_api.json has been updated since the last generation, unrelated methods may also change. Mitigation: verify the diff only adds the `PackedVector4Array`-related methods. If other methods change, review whether they are expected.
- [Ptrcall buffer aliasing] → Same as all container types; the copy constructor produces an owned value. The Go receiver must not `Destroy()` argument values.
- [GDScript PackedVector4Array construction] → GDScript supports `PackedVector4Array([Vector4(1,2,3,4), ...])` directly, so test assertions are straightforward.

## Migration Plan

No data migration. Backwards compatible: purely additive decode cases and a codegen entry that was previously missing. The `PackedVector4Array` type already exists and encodes correctly; this change only enables decoding it from Godot→Go. If a regression slips through, the codegen change is isolated to `templatefunctions.go` + regenerated `variant.gen.go`, and the decode changes are isolated to the two switches in `method_bind_reflect.go`.
