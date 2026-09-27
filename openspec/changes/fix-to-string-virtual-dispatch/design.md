# Design

## Context

Godot asks an extension class for a display string through the `GDExtensionClassToString` creation-info callback (`godot_headers/godot/gdextension_interface.h:241`):

```c
typedef void (*GDExtensionClassToString)(GDExtensionClassInstancePtr p_instance, GDExtensionBool *r_is_valid, GDExtensionStringPtr p_out);
```

The engine contract: the callback writes the result into `p_out` and sets `*r_is_valid` to true. If validity is not set, the engine discards `p_out` and uses the built-in `Object::to_string()`. The callback is wired at `pkg/core/classdb.go:624` (`C.cgo_classcreationinfo_tostring`).

Two verified defects in `GoCallback_ClassCreationInfoToString` (`pkg/core/classdb_callback.go:45-64`):

1. `r_is_valid = (*C.GDExtensionBool)(&isValid)` rebinds the local pointer parameter; nothing is written through to the engine's bool. The engine always sees `false`.
2. The string is hardcoded (`fmt.Sprintf("[ GDExtension::%s <--> Instance ID:%d ]", ...)`). The user's `to_string` virtual — `V_Example_ToString` in `test/pkg/example.go:920`, registered via `ClassDBBindMethodVirtual(t, "V_Example_ToString", "to_string", ...)` into `ci.VirtualMethodMap["to_string"]` — is never invoked.

Note the engine invokes `to_string_func` directly; it does NOT route through `call_virtual_with_data`. So dispatch must happen inside this callback, not via `GoCallback_ClassCreationInfoCallVirtualWithData`.

Encoding: Go `string` is UTF-8. The current code uses `GDExtensionStringPtrWithLatin1Chars`, which tells Godot to interpret the bytes as Latin-1 — correct only for ASCII. The UTF-8 constructor (`GDExtensionStringPtrWithUtf8Chars` → `string_new_with_utf8_chars`) is the correct choice for arbitrary Go strings.

## Goals / Non-Goals

**Goals:**
- The engine accepts the binding's string (validity written through the pointer).
- A user-registered `to_string` virtual is dispatched and its return value delivered to Godot.
- Classes with no override still get the existing default format, reported as valid.
- The out `String` is constructed as UTF-8 so non-ASCII content renders correctly.

**Non-Goals:**
- No change to the virtual registration API (`ClassDBBindMethodVirtual`) or the varcall dispatch path.
- No change to engine-default `to_string` for non-extension classes.
- No caching of the resolved virtual per instance.

## Decisions

**Decision 1: Write validity through the pointer.**
Replace the local rebind with `*r_is_valid = 1` (a `C.GDExtensionBool`) on every path that supplies a string. This matches the header contract and godot-cpp's `to_string` handling, which sets `*r_is_valid = true` alongside the out string.

**Decision 2: Dispatch via the existing virtual method map.**
Inside the callback, look up the class in `Internal.GDRegisteredGDClasses` and check `ci.VirtualMethodMap["to_string"]`. If present, invoke it with `GoMethodMetadata.Func.Call([]reflect.Value{reflect.ValueOf(inst)})` — the same reflection pattern used by `_notification`, `_get`, and the property-revert callbacks — and take `reflectedRet[0].String()` as the result. If absent (or the class lookup fails), fall back to the default format. Rationale: reuses the registration route the user already used (`"to_string"` gd-name), requires no new plumbing, and mirrors how sibling callbacks resolve virtuals.

Rejected alternative: routing through `GoCallback_ClassCreationInfoCallVirtualWithData`. The engine never calls the varcall path for `to_string`; it calls `to_string_func` directly, so that route would be dead code.

**Decision 3: UTF-8 out construction.**
Use `GDExtensionStringPtrWithUtf8Chars(p_out, value)` for both the dispatched and fallback strings. The fallback format is pure ASCII so behavior is unchanged for it; the dispatched user string gains correct non-ASCII rendering. Latin-1 is only correct when the source bytes are Latin-1, which Go strings are not.

**Decision 4: Keep the default format as the fallback.**
When no virtual is registered, emit the current `[ GDExtension::<class> <--> Instance ID:<id> ]` string with validity true. This preserves the format the demo assertion expects and guarantees a non-empty, identifying string without depending on the engine default.

## Risks / Trade-offs

- **Reflection return-type mismatch:** a user virtual registered as `to_string` that does not return a Go `string` would panic on `.String()`. Mitigation: the binding's registration already encodes the return type; guard with a type assertion and fall back to the default format with a warning log rather than panicking, keeping `to_string` (a display-only path) non-fatal.
- **UTF-8 switch changes behavior for hypothetical Latin-1 users:** any existing code relying on Latin-1 byte reinterpretation of non-ASCII would render differently. Accepted: UTF-8 is the correct interpretation of Go strings, and the Latin-1 path was only ever accidentally safe for ASCII.
- **Per-call map lookup:** resolving the virtual on every `to_string` call adds a map lookup. Negligible: `to_string` is display-only and low-frequency.
