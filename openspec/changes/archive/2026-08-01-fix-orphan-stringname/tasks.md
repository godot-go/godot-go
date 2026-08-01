## 1. ClassInfo Refactoring (pkg/core/types.go)

- [x] 1.1 Replace `NameAsStringNamePtr GDExtensionConstStringNamePtr` with `NameStringName StringName` (value field)
- [x] 1.2 Replace `ParentNameAsStringNamePtr GDExtensionConstStringNamePtr` with `ParentNameStringName StringName` (value field)
- [x] 1.3 Update `NewClassInfo` to assign `StringName` values into the new fields instead of storing pointers
- [x] 1.4 Update `Destroy()` to call `.Destroy()` on the value fields instead of dereferencing pointers
- [x] 1.5 Update `String()` method to read from value fields

## 2. ClassDB Registration Fixes (pkg/core/classdb.go)

- [x] 2.1 Update `ClassDBAddPropertyGroup` to read `ci.NameStringName` and pin before the Godot call
- [x] 2.2 Update `ClassDBAddPropertySubgroup` to read `ci.NameStringName` and pin before the Godot call
- [x] 2.3 Update `ClassDBAddProperty` to read `ci.NameStringName` and pin before the Godot call; remove redundant `NewStringNameWithLatin1Chars(cn)` allocation
- [x] 2.4 Update `ClassDBAddSignal` to read `ci.NameStringName` and pin before the Godot call; remove redundant `NewStringNameWithLatin1Chars(typeName)` allocation
- [x] 2.5 Pin all temporary `StringName` locals (property names, setter/getter names, signal names) at their respective call sites

## 3. Class Registration (pkg/core/lib.go)

- [x] 3.1 Update `ClassDBRegisterClass` to use the new `ClassInfo` value-stored `StringName` fields
- [x] 3.2 Ensure `ci.NameAsStringNamePtr` references are replaced with `ci.NameStringName.AsGDExtensionConstStringNamePtr()`
- [x] 3.3 Pin `ci.NameStringName` at registration call sites

## 4. Verification

- [x] 4.1 Run `make generate` and verify no generated code breaks
- [x] 4.2 Run `make test` — 43/43 tests pass, no cgo panics, no crashes (requires Godot binary)
- [x] 4.3 Godot correctly loads and registers classes/properties — verified by passing test suite (requires Godot binary)

## 5. Follow-Up (Known Issues, Out of Scope)

- [x] 5.1 Fix `NewSimpleGDExtensionPropertyInfo` (`pkg/core/object.go`) — stores `StringName` pointers to locals in `GDExtensionPropertyInfo`; used by `method_bind.go`. No explicit lifecycle management, no `Destroy()` called on returned property info. Consider storing `StringName` by value in `GDExtensionPropertyInfo` or adding explicit cleanup.
- [x] 5.2 Fix double-free in `ClassMethodInfo.Destroy()` (`pkg/ffi/class_method_info.go:68,72`) — calls destructor on `m.name` and `cm.name` which are the same pointer (type alias), causing double-unref.
- [x] 5.3 Add `C.free()` to `ClassMethodInfo.Destroy()` — struct allocated via `C.malloc()` in `NewGDExtensionClassMethodInfo` but never freed. Required for unload+reload safety.
- [x] 5.4 Address pre-existing callback risk — Godot-owned `GDExtensionConstStringNamePtr` in `classdb_callback.go` could be accidentally `Destroy()`'d. Consider read-only wrapper type or runtime guard.

## 6. Revise and Regenerate Documentation (docs/)

- [x] 6.1 `docs/overview.md` — fix typos ("bidnings", "Insteead", "acess"), update stale claims about virtual methods, packed arrays, default arguments, and static variables to match current bindings
- [x] 6.2 `docs/godot_gdextension_string.md` — correct `StringName` size from `[4]uint8` to `[8]uint8`; remove "StringNames don't need Destroy() calls in most cases" guidance; document the value-storage + local-copy-isolate-then-pin pattern
- [x] 6.3 `docs/godot-cpp-string-comparison.md` — remove references to the removed `NameAsStringNamePtr` API and the broken `NewGDExtensionPropertyInfo` flow; present the implemented fix (value storage + pinning + C-heap allocation) as the recommended pattern
- [x] 6.4 `docs/godot-cpp-string-comparison-notes.md` — mark the investigation as resolved; update task checkboxes and "Current Status" to reflect the implemented fix
- [x] 6.5 `docs/stringname-mutation-analysis.md` — verify consistency with the implemented code and keep as the authoritative reference
- [x] 6.6 Run `make test` and confirm 43/43 tests pass with no new orphaned StringName warnings (docs-only change, no code regeneration expected)

## 7. Eliminate Remaining Orphaned StringName (`Image`)

- [x] 7.1 Reproduce and isolate the `Image (static: 0, total: 1)` orphan via `main.gd` experiments — proven to be created when a valid RefCounted object crosses into Go
- [x] 7.2 Restore `defer snClassName.Destroy()` in `getObjectInstanceBinding()` (`pkg/builtin/variant.go`) — Godot's `object_get_class_name` placement-constructs the StringName and Go never destroyed it
- [x] 7.3 Fix the two ptrcall `object_get_class_name` call sites in `pkg/core/method_bind_reflect.go` to pass zero-value `StringName{}` storage instead of `NewStringName()` (pre-initialized storage was overwritten by placement-new, leaking the constructor refcount)
- [x] 7.4 Run `make test` — 43/43 tests pass with zero orphaned StringName warnings
- [x] 7.5 Update spec (remove "Godot-side Orphans Remain Acceptable" scenario) and retrospective (replace misdiagnosed "Godot-side limitation" section with actual root cause)
