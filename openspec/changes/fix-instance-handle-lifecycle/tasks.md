# Tasks

## 1. Spike: churn triggers and fixture selection

- [x] 1.1 Confirm a deterministic engine-side churn trigger for user-defined class instances: create N instances of a RefCounted-derived test class (existing hierarchy class or minimal new fixture) so the engine runs `CreateInstance2` → `FreeInstance` per iteration; verify counts by logging both callbacks. Verify: N create + N free callback lines in run output.
- [x] 1.2 Confirm the churn trigger for engine-class binding create: pass freshly created native objects into a Go method call (varcall decode) and check `GoCallback_GDExtensionBindingCreate` fires per object and `GoCallback_GDExtensionBindingFree` fires on object destruction. Verify: paired create/free callback lines; record the chosen trigger in tasks.md notes for task 4.2.

## 2. User-defined instance release (D1–D3)

- [x] 2.1 Rewrite `GoCallback_ClassCreationInfoFreeInstance` in `pkg/core/classdb_callback.go`: register `defer ptrHandle.Delete()` first, `defer wrapped.Instance.Destroy()` second, keep the SyncMap guard + `Delete(id)` body intact; replace the obsolete never-delete comment block with one citing the engine null-out (`_predelete`) and LIFO rationale. Verify: full `GODOT=... make test` runtime stage green (re-entrancy guard scenarios unchanged).
- [x] 2.2 In `WrappedPostInitialize` (`pkg/builtin/wrapped_gdclass.go`): remove all five `pnr.Pin` calls; add a body-scoped `runtime.Pinner` with `defer Unpin()` pinning only `cnPtr` and `&callbacks` across the two FFI calls. Verify: `grep -n "pnr.Pin" pkg/builtin/wrapped_gdclass.go` shows no hits in `WrappedPostInitialize`; suite green.
- [x] 2.3 Align `SetConstructInfo`: drop `pnr.Pin(&cbs)` (use the same body-scoped pinner pattern), store a `*WrappedClassInstance` payload built from `w.(Object)` instead of the raw `Wrapped`, and update the doc comment to name `FreeInstance` as the release owner. Verify: compile + suite green; payload shape matches the `FreeInstance` assertion (add a short package-level assertion or doc-test where cgo constraints allow).
- [x] 2.4 Make `GoCallback_GDClassBindingFree` an explicit documented no-op (comment names the shared-handle/D1 reason). Verify: comment present; no `Delete` call in that function.

## 3. Engine-class binding release (D4)

- [x] 3.1 Implement `GoCallback_GDExtensionBindingFree` in `pkg/gdclassinit/wrapped_gdextension_class.go`: nil-guard `p_binding`, convert to `cgo.Handle`, call `Delete()`; remove the commented-out body. Verify: with 2.1 landed, full runtime stage green (binding churn through existing #149/#160 fixtures exercises create→free pairs without panic).
- [x] 3.2 Update the stale never-delete comment in `GoCallback_GDExtensionBindingCreate` to describe the paired release. Verify: diff review; no other site references the removed policy.
- [x] 3.5 (Scope added during implementation, user-approved — design D7) Generator: mark handle-carrying slots (`cmd/generate/ffi/handle_slots.go` + `PrimativeType.IsHandleSlot`) and render them `uintptr_t` through the C shims / `cgo.Handle` on the Go side — `p_binding` params, `object_set_instance` `p_instance`, `object_get_instance_binding` + `callable_custom_get_user_data` returns; update Go consumers (`object_binding.go`, `wrapped_gdclass.go`). Verify: `make generate` regenerates `pkg/ffi/ffi_wrapper.gen.*` with `uintptr_t` shim slots; no `cgo_handle_to_ptr` call sites remain.
- [x] 3.6 Callback-side typing: `GoCallback_GDExtensionBindingCreate` returns `uintptr`, `GoCallback_GDExtensionBindingFree` takes `uintptr` (+ casts in `classes.callbacks.{c,h}.tmpl` → regenerated); `GoCallback_GDClassBindingFree` takes `uintptr` (+ `wrapped.c`/`wrapped.h`). Verify: build green; churn run shows zero `invalid pointer found on stack`.
- [x] 3.7 Struct userdata slots (`method_userdata`, `callable_userdata`): constructors take `cgo.Handle`, write through the new `cgo_set_handle_slot` shim in `pkg/log/cgo_ptr.h`; delete `cgo_handle_to_ptr`. Verify: `NewGDExtensionClassMethodInfo`/`NewGDExtensionCallableCustomInfo2` call sites pass handles directly; startup bind + callable fixtures green.

## 4. Churn regression gate (D5)

- [x] 4.1 Add user-defined instance-churn fixture in `test/pkg/` (pattern: `pin_flatness_tests.go`): batch create/free via the 1.1 trigger, force two GCs between samples, assert per-instance heap band (200 B/created — pre-fix ~270 B/instance leaks, post-fix runtime handle-table floor ~60-65 B/instance is linear and unavoidable, so scaling-vs-batch checks are invalid; band + weak are the discriminators); capture `weak.Pointer`s to wrappers and assert they report empty after the batch is freed. Verify: fixture passes in `make test` output; expected FAIL when 2.1 or 2.2 is temporarily reverted (prove in task 5).
- [x] 4.2 Add engine-class binding-churn fixture using the 1.2 trigger: exported atomic create/free counters in `pkg/gdclassinit` must report one free per create after churn (per amended design D5 — heap/weak assertions do not isolate the handle release because the generated constructor pin dominates). Verify: fixture passes; expected FAIL when 3.1 is temporarily reverted (prove in task 5).
- [x] 4.3 Wire fixtures into the GDScript driver and update gate baselines; record the new PASSES baseline in this file's notes. Verify: `check_test_output.sh` and `check_editor_output.sh` green with the updated counts; `make test_gate` fixture tests still pass.

## 5. Full verification and mutation proofs (D6)

- [x] 5.1 Run the complete two-stage gate (`GODOT=/home/pcting/bin/godot make test`): runtime stage incl. ObjectDB leak gate and misuse probes, editor stage incl. orphan-StringName checks. Verify: both stages PASS, exit 0. (Three consecutive green runs after the D7 boundary fix: exit 0, 1386 PASSES / 0 FAILURES, zero bad-pointer events each.)
- [x] 5.2 Mutation proof A: remove `defer ptrHandle.Delete()` → user-defined churn gate fails; restore. Mutation proof B: empty `GDExtensionBindingFree` → binding churn gate fails; restore. Mutation proof C (reworded during implementation — pinning only the callbacks struct copies ~24 B and the weak signal tracks the instance graph): re-add `pnr.Pin(inst)` in `WrappedPostInitialize` → user-defined churn gate fails; restore. Verify: three observed failures + final green run, results logged here.
- [x] 5.3 Run `revive` lint and `gofmt` check; confirm no new finding categories versus baseline. Verify: lint output compared to baseline. (gofmt clean; `go vet` shows only the pre-existing preprocessor unkeyed-fields findings; revive findings limited to baseline categories — receiver-name inconsistency, unused callback params — none in new/changed code.)

## 6. Documentation and spec sync

- [x] 6.1 Update `docs/overview.md` (and `docs/roadmap.md` status if tracked) to describe instance-scoped handle ownership. Verify: docs mention the single-release-owner rule. (New "Object Lifetime and Handle Ownership" section: one-release-owner rule, integer FFI crossings, pinning discipline. Roadmap does not track this issue.)
- [ ] 6.2 Sync delta specs into `openspec/specs/` (`class-instance-lifecycle`, `instance-binding-lookup`, `native-call-pinning`) at archive time per the repo's archive flow. Verify: `openspec validate fix-instance-handle-lifecycle --strict` passes before archive; main specs carry the reversed policy after archive.

## Workflow follow-up

- Land on branch `feat-fix-instance-handle-lifecycle` stacked on the current merge state of `main`; commit pattern `feat:` → `docs:` → `chore: archive`.
- PR body concatenates this change's `proposal.md` + `design.md` (per stack convention) and includes `Closes #152` so the reporter gets credit.
- Note for the follow-up change: generated `New*Owner` constructor pins (`pnr.Pin(inst)` in `pkg/gdclassimpl/classes.gen.go`) remain — the churn gate baselines here so the wrapper-lifetime change starts green.

### Notes (fill during implementation)

- 1.1 chosen user-defined trigger: GDScript `ExampleRef.new()` + local drop (refcount→0 → engine free → FreeInstance). Capture via `ObjectFromInstanceBinding` → `*ExampleRef`; `weak.Make` survival + heap band both effective. Observed pre-fix: +1328 KiB/5k → +6400 KiB/25k (linear), weak_alive 2048/2048, `Internal.GDClassInstances` live count returns to baseline (create+free fired N/N).
- 1.2 chosen binding trigger: GDScript `RefCounted.new()` + Go `ObjectFromInstanceBinding` (lazy binding create) + local drop (→ `~Object` → binding free). Observed pre-fix: +693 KiB/5k, weak_alive 2048/2048. DESIGN CORRECTION: wrapper stays pinned by the generated `New*Owner` constructor pnr pin (explicit non-goal), so weak/heap assertions cannot isolate the binding-handle release → 4.2 gates on exported binding create/free counter equality (once per create, no panic) instead of heap flatness. Design D5 amended accordingly.
- 4.3 new PASSES baseline: **1386** (4 new asserts: churn end×2 + churn verify + binding verify; previous 1382).
- Boundary fix stability: 3/3 consecutive full two-stage runs green post-D7 with `GODEBUG=invalidptr=1` armed; zero `invalid pointer found on stack`.
- 5.2 mutation results: A (remove `defer ptrHandle.Delete()`) → 3 FAILURES incl. "2048 freed-instance wrappers still reachable" + band breach, gate exit 2 ✓. B (empty `GDExtensionBindingFree`) → 1 FAILURE "5000 churned binding handles were not released", exit 2 ✓. C (re-add `pnr.Pin(inst)`) → 1 FAILURE "2048 freed-instance wrappers still reachable", exit 2 ✓. All mutations restored; final green run below.
