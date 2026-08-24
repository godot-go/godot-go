## Why

The container built-in types (`Array`, `Dictionary`, and the nine `Packed*Array` types) are decoded Godot→Go as *owned* copies: each ptrcall decode calls a copy constructor (`NewArrayWithArray(*pV)`) and each varcall decode calls a type-from-variant constructor (`arg.ToArray()`), both of which acquire a refcount. Neither `Ptrcall` nor `Varcall` ever releases these refcounts, so every call into a Go method that receives a container argument leaks one reference per container arg (flagged in code review of the container decode work). The established lifecycle convention for refcounted built-in args — set by `stringname-lifecycle-management` — is *borrow semantics* for argument decoding plus *copy-constructor* ownership transfer for ptrcall returns; containers currently deviate from it.

## What Changes

- Change ptrcall container argument decoding in `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs` from owned copy-constructor (`NewXWithX(*pV)`) to byte-copy borrow (`v := *(*X)(arg)`), matching the existing `String`/`StringName`/`NodePath` ptrcall pattern, for `Array` and the nine `Packed*Array` types.
- Change ptrcall container return encoding in `GDExtensionTypePtrFromReflectValue` from shallow byte-copy (`XEncoder.EncodeTypePtrArg`) to GDExtension copy-constructor + destroy, matching the existing `StringName`/`NodePath` return pattern, so echoing a borrowed container argument still hands Godot an owned value.
- Change varcall container argument decoding in `convertVariantToGoTypeReflectValue` from owned (`arg.ToX()` returned directly) to borrow (construct, byte-copy, destroy), matching the existing `StringName`/`NodePath` varcall pattern, for `Array`, `Dictionary`, and the nine `Packed*Array` types.
- Add echo and arg-consume (non-echo) round-trip tests for container types in both call styles to prove correct values and no leaked references.

## Capabilities

### New Capabilities

- `container-lifecycle-management`: Refcount ownership rules for container built-in types crossing the Godot↔Go boundary — borrow semantics for argument decoding and copy-constructor ownership transfer for ptrcall returns.

### Modified Capabilities

None. (The container decode/encode paths themselves are introduced by the active `implement-container-built-in-types` change; this change governs their ownership semantics, paralleling how `stringname-lifecycle-management` relates to the basic types.)

## Impact

- `pkg/core/method_bind_reflect.go` — ptrcall decode (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`) and varcall decode (`convertVariantToGoTypeReflectValue`) container cases.
- `pkg/core/variant_refect_value.go` — ptrcall return encoding (`GDExtensionTypePtrFromReflectValue`) container cases.
- `pkg/builtin/` — new container copy-constructor helpers (following `StringNameCopyConstructor` in `char_string.go`).
- `test/pkg/example.go`, `test/demo/main.gd` — echo + arg-consume container tests in both call styles.
- Behavior/contract change: a Go method receives container arguments as borrows valid only for the duration of the call; storing one past the call requires an explicit copy. Mutating an `Array`/`Dictionary` argument is visible to the caller (shared-reference types); borrowed packed arrays are read-only, since mutating one may free or reallocate the caller's buffer (see the `container-lifecycle-management` spec). This matches the existing StringName/NodePath contract. The container decode paths are new and unreleased, so no existing callers are broken.

## Non-goals

- `String` argument lifecycle — same leak pattern but a basic type; tracked separately from the container types.
- `PackedVector4Array` decode cases — tracked by `add-packed-vector4-array-decode`; once that lands, its lifecycle follows this capability.
- Changing Go→Godot encoding of Go-created container return values beyond the ptrcall copy-constructor alignment above.
- Adding a GC/finalizer-based release mechanism for container types.
