# Tasks

## 1. Pin the defect before coding (investigation, no code changes)

- [x] 1.1 Header contract recorded in `design.md`: the binding is an extension-defined opaque `void*`, keyed by the library token, created on demand through the `create` callback. The shape is entirely this project's choice, which is why two shapes in one slot is a bug and not an engine quirk. (`../gdext` is not present in this workspace; the engine headers and the project's own generated setter were used as the authority instead.)
- [x] 1.2 Inventory recorded. **Four writers, three shapes**: `SetConstructInfo` stores a `cgo.Handle` to `Wrapped`; `WrappedPostInitialize` stores a `cgo.Handle` to `*WrappedClassInstance`; `GoCallback_GDExtensionBindingCreate` stored `&inst`, the address of a Go interface local; `GoCallback_GDClassBindingCreate` stores null. Confirmed the generated setter types the parameter `p_binding cgo.Handle` (`pkg/ffi/ffi_wrapper.gen.go:865`, `:881`, `:4226`) and that the FFI generator special-cases it deliberately (`cmd/generate/ffi/templatefunctions.go:24`).
- [x] 1.3 Reproduced. `TestHierarchicalDerived` into `TestObjectArgAddChild(child Node)` gives `SIGSEGV ... addr=0x85` at `variant.go:131` via `Variant.ToObject`. The tiny fault address is the proof: a handle number used as a pointer. Also confirmed `GDNativeConstructors` is populated only with engine class names by the generated init, so no name-keyed path can resolve a user-defined class.

## 2. Canonicalize the binding shape on cgo.Handle

- [x] 2.1 `objectFromBindingPtr` in `pkg/builtin/object_binding.go`, plus `ObjectBindingError`, `ObjectFromInstanceBinding`, `engineObjectClassName`, `godotObjectPtrFromVariant`, and `ObjectFromVariant`.
  - **Deviation from the task's signature.** The helper takes `uintptr`, not `unsafe.Pointer`. Two reasons: the value being decoded is a pointer-sized integer that must never be dereferenced, only reinterpreted as a handle, so `uintptr` keeps an unsafe conversion off the API surface; and this project's build rejects cgo in `_test.go` files ("use of cgo in test ... not supported"), so an `unsafe.Pointer` parameter could not be exercised by the unit tests without routing through the C shim.
  - 8 Go unit tests in `pkg/builtin/object_binding_test.go`, no engine needed: null slot, `*WrappedClassInstance` happy path, `*WrappedClassInstance` with a nil `Instance`, plain `Object`, unrecognised value (`42`), **deleted handle** (`cgo.Handle.Value()` panics and the recover turns it into the typed error), and both error-message forms. `fakeObject` satisfies `Object` by embedding the interface, since only identity is under test.
- [x] 2.2 `GoCallback_GDExtensionBindingCreate` now returns `cgo.NewHandle(inst)` packed through the C shim. The shim (`pkg/log/cgo_ptr.h`) exists precisely because `unsafe.Pointer(uintptr(h))` trips `go vet`'s unsafeptr check, which cannot know a Handle is pointer-stable; the first attempt without it failed vet. No delete is added, matching the never-delete-after-retrieval policy that the previous permanently-pinned `&inst` already followed.
- [x] 2.3 Confirmed both user-defined writers store value types the decode helper normalises (`Wrapped`, `*WrappedClassInstance`), and **every** set/get site passes the same `FFI.Token` — verified by grepping all six call sites.

## 3. Rewire the lookup and decode path

- [x] 3.1 `getObjectInstanceBinding` reduced to a delegation to `ObjectFromInstanceBinding`, returning nil on typed failure so `Variant.ToObject` keeps its documented contract. The `(*Object)` cast, the dereference, and `log.Panic("unable to get instance")` are gone — 66 lines replaced by 13. `Variant.ToObject` was also refactored to share `godotObjectPtrFromVariant` rather than duplicate the pointer extraction.
- [x] 3.2 The `gdObjectType` branch of `convertVariantToGoTypeReflectValue` now resolves via `ObjectFromVariant` and passes the resolved wrapper through when it implements the declared parameter type. The `GDNativeConstructors` re-wrap remains only as the fallback, and its `log.Panic("unsupported interface class name")` became a returned `fmt.Errorf`.
- [x] 3.3 **(found during implementation, not in the original plan) A third reader existed.** `ObjectCastTo` (`pkg/builtin/wrapped.go`) cast the binding to `*WrappedClassInstance` and dereferenced `wci.Instance` — a third shape assumption, carrying its own `// TODO: validate this is working as expected`. It has **no callers anywhere in the repo**, which is why it never surfaced. Fixed to delegate to `ObjectFromInstanceBinding` rather than leave a known-wrong reader behind after unifying the shape; its redundant callbacks lookup and pins went with it.

## 4. Tests

- [x] 4.1 `test/pkg/user_defined_object_arg_tests.go`, driven from `test/demo/main.gd`:
  - `test_user_defined_node_arg` — `resolved TestHierarchicalDerived instance 28068283888 through a Node parameter`, instance id matches
  - `test_user_defined_arg_stable` — two resolutions of the same owner pointer produce the **identical** wrapper, so the pass-through is not building a new one per call
  - `test_engine_class_arg_still_resolves` — plain `Node` survives the create-callback shape change
  - `test_unresolvable_binding_is_typed_error` — see 4.3
  - Untyped GDScript vars on purpose: these must travel the varcall path; a typed declaration would route them to ptrcall.
- [x] 4.2 Engine-class regression: the existing object-argument round-trips still pass, plus the dedicated `test_engine_class_arg_still_resolves`.
- [x] 4.3 **Partially satisfied, and the remainder is a separate change.** The lookup does return a typed `*ObjectBindingError` naming the class and reason, and it never SIGSEGVs — verified directly. But the requirement that the rejected varcall be "reported to the engine **without aborting the process**" is not met, and cannot be met here: the pre-existing varcall boundary `log.Panic`s on any conversion error (`method_bind_reflect.go:42`) and there is **no `recover()` anywhere in the production varcall path** (only in tests). The falsification in 4.4 shows exactly this — the typed error arrives intact, then that boundary aborts godot with exit 134. Making the boundary non-fatal is a change to `method-call-error-reporting` itself, deliberately not absorbed here.
- [x] 4.4 Falsified both halves of the fix.
  - **Reader shape**: restored the pre-fix `instPtr := (*Object)(unsafe.Pointer(binding)); return *instPtr` and got `SIGSEGV ... addr=0x86` — the same class of fault as the original `0x85`, caught by the harness as `FAIL[no-driver-summary]`.
  - **Pass-through (D4)**: disabled `reflect.TypeOf(obj).Implements(t)` and got `unsupported object class "TestHierarchicalDerived" for parameter type builtin.Node` at `arg_index: 0`. This proves the pass-through is load-bearing rather than incidental, and that the name-keyed fallback genuinely cannot resolve a user-defined class.

## 5. Verification

- [x] 5.1 `go build ./...` clean.
- [x] 5.2 `GODOT=/home/pcting/bin/godot make test` — **1097 assertions, 0 failures, 0 leaked engine objects**, exit 0.
- [x] 5.3 No `*.gen.*` file was hand-edited: `make generate` produces **no diff in any generated file**. This change touches no template.
- [x] 5.4 `go vet ./pkg/... ./test/pkg/...` clean; `go test ./pkg/...` green.

## Follow-up (deliberately out of scope)

A varcall argument that fails to convert still aborts godot, because `reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs` `log.Panic`s and nothing recovers at the cgo boundary. The typed errors this change produces are the prerequisite for fixing that, but the fix belongs in `method-call-error-reporting`: report the rejection to the engine and return without calling the method, rather than taking the process down.
