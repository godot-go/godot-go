## Why

Godot-Go reports orphaned StringName warnings during `make test` due to Go-side lifecycle bugs in how `StringName` handles are passed to the GDExtension interface:

1. **Dangling pointers in `ClassInfo`**: `NewClassInfo` stored a pointer to a local `StringName` (`NameAsStringNamePtr`) that dangled after the constructor returned.
2. **GC memory movement**: Go's GC can relocate heap-allocated `[8]uint8` `StringName` values during registration calls, producing stale pointers and double-free errors.

Godot only ever **reads** the 8 bytes of a caller-provided `StringName` during registration (copy ctor, refcount increment). This makes storing `StringName` by value safe and is the basis for the fix.

Additionally, the project's documentation (`docs/`) describes the pre-fix, broken patterns and contains stale or incorrect claims (wrong `StringName` size, wrong Destroy() guidance, references to removed APIs). It must be revised to describe the implemented lifecycle management.

## What Changes

- **Store `StringName` by value in `ClassInfo`**: Replace the dangling pointer fields with `NameStringName` / `ParentNameStringName` value fields.
- **Pin at registration call sites**: Use `runtime.Pinner` on local copies of the value (local-copy-isolate-then-pin pattern) so GC cannot move them during the Godot call.
- **C-heap allocation for `GDExtensionClassMethodInfo`**: Allocate via `C.malloc()` so the Go 1.26 cgo checker (which cannot track pinning across call boundaries) allows the pointer through.
- **Fix related lifecycle bugs**: `GoMethodMetadata.Destroy()`, variadic `StringName` reuse, `PropertyInfo.Destroy()` switch-case bug, signal name retention in `ClassInfo`, and the double-free in `ClassMethodInfo.Destroy()`.
- **Revise all documentation** under `docs/` to describe the implemented `StringName` lifecycle management and remove stale/incorrect claims.

## Capabilities

### New Capabilities
- `stringname-lifecycle-management`: Manages the lifetime of `StringName` objects when passed to the GDExtension interface, eliminating Go-side "orphaned StringName" warnings and the panic/double-free hazards.

### Modified Capabilities
- None

## Impact

- **Code**: `pkg/core/types.go`, `pkg/core/classdb.go`, `pkg/core/lib.go`, `pkg/core/method_bind.go`, `pkg/ffi/class_method_info.go`, `pkg/ffi/property_info.go`.
- **Docs**: `docs/overview.md`, `docs/godot_gdextension_string.md`, `docs/godot-cpp-string-comparison.md`, `docs/godot-cpp-string-comparison-notes.md`, `docs/stringname-mutation-analysis.md`.
- **Verification**: `make test` — 43/43 tests pass, no cgo panics, no double-frees.

## Non-goals

- Eliminating Godot-side orphan warnings at exit (e.g. `Image` static refs) — Godot-side behavior, not addressable from Go.
- Changing the general-purpose `StringName` API for end-users.
- Modifying generated `*.gen.*` files beyond what `make generate` produces.
