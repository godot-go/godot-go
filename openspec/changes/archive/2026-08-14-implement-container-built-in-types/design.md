## Context

See proposal.md. Every Godot built-in type is a fixed-size byte array (e.g. `Array` is `[8]uint8` — `pkg/builtin/builtinclasses.gen.go`). Method argument decoding is dispatched by `reflect.Kind` at the call boundary (see `pkg/core/method_bind_reflect.go`). The audit reveals that the Go→Godot *encoding* paths (`GDExtensionTypePtrFromReflectValue` and `GDExtensionVariantPtrFromReflectValue` in `pkg/core/variant_refect_value.go`) already handle all container types; the gap is exclusively in the two Godot→Go *decode* switches:

1. **Varcall decode** (`convertVariantToGoTypeReflectValue`, `method_bind_reflect.go:225-336`) — the `reflect.Array` case handles all nine `Packed*Array` types and `Dictionary`/`Signal`/`Callable`, but **`Array` is absent**. The Variant type has a generated `ToArray()` accessor (`variant.gen.go:1199`).
2. **Ptrcall decode** (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`, `method_bind_reflect.go:606-785`) — the `reflect.Array` case handles vector types, `Variant`, `PackedInt64Array`, `String`, `StringName`, `NodePath`, but is **missing `Array` and eight of nine `Packed*Array` types**. Only `PackedInt64Array` is present (line 763-773), using a byte-copy constructor pattern (`NewPackedInt64ArrayWithPackedInt64Array`).

Each container type has a generated `New<Type>With<Type>(from)` copy constructor (e.g. `NewPackedByteArrayWithPackedByteArray`, `NewArrayWithArray`) that allocates an owned Godot value from a source pointer, matching the existing `NewPackedInt64ArrayWithPackedInt64Array` pattern at `method_bind_reflect.go:771`.

## Goals / Non-Goals

**Goals:**
- Decode all container built-in types on both the varcall and ptrcall argument paths without panicking.
- Round-trip tests exercising every container type as both an argument and a return value in both call styles.
- Update `docs/overview.md` to reflect completed container support.

**Non-Goals:**
- Go-native slice conversion (`[]Variant`, `[]byte`, `[]int32`, etc.) as automatic argument/return marshalling — the boundary uses Godot built-in container types, consistent with `Vector2`/`Color`/etc.
- Changes to the Go→Godot encoding paths, which already handle all container types.
- Adding new container types beyond those listed in `docs/overview.md`.

## Decisions

### Decision 1: Varcall decode for `Array` uses `arg.ToArray()`, matching existing Packed patterns

The `reflect.Array` case in `convertVariantToGoTypeReflectValue` already calls `arg.ToPackedByteArray()`, `arg.ToPackedInt32Array()`, etc. for the Packed types. Adding `case Array: v := arg.ToArray(); return reflect.ValueOf(v), nil` follows the identical pattern, using the generated `Array.ToArray()` accessor.

**Alternatives considered:** Using `Rect2iEncoder.DecodeVariantPtr(arg.NativeConstPtr())` style (as `Rect2`/`Rect2i` do at lines 251/254). Rejected — the Packed types already use the `arg.To*()` accessor directly, so `Array` should match for consistency.

### Decision 2: Ptrcall decode uses copy-constructor pattern, matching `PackedInt64Array`

The ptrcall path receives a `GDExtensionConstTypePtr` (`arg`) that points to a borrowed Godot buffer for the duration of the call. For `PackedInt64Array` (the only existing case), the code dereferences the pointer and passes the value to a copy constructor (`NewPackedInt64ArrayWithPackedInt64Array`). The same pattern applies to every container type:

```
case PackedByteArray:
    pV := (*PackedByteArray)(unsafe.Pointer(arg))
    if pV == nil { log.Panic(...) }
    v := NewPackedByteArrayWithPackedByteArray(*pV)
    args[i+1] = reflect.ValueOf(v)
```

and for `Array`:

```
case Array:
    pV := (*Array)(unsafe.Pointer(arg))
    if pV == nil { log.Panic(...) }
    v := NewArrayWithArray(*pV)
    args[i+1] = reflect.ValueOf(v)
```

**Alternatives considered:** Raw byte-copy (`v := *(*Array)(arg)`), as used for `String`/`StringName`/`NodePath`. Rejected — container types hold refcounted internal data (arrays, strings, etc.) that require a proper copy constructor to avoid aliasing the borrowed ptrcall buffer. The existing `PackedInt64Array` case uses the copy-constructor pattern, so container types follow that precedent rather than the raw-copy pattern used for scalar-sized fixed structs.

### Decision 3: Decoded container arguments are owned copies

Unlike `String`/`StringName`/`NodePath` (which are borrowed byte-copies of SSO-safe `[8]uint8` buffers), container types use Godot-owned internal data via copy constructors. The copy-constructor pattern produces an owned value the Go receiver can safely use for the duration of the call. The Go receiver does not need to (and should not) call `Destroy()` on these argument values — matching the convention for all other ptrcall decoded args (vectors, Variant, etc. are all owned copies, never destroyed by the caller).

**Risk:** If the receiver stores the value beyond the call scope, the borrowed ptrcall source buffer may be invalidated. Mitigation: documented via tests and the existing convention — all ptrcall decoded args follow this pattern.

## Risks / Trade-offs

- [Ptrcall buffer aliasing for container types] → The ptrcall `arg` pointer is borrowed for the call duration. Copy constructors (`New*With*`) produce owned values, matching the `PackedInt64Array` precedent. The Go receiver must not retain or `Destroy()` argument values.
- [Varcall `Array` decode produces an owned value] → `ToArray()` returns an owned Godot `Array` (via the type-from-variant constructor). The Go receiver gets a value that must outlive the call — consistent with all other varcall decoded args (vectors, Packed types) which are also owned.
- [Test coverage breadth] → Each container type needs a round-trip test in both call styles, which is ~18 test cases. Mitigation: table-driven GDScript assertions + Go methods using the existing `Echo` pattern.

## Migration Plan

No data migration. Backwards compatible: purely additive decode cases that convert panics to successful decode. If a regression slips through, the changes are isolated to the two decode switches in `pkg/core/method_bind_reflect.go`; each case is independently revertible.
