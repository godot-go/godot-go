# Tasks

## 1. Pin the defect before coding (investigation, no code changes)

- [ ] 1.1 Read the `object_get_instance_binding` / `object_set_instance_binding` contract in `godot_headers/godot/gdextension_interface.h` (~2525-2550) and record in `design.md`: the binding is an extension-defined opaque `void*`, the token must match between set and get, and callbacks create the binding on demand. Cross-check `../gdext` (godot-rust) for how it shapes the binding value if present.
- [ ] 1.2 Inventory every writer of the binding slot — `SetConstructInfo` (`pkg/builtin/wrapped_gdclass.go:58`), `WrappedPostInitialize` (`:103`), `GoCallback_GDExtensionBindingCreate` (`pkg/gdclassinit/wrapped_gdextension_class.go:27-29`), `GoCallback_GDClassBindingCreate` (`:112-114`) — recording the stored shape (`cgo.Handle` value vs `*Object` pointer) and value type (`Wrapped` vs `*WrappedClassInstance`) per site; confirm the generated setter types `p_binding cgo.Handle`.
- [ ] 1.3 Reproduce the crash: a GDScript call passing a user-defined instance into a Go method with a plain `Node` parameter; confirm the fault is the `*instPtr` deref at `pkg/builtin/variant.go:115` and that the returned value equals a live `cgo.Handle` number. Also confirm `GDNativeConstructors` has no entry for user-defined class names. Write findings into `design.md` before any code edit.

## 2. Canonicalize the binding shape on cgo.Handle

- [ ] 2.1 Add `objectFromBindingPtr(unsafe.Pointer) (Object, error)` in `pkg/builtin`: decode the pointer as a `cgo.Handle` value, normalize `Wrapped` and `*WrappedClassInstance` values to `Object`, and return a typed error (class name + reason) for null, non-handle, or nil-valued results. Go unit tests cover each shape.
- [ ] 2.2 Change `GoCallback_GDExtensionBindingCreate` to return a `cgo.NewHandle`-shaped binding instead of `&inst`; keep the wrapper alive per the existing never-delete-after-retrieval policy; verify engine-class binding creation still round-trips through the new decode helper.
- [ ] 2.3 Align `SetConstructInfo` and `WrappedPostInitialize` so both store the value type the decode helper normalizes, and confirm every set/get site passes the same `FFI.Token`.

## 3. Rewire the lookup and decode path

- [ ] 3.1 Rewrite `getObjectInstanceBinding` (`pkg/builtin/variant.go:106-169`) to use the decode helper for both the nil-callbacks first call and the callbacks second call; remove the `(*Object)` cast/deref and replace `log.Panic("unable to get instance")` with the typed error. Keep `Variant.ToObject`'s signature (nil on typed failure).
- [ ] 3.2 Update the `gdObjectType` branch of `convertVariantToGoTypeReflectValue` (`pkg/core/method_bind_reflect.go:220-254`) to use the error-returning lookup and pass the resolved wrapper through when it implements the declared parameter type, falling back to the `GDNativeConstructors` re-wrap only for engine-class names; propagate typed errors into the existing varcall error reporting.

## 4. Tests

- [ ] 4.1 Add a `test/pkg` test (exposed to `test/demo/main.gd`): pass a `TestHierarchicalDerived` instance into a Go method taking a plain `Node` argument; assert the argument is non-nil, reports the right class, and carries the same Godot instance id as the original; call twice and assert the same wrapper resolves.
- [ ] 4.2 Engine-class regression: existing object-argument round-trips (`test_node_argument`, `test/pkg/object_arg_tests.go`) still pass with the create-callback shape change.
- [ ] 4.3 Typed-failure test: a lookup that cannot resolve (class with no registered callbacks / binding under a foreign token) returns the typed error and the rejected varcall is reported to the engine without aborting the process.

## 5. Verification

- [ ] 5.1 `go build ./...` succeeds.
- [ ] 5.2 `GODOT=/home/pcting/bin/godot make test` is green, ignoring the known startup warnings (`get_variant_get_internal_ptr_func`, `callable_custom_get_user_data`, `classdb_register_extension_class_5/6`).
- [ ] 5.3 Confirm no `*.gen.*` file was hand-edited; any template need went through `cmd/generate` + `make generate`.
