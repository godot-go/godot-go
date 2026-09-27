# Proposal

## Why

Passing a user-defined Go class instance (e.g. `TestHierarchicalDerived`, a Control-derived extension class from `test/pkg/hierarchical.go`) from GDScript into a Go method that declares a plain engine-class parameter (`Node`) crashes with SIGSEGV in the inbound varcall decoder. The fault is at `pkg/builtin/variant.go:115`: `getObjectInstanceBinding` casts the `void*` returned by `object_get_instance_binding` to `*Object` and dereferences it, but the engine hands back the binding this extension itself stored — a `cgo.Handle` value (`addr=0x76`, a small integer, not an address). The getter decodes a different shape than the setter registers.

The header contract (`godot_headers/godot/gdextension_interface.h:2525-2550`) makes the binding an opaque `void*` defined by the extension; getter and setter must agree on shape and token. Today they do not:

- User-defined classes register the binding as a `cgo.Handle` value (`pkg/builtin/wrapped_gdclass.go:58` and `:103`); the generated setter even types the parameter `p_binding cgo.Handle` (`pkg/ffi/ffi_wrapper.gen.go:4245`).
- Engine-class bindings are created by `GoCallback_GDExtensionBindingCreate` returning `&inst` — a pointer to a Go interface — which happens to match the getter's cast, so engine-class arguments resolve today.
- The user-defined create callback `GoCallback_GDClassBindingCreate` returns `nullptr`, so the callback-driven fallback cannot produce a wrapper either, and the decode branch's `GDNativeConstructors` map holds only engine-class names.

This surfaced while testing the just-landed object-argument encoding change (`e727e39`) and lives in the **inbound varcall decode path** (`convertVariantToGoTypeReflectValue` → `Variant.ToObject`), independent of that fix; its task 5.2 explicitly excluded it.

## What Changes

- Unify the instance-binding shape on `cgo.Handle`, matching the generated setter's typed signature: the engine-class binding create callback returns a handle instead of `&inst`, and `getObjectInstanceBinding` decodes the returned pointer as a `cgo.Handle` whose value resolves to an `Object`, never dereferencing the raw pointer.
- Make the decode path use the resolved wrapper: when the binding resolves to a Go wrapper implementing the declared parameter type, pass it through instead of re-wrapping by class name (user-defined class names are absent from `GDNativeConstructors`).
- Defensive failure: a lookup that cannot resolve (null binding, non-handle value, no live `Object`) returns a typed error naming the class and reason, surfaced through existing varcall error reporting, never a crash.

## Capabilities

### New Capabilities

- `instance-binding-lookup`: How the GDExtension instance-binding slot is read and decoded so any object crossing Godot→Go — engine-class or user-defined extension class — resolves to its Go wrapper, with a typed failure when it cannot.

### Modified Capabilities

(none)

## Impact

- `pkg/builtin/variant.go` — `getObjectInstanceBinding` decode plus an error-returning lookup helper.
- `pkg/gdclassinit/wrapped_gdextension_class.go` — engine-class create callback returns a handle.
- `pkg/builtin/wrapped_gdclass.go` — registration sites agree on the handle value type.
- `pkg/core/method_bind_reflect.go` — `gdObjectType` branch uses the resolved wrapper.
- `test/pkg/*.go` exposed to `test/demo/main.gd` — new object-argument tests.

## Non-goals

- No change to the `object_set_instance` class-instance slot (governed by `class-instance-lifecycle`).
- No change to outbound Go→Godot object-argument encoding (landed in `e727e39`).
- No hand-editing of `*.gen.*` files; any template need goes through `cmd/generate` and `make generate`.
- No handle-deletion or binding-free lifecycle redesign; free callbacks stay as-is.
