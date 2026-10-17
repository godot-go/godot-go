# Proposal

## Why

Issue #152: every extension class instance created by the engine allocates a `cgo.Handle` that is never deleted, and `WrappedPostInitialize` pins per-instance values on the program-lifetime package pinner `pnr`. Nothing releases either when Godot frees the instance, so each instance and its Go reference graph stays reachable until process exit — a steady leak in long editor sessions and any instance churn. The safety rationale for never deleting (stale-handle lookups "return garbage") is obsolete: `cgo.Handle.Value()` panics (it does not corrupt), handle ids are never reused, `objectFromBindingPtr` already recovers dead-slot reads into typed errors (#149), and Godot nulls `_extension`/`_extension_instance` immediately after `free_instance` returns (`object.cpp` `_predelete`), so no engine callback follows a freed instance with the stale pointer.

## What Changes

- `GoCallback_ClassCreationInfoFreeInstance` deletes the instance `cgo.Handle` after `Destroy()` returns (deferred LIFO so re-entrant callbacks during `Destroy` still see a live handle).
- `WrappedPostInitialize` stops calling `pnr.Pin` for per-instance values (`obj`, `owner`, `inst`, `cnPtr`, `&callbacks`): the handle itself keeps the value reachable, `owner`/`cnPtr` are C pointers (pins are no-ops), and Godot copies the callbacks struct out by value.
- `SetConstructInfo` (dormant writer, no callers today) drops `pnr.Pin(&cbs)` and stores a `*WrappedClassInstance`-typed handle so the shared `FreeInstance` release path matches its payload shape.
- `GoCallback_GDExtensionBindingFree` deletes the binding `cgo.Handle` created by `GoCallback_GDExtensionBindingCreate` — the engine-class binding path's only release hook (fires from `Object::~Object`). `GoCallback_GDClassBindingFree` stays a deliberate no-op: its `p_binding` is the same handle `FreeInstance` already released, and deleting there would double-delete.
- Handle values stop crossing the C boundary inside GC pointer slots. Churn exposed a latent fatal (same boundary, opposite direction of the leak): the runtime rejects pointer-bitmap stack slots holding `0 < value < minLegalPointer` (4096) under stack scans, armed by the gate's `GODEBUG=invalidptr=1` — a packed handle is exactly such a value, and cgo call-argument slots are forced-spill live across the C call. Fix: the ffi generator marks handle slots (`p_binding` params, `object_set_instance`'s `p_instance`, the binding/userdata getters' return) and renders them `uintptr_t` through the C shims / `cgo.Handle` on the Go side; binding-callback signatures cross as `uintptr`; struct userdata fields are written via a `cgo_set_handle_slot` C shim; the `cgo_handle_to_ptr` packer is deleted.
- Add an in-engine instance-churn heap-flatness regression gate (same pattern as `pin_flatness_tests.go`), and reverse the spec policy that mandates the leak.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `class-instance-lifecycle`: the "cgo.Handle SHALL NOT be deleted" policy is reversed — the handle and all per-instance retention are released in `FreeInstance` after `Destroy`; `SetConstructInfo` payload shape and pinning change accordingly.
- `instance-binding-lookup`: the engine-class binding handle is released when the engine frees the binding; the never-delete policy for that path is removed; handle-slot values cross the C boundary as integers, never through GC pointer slots.
- `native-call-pinning`: program-lifetime `pnr` must not retain per-instance values from construction paths; registration-time retention remains allowed.

## Impact

- Code: `pkg/builtin/wrapped_gdclass.go`/`object_binding.go`/`wrapped.c`/`wrapped.h`, `pkg/core/classdb_callback.go`, `pkg/gdclassinit/wrapped_gdextension_class.go`, `pkg/ffi/class_method_info.go`/`callable_custom_info.go`, `pkg/log/cgo_ptr.h`; generator: `cmd/generate/ffi/handle_slots.go` (new)/`templatefunctions.go`/`generate.go`/`ffi_wrapper.c.tmpl`, `cmd/generate/gdclassinit/classes.callbacks.{c,h}.tmpl`, `cmd/gdextensionparser/clang/parser.go`; regenerated `pkg/ffi/ffi_wrapper.gen.{go,c,h}` and `pkg/gdclassinit/classes.callbacks.gen.{c,h}`.
- Tests: new churn gate in `test/pkg/`; full runtime + editor gate must stay green (ObjectDB leak gate, misuse probes).
- No public Go API signature changes outside the ffi package's generated wrappers and two `pkg/ffi` constructor parameter types (`unsafe.Pointer` → `cgo.Handle`).

## Non-goals

- Generated-owner-constructor pins (`pnr.Pin(inst)` inside `NewGDExtensionClassFrom*Owner` / `New*WithGodotOwnerObject`): these retain native-object wrapper structs and need their own wrapper-lifetime design; a follow-up change will move or remove them. This change still deletes their handles where a release hook exists.
- `cgo.Handle` for method-bind metadata (`method_bind.go`) — program-lifetime registration retention by design. Its cross-boundary typing is fixed here (integer slots), its lifetime is not.
- Go-callable handles — already released via `ud.Delete()`; only their cross-boundary typing (constructor parameter + slot write) changes here, lifetime semantics untouched. Engine→Go callbacks that convert their packed-handle argument to `cgo.Handle` at function entry are safe by liveness (the pointer slot is dead before any GC scan point) and are documented, not converted.
- Any change to the call-scoped pinning pattern shipped by `fix-generated-call-pin-leak`.
