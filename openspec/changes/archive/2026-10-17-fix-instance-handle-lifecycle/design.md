# Design

## Context

See proposal.md for motivation. The current retention map around instance-scoped `cgo.Handle`s (verified on the stack tip):

| Writer | Handle payload | Lifetime anchor today | Release hook |
|---|---|---|---|
| `WrappedPostInitialize` (`pkg/builtin/wrapped_gdclass.go`) | `*WrappedClassInstance` | `cgo.NewHandle` + 5 `pnr.Pin` calls | none — leak (issue #152) |
| `SetConstructInfo` (same file; dormant, no callers) | `Wrapped` | `cgo.NewHandle` + `pnr.Pin(&cbs)` | none — leak |
| `GoCallback_GDExtensionBindingCreate` (`pkg/gdclassinit/wrapped_gdextension_class.go`) | `Object` wrapper | `cgo.NewHandle` only | `GoCallback_GDExtensionBindingFree` exists but is a commented-out no-op |

Engine facts pinned from `../godot` (4.x `core/object/object.cpp`):

1. `Object::_predelete` calls `_extension->free_instance(class_userdata, _extension_instance)` and then immediately sets `_extension = nullptr; _extension_instance = nullptr`. No extension-instance callback can follow a returned `free_instance` on the same object.
2. `Object::set_instance_binding(p_token, p_binding, p_callbacks)` copies the four fields out of `p_callbacks` — the engine never retains the callbacks struct pointer. The same copy is why a per-instance `pnr.Pin(&callbacks)` is pointless.
3. Instance-binding free callbacks fire from `Object::~Object` (every slot), from explicit `Object::free_instance_binding(token)`, and from `Object::clear_internal_extension` (TOOLS hot-reload). Each removes the slot first/atomically — one notification per slot, never two.
4. For our own user-defined classes, `WrappedPostInitialize` stores the *same* handle value in both the extension instance slot (`ObjectSetInstance`) and the binding slot (`ObjectSetInstanceBinding` with `FFI.Token`). The binding slot's free callback (`GoCallback_GDClassBindingFree`) therefore runs **after** `free_instance`, holding an already-deleted value.
5. `cgo.Handle` semantics (Go stdlib `runtime/cgo`): ids are monotonic and never reused; `Value()` on a deleted handle panics (it does not return garbage); `Value()` keeps the stored value reachable while the handle exists — the handle alone is a lifetime anchor. `#149` already made the binding-slot reader (`objectFromBindingPtr`) recover that panic into a typed error.

## Goals / Non-Goals

**Goals:**
- Every instance-scoped handle has exactly one delete point, tied to the engine's release notification for its shape.
- No per-instance Go value survives on the package pinner `pnr`.
- A churn regression gate that fails when either release is removed.

**Non-goals:**
- The `pnr.Pin(inst)` calls inside generated `NewGDExtensionClassFrom*Owner` / `New*WithGodotOwnerObject` constructors (`pkg/gdclassimpl/classes.gen.go` and its `cmd/generate` template). Those retain native-object wrapper structs created on the Go side and need a wrapper-lifetime design of their own (proposal Non-goals). This change does not touch the template.
- The `GDClassInstances` SyncMap re-entrancy guard semantics (#138) — preserved verbatim.
- Callable handles (`go_callable.go` already calls `ud.Delete()`) and method-metadata handles (program-lifetime by design).

## Decisions

### D1. Release in `FreeInstance` with LIFO-defer ordering

```go
ptrHandle := cgo.Handle(unsafe.Pointer(ptr))
wrapped := ptrHandle.Value().(*WrappedClassInstance)
defer ptrHandle.Delete()              // registered first → runs last
defer wrapped.Instance.Destroy()      // runs first
// existing SyncMap guard + Delete(id) stay in the body before the defers fire
```

Deferred calls run LIFO, so any re-entrant callback issued while `Destroy()` executes still sees a live handle — this is exactly the scenario that motivated the original never-delete policy. By the time `Delete()` runs, `Destroy()` has returned and the engine (fact 1) will null its copy of the pointer without another callback.

*Rejected — delete from `GoCallback_GDClassBindingFree` instead:* the binding slot (fact 4) carries the same handle and fires later, at `~Object`; `free_instance` is strictly earlier and is also the hook used by hot-reload's `clear_internal_extension`. Making the binding callback the owner would force `FreeInstance` to skip deletion and would still keep the wrapper alive through the whole destructor chain.
*Rejected — keep never-delete (status quo):* the leak is the issue; the original "returns garbage" rationale is factually wrong (fact 5).

### D2. Pin nothing per instance; call-scoped pins only where values cross

Replace the five `pnr.Pin` calls in `WrappedPostInitialize` with a body-scoped `runtime.Pinner` (`defer pinner.Unpin()`) pinning only the two values that cross the FFI calls within the function: `cnPtr` and `&callbacks`. Drop `obj`, `inst`, `owner` entirely:

- `obj`/`inst` — reachable through the handle for exactly as long as the engine can call back (fact 5); a pin adds nothing.
- `owner` — a C `GodotObject` pointer; `Pin` is a no-op.
- `cnPtr`/`&callbacks` — passed into `ObjectSetInstance`/`ObjectSetInstanceBinding` within the call; the scoped pinner satisfies the cgo pointer rules for the call duration, and the engine's by-value copy (fact 2) covers the callbacks afterward.

`SetConstructInfo` gets the same treatment for its `cbs` parameter: scoped pinner, no `pnr`.

*Rejected — the issue's literal per-instance `runtime.Pinner` stored on `WrappedClassInstance` and released in `FreeInstance`:* it preserves pins whose function the handle already serves, adds a field and an unpin step, and would keep `inst` alive one callback longer than necessary during teardown. The handle *is* the instance-owned anchor.

### D3. `SetConstructInfo` payload becomes `*WrappedClassInstance`

Dormant API (no callers today), so the shape change is zero-runtime-risk. Storing `Wrapped` would break `FreeInstance`'s `.(*WrappedClassInstance)` assertion on the first future user; aligning now lets the single release path (D1) serve both writers. `objectFromBindingPtr` already decodes `*WrappedClassInstance` (its `#149` switch).

*Rejected — broaden `FreeInstance` to a type switch over `Wrapped`:* extra branching for a shape nothing needs; the repo standardized on `*WrappedClassInstance` in #149.

### D4. `GoCallback_GDExtensionBindingFree` deletes the engine-class binding handle

`p_binding` round-trips the create callback's `cgo_handle_to_ptr(handle)` value unchanged. The callback converts back (`cgo.Handle(uintptr(p_binding))`) and calls `Delete()`, guarding a nil `p_binding`. One notification per slot (fact 3) makes double-delete impossible on this path. `GoCallback_GDClassBindingFree` **stays an explicit no-op** with a comment naming D1/fact 4 as the reason (its slot value was already deleted by `FreeInstance`).

### D5. Churn regression gate mirrors `pin_flatness_tests.go`

New in-engine fixture tests in `test/pkg/` measuring live Go heap across forced GCs, reusing the band + scaling-check pattern:

- **User-defined instance churn:** a RefCounted-derived test class instantiated and dropped in batches from GDScript or via `ClassDB` construction, so the engine runs create → free per iteration. Assert heap-band flatness and, additionally, that `weak.Pointer`s to captured wrappers report empty after the batch is freed — proving both handle deletion and unpinning.
- **Engine-class binding churn:** pass freshly created native objects into Go method calls (varcall decode triggers the binding create callback), then free the objects. **Amended during task 1.2:** the binding wrapper is *also* pinned by the generated `New*Owner` constructor pinner (explicit non-goal), so weak-survival and heap-band assertions cannot isolate the handle release on this path. The binding gate instead asserts release accounting: exported atomic create/free counters in `pkg/gdclassinit` must show exactly one free per create after churn, with no panic through teardown.

Baseline for the gate: current `pin_flatness` band logic (`band = 1 MiB`, scaling check with slack), retuned per churn cost if needed.

### D6. Verification by mutation proof

Full gates (`GODOT=... make test`: runtime stage incl. ObjectDB leak gate, editor stage) green; then mutation checks: (a) remove `defer ptrHandle.Delete()` → user-defined churn gate must fail; (b) empty `GDExtensionBindingFree` → binding churn gate must fail; (c) re-add `pnr.Pin(&callbacks)` → user-defined churn gate must fail (proves the pin path is observed). Restore and confirm green.

### D7. Handle values cross the C boundary as integers (added during implementation)

Churn runs exposed a fatal class orthogonal to the leak: `runtime.adjustframe`
throws `invalid pointer found on stack` when a pointer-bitmap stack slot holds
`0 < value < minLegalPointer` (4096), and the gate env arms the check via
`GODEBUG=invalidptr=1`. A packed `cgo.Handle` is exactly such a value. The
observed crash frame was a generated `CallFunc_...ObjectSetInstanceBinding`
cgo trampoline: `C.cgo_handle_to_ptr` returns C `void*` → a Go
`unsafe.Pointer`-typed argument slot, live (forced spill) across the C call;
GC stack scan (copystack/shrink of a goroutine parked in the call) rejected
handle id `0xb7d`. Ids ≥4096 are tolerated as C pointers — which is why low-
churn runs with few created handles never hit it, and 25k-instance churn
walks the counter through the danger zone while parked in create calls.

Decision: the ffi generator treats handle-carrying slots as a typed fact
(`PrimativeType.IsHandleSlot`, marked by `cmd/generate/ffi/handle_slots.go`)
and renders them `uintptr_t` through the C shims and `cgo.Handle` on the Go
side, for: `p_binding` parameters, `GDExtensionInterfaceObjectSetInstance`'s
`p_instance`, and the void* results of `object_get_instance_binding` /
`callable_custom_get_user_data`. Engine→Go binding callbacks
(`GDClassBindingFree`, `GDExtensionBindingCreate/Free`) cross the handle as
`uintptr` with casts in the (generated) C thunks; struct userdata fields
(`method_userdata`, `callable_userdata`) are written via the
`cgo_set_handle_slot` shim from `pkg/log/cgo_ptr.h`. `cgo_handle_to_ptr` is
deleted so the hazard cannot reappear.

Residual, accepted: engine→Go callbacks that receive a packed handle as a
pointer-typed argument (`free_instance`'s `p_instance`, ptrcall method-data,
callable userdata) convert to `cgo.Handle` at function entry, so the slot is
dead before any scan point; Go-allocated struct values holding handles
(callable info) stay protected by their existing `runtime.Pinner` discipline
(pinned frames are not moved or scanned). Three consecutive churn runs
(25k+5k instances) after the fix show zero bad-pointer events.

## Risks / Trade-offs

- [Stale handle value sits in the binding slot between `_predelete` and `~Object`] → readers already degrade to typed errors (`objectFromBindingPtr` recover, #149); no unguarded reader runs in that window on engine paths.
- [Double-delete between the two free callbacks] → impossible by construction (D1 owner + D4 no-op), enforced by code comments citing fact 4 and a unit-level doc test of the no-op.
- [Engine calls a bare `cgo.Handle(...).Value()` callback site after free (would panic across cgo)] → engine nulls `_extension`/`_extension_instance` before returning (fact 1); the 12 bare readers in `classdb_callback.go` are only reached via that pointer. Runtime + editor gates would surface any violation as a crash.
- [Churn gate flakiness from GC timing] → force two `runtime.GC()` cycles between samples like `pin_flatness`; band includes allocator noise slack; scaling check catches systematic retention.
- [`SetConstructInfo` shape change has no runtime coverage (dormant)] → compile-time check plus the payload-shape unit assertion in `pkg/builtin` where cgo allows; behavior covered transitively by the `WrappedPostInitialize` churn gate.
- [Native-class wrapper structs remain pinned by generated constructors] → explicitly out of scope (Non-goals); this change's churn gate measures only the instance/binding graphs it releases, so the follow-up change starts from a green baseline.

## Migration Plan

Internal binding change; no user-facing API migration. Rollback = revert the feature commit (specs/docs follow). Order: land after the current PR stack merges to avoid rebase churn on `classdb_callback.go`.

## Open Questions

- Exact churn trigger for engine-class bindings (varcall decode vs. signal payload vs. explicit `GetInstanceBinding`) — spike task picks whichever the existing harness can drive deterministically.
- RefCounted-derived fixture class to reuse (existing test hierarchy class vs. a new minimal `ExampleChurn` fixture) — spike decides; either satisfies D5.
