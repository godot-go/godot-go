# Proposal

## Why

`Object.to_string()` on Go-registered GDExtension classes is broken in two independent ways inside `GoCallback_ClassCreationInfoToString` (`pkg/core/classdb_callback.go`):

1. **The validity flag never reaches the engine.** The callback assigns `r_is_valid = (*C.GDExtensionBool)(&isValid)`, which rebinds the local copy of the pointer parameter instead of writing through it. The engine's `GDExtensionBool *r_is_valid` (see `GDExtensionClassToString` in `godot_headers/godot/gdextension_interface.h:241`) stays `false`, so Godot discards the returned string and falls back to its default `Object::to_string()`.
2. **The callback never dispatches to the user's virtual.** It hardcodes `fmt.Sprintf("[ GDExtension::%s <--> Instance ID:%d ]", ...)`, so a user-registered `to_string` virtual — e.g. `V_Example_ToString` in `test/pkg/example.go:920`, registered into `ci.VirtualMethodMap["to_string"]` — is dead code.

Observed symptom: `example.to_string()` returns `Example:<Example#...>` (engine default) and the assertion at `test/demo/main.gd:28` fails. This is currently the only failing assertion in the suite and is pre-existing.

## What Changes

- In `GoCallback_ClassCreationInfoToString`, write the validity flag **through the pointer** (`*r_is_valid = 1`) so the engine accepts the result.
- Dispatch to the class's registered `to_string` virtual via `Internal.GDRegisteredGDClasses` / `ci.VirtualMethodMap["to_string"]` (the same registration route other virtuals use), calling the Go method by reflection and delivering its returned Go `string`.
- When the class registers no `to_string` virtual, fall back to the current default format `[ GDExtension::<class> <--> Instance ID:<id> ]`, still reported as valid.
- Construct the out `String` with `GDExtensionStringPtrWithUtf8Chars` (`string_new_with_utf8_chars`) instead of the Latin-1 variant: Go strings are UTF-8, and the Latin-1 constructor reinterprets each byte as a Latin-1 code point, corrupting any non-ASCII content. The ASCII-only default format masked this today.

## Capabilities

### New Capabilities
- `to-string-virtual-dispatch`: Godot's `to_string` creation-info callback reports validity through the out parameter, dispatches to a user-registered `to_string` virtual when present, falls back to a sensible non-empty default otherwise, and encodes the result as UTF-8.

### Modified Capabilities
(none)

## Non-goals

- No change to how virtuals are registered (`ClassDBBindMethodVirtual`) or to the varcall dispatch path.
- No change to the engine-default `Object::to_string()` behavior for non-extension classes.
- No new user-facing API; the `V_*_ToString` convention is unchanged.
