# Design

## Context

See `proposal.md` — Why, for the defect. The verified mechanics:

The GDExtension contract (`godot_headers/godot/gdextension_interface.h:2525-2550`) treats the instance binding as an opaque `void*`: `object_set_instance_binding(p_o, p_token, p_binding, p_callbacks)` stores whatever the extension passes, and `object_get_instance_binding(p_o, p_token, p_callbacks)` returns it (creating it via `p_callbacks` if absent). The extension alone defines the shape; getter and setter must agree, and the token must match.

godot-go currently writes that slot from three places with two incompatible shapes, and reads it with a cast that only matches one of them:

| Site | Shape stored | Matches getter? |
|---|---|---|
| `SetConstructInfo` (`pkg/builtin/wrapped_gdclass.go:58`) | `cgo.Handle` value wrapping `Wrapped` | No — getter casts to `*Object` and derefs |
| `WrappedPostInitialize` (`pkg/builtin/wrapped_gdclass.go:103`) | `cgo.Handle` value wrapping `*WrappedClassInstance` | No |
| `GoCallback_GDExtensionBindingCreate` (`pkg/gdclassinit/wrapped_gdextension_class.go:27-29`) | `&inst` — pointer to a pinned `Object` interface | Yes (accidentally) |
| `GoCallback_GDClassBindingCreate` (`pkg/builtin/wrapped_gdclass.go:112-114`) | `nullptr` | n/a — fallback cannot produce a wrapper |

The getter (`getObjectInstanceBinding`, `pkg/builtin/variant.go:106-169`) does `(*Object)(ret)` then `*instPtr`. For a user-defined instance the returned value is the handle number (e.g. `0x76` = 118), so the dereference faults. For engine-class objects the create callback's `&inst` happens to match, which is why engine-class arguments work today.

Downstream, the `gdObjectType` decode branch (`pkg/core/method_bind_reflect.go:220-254`) calls `arg.ToObject()` (the crash site) and then re-wraps by class name via `GDNativeConstructors` — a map populated only for engine classes by `classes.init.go.tmpl`, so a user-defined class name would fail there even if the binding decoded.

## Goals / Non-Goals

**Goals:**
- One canonical binding shape across every set/get site, with the token contract (`FFI.Token`) unchanged.
- User-defined extension-class instances passed as plain engine-class arguments decode to their existing Go wrapper, identity preserved.
- Engine-class argument decoding behavior unchanged from the caller's point of view.
- Every unresolvable lookup produces a typed, class-named error that flows into the existing varcall error reporting (`method-call-error-reporting`) instead of a panic or SIGSEGV.

**Non-Goals:**
- Changing the `object_set_instance` class-instance slot or its `cgo.Handle` policy (`class-instance-lifecycle` spec).
- Changing outbound Go→Godot object-argument encoding (`e727e39`).
- Redesigning binding free/lifecycle: free callbacks stay no-ops; handle retention follows the existing never-delete-after-retrieval policy.
- Any hand-edit of `*.gen.*` files.

## Decisions

**D1: Canonical binding shape is `cgo.Handle`.** Chosen because the generated setter already types its parameter `p_binding cgo.Handle` (`pkg/ffi/ffi_wrapper.gen.go:4245`) and both user-defined registration sites already store a handle — the smaller and more type-honest change. Alternative considered: store `*Object` everywhere (keeps the getter's cast) — rejected because it requires permanently pinning Go interface headers in engine memory and contradicts the setter's typed signature.

**D2: A single decode helper owns the shape knowledge.** Add `objectFromBindingPtr(unsafe.Pointer) (Object, error)` in `pkg/builtin`: treat the pointer as a `cgo.Handle` value, resolve `.Value()`, normalize the two known value types (`Wrapped`, `*WrappedClassInstance`) to `Object`, and return a typed error (class name + reason) for null, non-handle, or nil-valued results. `getObjectInstanceBinding` uses it for both the first (nil-callbacks) and second (callbacks) calls; the `log.Panic("unable to get instance")` path becomes the typed error. `Variant.ToObject` keeps its signature (returns nil on typed failure) so unrelated callers are unaffected; the varcall decode branch uses the error-returning form.

**D3: The engine-class create callback returns a handle.** `GoCallback_GDExtensionBindingCreate` wraps its constructed wrapper in `cgo.NewHandle` and returns the handle value instead of `&inst`, so engine-class bindings match D1. The wrapper is kept alive by the handle per the existing policy; no deletion is added.

**D4: The decode branch prefers the resolved wrapper over class-name re-wrap.** In `convertVariantToGoTypeReflectValue`'s `gdObjectType` case, if the resolved binding wrapper implements the declared parameter type, pass it through directly; fall back to the `GDNativeConstructors` re-wrap only when the binding did not resolve to a usable wrapper. This is what makes user-defined class names (absent from `GDNativeConstructors`) work and removes a redundant second wrapper in the common case.

**D5: Investigation precedes code.** Because the exact defect could plausibly be return-type mapping, registration shape, or a missing callbacks entry, task 1 is a read-and-record pass (header contract, every set/get site, a reproduction) whose findings are written back into this document before any edit.

## Risks / Trade-offs

- **Handle stored in engine memory outlives Go's control.** If a handle were ever deleted while the engine still holds it, `Value()` panics. Mitigated by the existing never-delete policy and by D2's typed failure on unresolvable handles; residual cost is handle retention, equivalent to the pinned `&inst` memory leaked today.
- **Object identity change for engine-class arguments.** D4 returns one wrapper where the old path created a second one per call. Code comparing wrapper pointers (rather than instance ids) could notice; identity by instance id is preserved, and the test suite pins round-trip identity.
- **`ToObject` returning nil on typed failure** is a behavior change from crash→nil for direct Go callers. Considered acceptable: a nil `Object` is already the documented result for a nil variant, and the varcall path — the one that matters for engine calls — propagates the typed error.
- **Changing the create callback touches every engine-class binding creation.** Blast radius is broad but mechanical; the full `make test` suite plus the existing object-argument round-trip tests (`test/pkg/object_arg_tests.go`) cover it.
