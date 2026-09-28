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

## Investigation findings (task 1, recorded before any code edit)

### The header contract

`object_get_instance_binding(GDExtensionObjectPtr p_o, void *p_token, const GDExtensionInstanceBindingCallbacks *p_callbacks) -> void*`
and `object_set_instance_binding(p_o, p_token, void *p_binding, p_callbacks)`. The binding is an
**extension-defined opaque `void*`**: the engine never interprets it, it only stores and hands it
back, keyed by the library `p_token`. If no binding exists yet and `p_callbacks` is non-null, the
engine calls `create_callback(token, instance)` to make one. The callbacks struct is
`{create, free, reference}`. So the shape of the binding is entirely this project's choice —
which is precisely why two different shapes in the same slot is a bug and not an engine quirk.

### Every writer of the binding slot, and the shape each stores

| # | Site | Stores | Value behind the shape |
|---|---|---|---|
| 1 | `SetConstructInfo` — `pkg/builtin/wrapped_gdclass.go:49→58` | `cgo.NewHandle(w)` value | `Wrapped` (interface) |
| 2 | `WrappedPostInitialize` — `pkg/builtin/wrapped_gdclass.go:95→103` | `cgo.NewHandle(inst)` value | `*WrappedClassInstance` |
| 3 | `GoCallback_GDExtensionBindingCreate` — `pkg/gdclassinit/wrapped_gdextension_class.go:27-29` | `&inst` — **address of a Go interface local** | `Object` |
| 4 | `GoCallback_GDClassBindingCreate` — `pkg/builtin/wrapped_gdclass.go:112` | `nullptr` | none |

The generated setter types the parameter `p_binding cgo.Handle` (`pkg/ffi/ffi_wrapper.gen.go:865`,
`:881`, `:4226`), and the FFI generator special-cases it on purpose
(`cmd/generate/ffi/templatefunctions.go:24` — "void* parameter called p_binding is a cgo.Handle
slot, not a raw…"). So D1 is not a preference; it is what the generated interface already asserts.
Site 3 is the sole violator.

### The reader assumes site 3's shape

`getObjectInstanceBinding` (`pkg/builtin/variant.go:121-132`) does

```go
instPtr := (*Object)(CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(...))
if instPtr != nil && *instPtr != nil {
    return *instPtr
}
```

That is correct **only** for site 3's pointer-to-interface. For sites 1 and 2 the returned
`void*` *is* the handle number, so `*instPtr` dereferences a small integer.

### Reproduction

Passing a `TestHierarchicalDerived` (a registered extension class, `Control`-derived) into the
existing `TestObjectArgAddChild(child Node)` from GDScript:

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x85 pc=0x7f7a84e2c7bb]
github.com/godot-go/godot-go/pkg/builtin.getObjectInstanceBinding(0x55aa5ebd8920)
    pkg/builtin/variant.go:131 +0xdb
github.com/godot-go/godot-go/pkg/builtin.(*Variant).ToObject(...)
    pkg/builtin/variant.go:116 +0x17a
```

`addr=0x85` is the decisive detail: a fault at a tiny address is a handle number being used as a
pointer, not a null or a wild address. The proposal named line 115; the code has drifted to 131 and
the deref is the same one.

### Why engine-class arguments work today

Engine classes get their binding through site 3, whose shape the reader matches. User-defined
extension classes get theirs through sites 1 and 2. That is the whole split: the reader was written
against one writer and never reconciled with the other two.

### `GDNativeConstructors` and user-defined names

`GDNativeConstructors` is populated only by the generated init
(`pkg/gdclassinit/classes.init.gen.go:6244+`) with **engine** class names. `TestHierarchicalDerived`
has no entry, confirming that any decode path keyed on that map cannot resolve a user-defined class,
which is why D4's pass-through of the already-resolved wrapper is required rather than a nicety.

## Corrections found during implementation

**There was a third reader, not two.** `ObjectCastTo` (`pkg/builtin/wrapped.go`) assumed a
third shape — it cast the binding to `*WrappedClassInstance` and dereferenced `wci.Instance`,
under a `// TODO: validate this is working as expected` that was never honoured. It has no
callers anywhere in the repo, which is the only reason it never surfaced. The inventory in this
document said four writers and one reader; it is four writers and **three** readers. It was fixed
rather than left behind, because unifying the shape while leaving a known-wrong reader is how the
next version of this bug gets written.

**The decode helper takes `uintptr`, not `unsafe.Pointer`.** Two independent reasons pushed the
same way: the value being decoded is a pointer-sized integer that must never be dereferenced, only
reinterpreted as a `cgo.Handle`, so `uintptr` keeps an unsafe conversion off the API surface; and
this project's build rejects cgo in `_test.go` files, so an `unsafe.Pointer` parameter could not
be exercised by unit tests without routing through the C packing shim. The result is that all
seven shapes of the slot are now covered by a plain `go test`, with no engine.

**Task 4.3's "without aborting the process" is not achievable inside this change.** The typed
error is produced and named, and never SIGSEGVs, but the pre-existing varcall boundary
`log.Panic`s on any conversion error and there is no `recover()` anywhere in the production
varcall path. The D4 falsification shows the typed error arriving intact and then that boundary
aborting godot with exit 134. Making the boundary non-fatal is a change to
`method-call-error-reporting` itself; absorbing it here would have hidden a second defect inside
the first one's fix.
