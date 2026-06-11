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
