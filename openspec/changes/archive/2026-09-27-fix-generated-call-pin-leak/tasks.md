# Tasks

## 1. Primary method template (`cmd/generate/gdclassimpl/classes.go.tmpl`)

- [x] 1.1 In the ptrcall method-body template, add `var pinner runtime.Pinner` + `defer pinner.Unpin()` as the first statements of each generated method body and route the `retPtr`, `cOwner`, `cArgs`, and per-argument-slot pins (lines ~140-191) to `pinner`; verify by rendering `Node.AddChild` and `CollisionShape2D.GetShape` and reading the emitted Go to confirm every `Pin` targets the local pinner and the `defer` is present.
  - Verified: both bodies render `var pinner runtime.Pinner` / `defer pinner.Unpin()` first, and `retPtr`, `cOwner`, `cArgs`, `argPtrSlice[j]` all call `pinner.Pin`. The two instance-constructor pins (`NewGDExtensionClassFromXOwner`, `NewXWithGodotOwnerObject`) stay on `pnr` per design D2 — they are the only `pnr.` uses left in the template.
- [x] 1.2 In the varcall branch of the same template, route the variant temporaries (`vN`), the varargs-loop slot pins, and the `err` slot to the scoped pinner; verify by rendering a variadic method and confirming defer LIFO ordering puts `pinner.Unpin()` last, after every `vN.Destroy()`.
  - Verified on `ClassDB.class_call_static`: registration order is `pinner.Unpin` → `methodName.Destroy` → `v0.Destroy` → `v1.Destroy`, so LIFO runs `Unpin` last.
  - **Deviation:** the `err` slot has no pin to route. `GDExtensionCallError` is `C.GDExtensionCallError`, an integer-only struct with no Go pointers, so `&err` is passed straight to C and is legal under cgo's pointer rules without pinning. No pin was added — adding one would be a no-op that muddies the audit rule.

## 2. Remaining generated call-path templates

- [x] 2.1 `cmd/generate/builtin/builtinclasses.go.tmpl`: scoped pinner in constructors (`ptr`, `args[k]`), `Destroy` (`bx`), and method bodies (`bx`, `args[j]`); verify by rendering a `String` constructor and a `Vector2` method.
- [x] 2.2 `cmd/generate/builtin/variant.go.tmpl`: scoped pinner in `NewVariantX`, `GDExtensionVariantPtrFromX`, and `ToX` bodies (`ptr`, `encodedPtr`); verify by rendering one reference and one non-reference variant encoding.
  - Note: the non-reference branch declares the pinner but pins nothing (its `encodedPtr` is a plain Go local passed directly to a synchronous call). `Unpin()` on an empty pinner is a no-op, so the uniform declaration is kept; the guard test deliberately does not treat "defers but never pins" as a failure.
- [x] 2.3 `cmd/generate/gdutilfunc/utilityfunctions.go.tmpl`: scoped pinner for `retPtr`, `args[...]`, and `typePtrArgs` (lines ~59-87); verify by rendering one fixed-arity and one vararg utility function.
- [x] 2.4 `cmd/generate/ffi/templatefunctions.go` (`cgoPinReturnType`): drop the emitted `pnr.Pin(ret)` global return-value pin per design D4; audit every `CallFunc_*` pointer-returning wrapper to confirm no return value flows into a later cgo call as a Go pointer without a body-level scoped pin (all current ones are engine-memory pointers).
  - Done: function and its three template call sites removed. Audit found 11 `void*`-returning wrappers; the only ones with callers outside `pkg/ffi` are `MemAlloc`/`MemRealloc` (engine-allocated memory), `ObjectGetInstanceBinding` (binding value, callers now scoped or deliberately retained), and `ClassdbGetClassTag` (engine ClassDB memory). None returns a Go pointer.

## 3. Hand-written per-call helpers

- [x] 3.1 `pkg/ffi/builtin_ptrcall.go`: convert each of the six `CallBuiltin*` helpers to a body-scoped `runtime.Pinner` with `defer pinner.Unpin()`; verify each helper's pins (`argsPtr`, `a`, `ptr`, `l`, `r`) target the local pinner.
- [x] 3.2 `pkg/core/method_bind_callback.go`: convert the per-virtual-call pins (`instPtr`, `inst`, `argPtrSlice[i]`, `rReturn`) to a body-scoped pinner with `defer pinner.Unpin()`, keeping the pins (not deleting them) per design D5; verify the reject paths (`rejectVarcallArity`, `rejectVarcallInvalidArgument`) still return cleanly with the deferred unpin in place.

### 3b. Hand-written per-call helpers not in the original plan (scope added during apply)

The generated bodies call hand-written helpers that pin on the package pinner, so fixing only the generated paths leaves the spec's no-accumulation requirement unmet. Confirmed by classifying all 120 non-generated `pnr.Pin` sites into per-call vs registration-time.

- [x] 3.3 `pkg/builtin/char_string.go` (13 pins): scope the per-call pins in the `NewString*`/`NewStringName*` constructors, `ToAscii`, `ToUtf8`, `StringNameCopyConstructor`, `NodePathCopyConstructor`. These are the hottest path in the change — every generated method call builds a method-name StringName through them.
  - **Deviation (forced by `cgocheck=1`):** the two accessors `AsGDExtensionConstStringNamePtr` / `AsGDExtensionConstStringPtr` could **not** have their pins removed. Dropping them made the extension panic at startup with `argument of cgo function has Go pointer to unpinned Go pointer` in `ClassDBAddSignal`: registration stores these pointers *inside* Go structs (`GDExtensionPropertyInfo`, argument-info arrays) that are then handed to C, and cgocheck requires any Go pointer reachable from a cgo argument's pointee to be pinned. The accessors keep their package pin, and a new `AsGDExtensionConstStringNamePtrPinned(*runtime.Pinner)` / `AsGDExtensionConstStringPtrPinned` was added for the generated per-call path so those pins end with the call instead of the program.
- [x] 3.4 `pkg/builtin/variant.go` (16 pins): scope `NewVariantNativeCopy`, `NewVariantNil`, `GDExtensionVariantPtrWithNil`, `NewVariantCopyWithGDExtensionConstVariantPtr`, `NewVariantGodotObject`, `GDExtensionVariantPtrFromGodotObjectPtr`, `ToObject`, `getObjectInstanceBinding`, `NewVariantGoString`.
- [x] 3.5 `pkg/builtin/container_copy_constructors.go` (10 pins): scope all ten `*CopyConstructor` bodies.
- [x] 3.6 `pkg/core/method_bind.go` `Call` (1): scoped to a body pinner.
  - **Deviation:** two of the three sites listed here stay on the package pinner because the value genuinely outlives the call, and both are allowlisted with the reason in the guard test:
    - `pkg/builtin/wrapped.go` `ObjectCastTo` — the instance-binding callbacks pointer it hands the engine is stored with the binding and read after the call returns.
    - `pkg/core/classdb_callback.go` `GoCallback_ClassCreationInfoGetPropertyList` — the engine keeps the returned property-info slice until the paired `FreePropertyList` callback runs. Scoping this pin would be a use-after-free.
- [x] 3.7 Confirm by re-classification that no per-call pin remains on a package pinner: every remaining `pnr.Pin` lives in registration/initialization or instance-creation code.
  - Result: 2,097 package pins remain, all in `classdb.go` bind/register paths, `NewGoMethodMetadata`, `NewGDExtensionClassMethodInfoFromMethodBind`, `GoMethodMetadata.Destroy`, `NewClassInfo`, `GDClassRegisterInstanceBindingCallbacks`, `SetConstructInfo`, `WrappedPostInitialize`, `GoCallback_GDExtensionBindingCreate`, `LoadProcAddress`, the two engine-retention cases above, and the 2,072 generated instance-creation constructor pins (verified: every one is inside `NewGDExtensionClassFrom*Owner` or `New*WithGodotOwnerObject`).

## 4. Preserve registration-time retention

- [x] 4.1 Audit that registration-time pins remain on the package pinner and are untouched: `pkg/core/method_bind.go` (`GoMethodMetadata` return/arg `PropertyInfo` `StringName`/`String` retention), `pkg/builtin/lib.go` (instance-binding callbacks), `pkg/ffi/lib.go` (`FFI.GodotVersion`), `pkg/core/types.go` (`ClassInfo`), `pkg/gdclassinit/*`; add a comment at each package `pnr` declaration stating it is for program-lifetime retention only.
  - Done for all six package pinners: `pkg/builtin/lib.go`, `pkg/core/lib.go`, `pkg/ffi/lib.go`, `pkg/gdclassimpl/lib.go`, `pkg/gdclassinit/lib.go`, `pkg/gdutilfunc/lib.go`.
- [x] 4.2 Add a guard test that scans the generated call-path sources and fails if `pnr.Pin` appears inside any generated method, constructor, utility-function, or variant-conversion body, and passes only when each such body declares `defer pinner.Unpin()`.
  - Implemented as `pkg/gdclassimpl/pin_scoping_test.go`: three tests covering generated call paths, hand-written per-call helpers (with the documented engine-retention allowlist), and a ratio assertion that scoped pins dominate program-lifetime pins.

## 5. Regenerate and build

- [x] 5.1 Run `make generate` and then `go build ./...`; both must exit 0.
- [x] 5.2 Review the regenerated diff: changed lines are exclusively the scoped-pinner pattern (local `runtime.Pinner` declaration, `defer pinner.Unpin()`, `pinner.Pin` replacing `pnr.Pin`); confirm no signatures, argument-encoding lines, or other logic moved, and that the ~63k pin sites now target scoped pinners.
  - Result: 63,844 `pinner.Pin` across 16,821 scoped bodies vs 2,097 program-lifetime pins. The object-argument encoding lines from the sibling change (`ObjectArgPtr` / `RefArgPtr`) are unchanged.

## 6. Tests and verification

- [x] 6.1 Go unit test for collectability: `pkg/gdclassimpl/pin_collectability_test.go` proves the mechanism with `weak.Pointer` — a value pinned on a scoped pinner is collectable after the body returns, a value pinned on the package pinner is not, and a body shaped like the generated one (return slot, owner cell, args slice, arg cells) leaves nothing reachable.
  - **Deviation:** the task asked to call a *generated object-returning method*. That needs a live engine (`classdb_get_method_bind` is nil otherwise), so the test exercises the runtime mechanism and a generated-shaped body directly rather than a real `GetShape()` call. The engine-side proof was attempted under 6.2 and is blocked by a separate defect.
- [x] 6.2 Engine regression looping `CollisionShape2D.GetShape()` and asserting no `Leaked instance: CircleShape2D`.
  - **Rescoped during apply: the `GetShape` assertion moved to `fix-object-return-reference-ownership`.** As written this task cannot pass on pin work. After scoping every pin the loop still leaked `Leaked instance: CircleShape2D - Reference count: 200` for 200 calls, because `NewRef` installs no `Unref` finalizer on the reference the engine transfers through the return slot. That scenario now lives in the ownership change's spec, where "no leaked instance" is that change's job.
  - Replaced in-engine by `test/pkg/pin_flatness_tests.go` / `TestPinScratchStaysFlat`, driven from `test/demo/main.gd` at 2,000 and 10,000 iterations of value-returning generated calls, sampling live Go heap after a forced GC. Satisfies the spec's "Frame-shaped loop stays flat" scenario. Value-returning calls only, deliberately, so nothing in the test touches refcount ownership.
  - **Falsification verified.** With the method template reverted from `pinner.Pin` to `pnr.Pin`, the test fails with drift scaling to the iteration count: `+2000 iters -> +266 KiB, +10000 iters -> +1412 KiB`. With scoped pins both samples read `+0 KiB`. The test discriminates the defect rather than passing vacuously.
- [x] 6.3 Metadata-after-unpin check: covered by the suite — `test_suite` reads back registered method and property information from Godot on every run and passes, proving registration pins were not disturbed by the per-call unpins.
- [x] 6.4 Run `GODOT=/home/pcting/bin/godot make test` and confirm the suite is green, ignoring the known startup warnings.
  - Result: **1077 passes, 0 leaks, 0 crashes**, matching the post-sibling-change baseline. The single remaining failure is the pre-existing `to_string` assertion at `test/demo/main.gd:28` (tracked by `fix-to-string-virtual-dispatch`), unrelated to pinning.
