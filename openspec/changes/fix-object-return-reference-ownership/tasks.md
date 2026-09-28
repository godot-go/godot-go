# Tasks

## 1. Classify call sites and decide the fix
- [x] 1.1 Inventory all `NewRef[...]` call sites and classify owning vs borrowing.
  - **1,325 call sites**, not 1,323: 661 `NewRefXGDExtensionIternalConstructor` + 661 `NewRefXAsRef` in `classes.refs.gen.go`, plus 3 nil placeholders in `test/pkg` (`example.go:597,621`, `object_arg_tests.go:135`). The 1,326th grep hit is the `func NewRef[T ...]` definition itself, not a call.
  - The classification in the design's table turned out to be wrong for half the `IternalConstructor` sites — see 1.2.
- [x] 1.2 Cross-reference the engine for ptrcall return handling; determine whether `NewXWithGodotOwnerObject` receives a transferred reference.
  - **Return path transfers, confirmed.** `godot/core/variant/method_ptrcall.h`: `PtrToArg<Ref<T>>::convert` returns `Ref<T>(*reinterpret_cast<T *const *>(p_ptr))`, and constructing a `Ref` calls `reference()`. The +1 lands on whoever reads the slot. Matches the original evidence of exactly one orphaned reference per call.
  - **`WithGodotOwnerObject` does NOT transfer.** It passes the factory a Go-built struct (`inst := &XImpl{}; inst.Owner = owner`) — no engine write, no `convert`, no `+1` anywhere. The object's lifetime belongs to whoever supplied `owner`. Wrapping it as owning would release a reference that was never acquired.
  - Blast radius narrowed by two checks: the only in-repo caller is `GetInputSingleton()`, and `Input` is `is_refcounted: false` in `extension_api.json`, so it takes the plain `return inst` branch. The refcounted branch of `WithGodotOwnerObject` has **no callers in this repo** — the risk is exported-API surface, not live breakage.
  - Generated evidence after the fix: 661 `IternalConstructor(inst)` (all from `WithGodotOwnerObject`, borrowing) and 574 `ReturnOwner(&ret)` (all ptrcall returns, owning), with **zero** `IternalConstructor(&ret)` left.
- [x] 1.3 Confirm or revise the option (a) vs (b) decision.
  - Option (b), a distinct owning constructor, confirmed and approved. **Wiring revised**: because the factory is shared by two callers with opposite semantics, the owning behaviour could not be put into `IternalConstructor` as Decision 3 assumed. Added `NewRefXGDExtensionReturnOwner` used only by the return line; `IternalConstructor` left borrowing.

## 2. Implement the owning return path
- [x] 2.1 Add the owning constructor to `pkg/builtin/ref_generic.go`.
  - `NewRefTransfer[T]` wraps without a refcount change and installs the `Unref` finalizer. It skips the finalizer when `ObjectArgPtr(obj) == nil`: `Unref` issues a live `unreference` ptrcall, so a finalizer on a wrapper over no engine object would dereference a null owner. A `get_shape()` on an empty collision shape hits exactly that.
  - `NewRef`'s doc rewritten. It previously claimed "transfer semantics", which is what made it read as the right home for the return path; it is the borrowing constructor and is now documented as such, pointing transferred-reference callers at `NewRefTransfer`.
- [x] 2.2 Update `cmd/generate/gdclassimpl/classes.refs.go.tmpl` so the return path uses the owning constructor; leave `AsRef` and `IternalConstructor` borrowing.
- [x] 2.3 Run `make generate`; verify the regenerated diff touches only generated outputs and matches the template change (no hand edits to `*.gen.*`).

## 3. Regression tests
- [x] 3.1 Refcount-stability test: N repeated object returns keep the count constant.
  - `TestReturnRefcountStability`, driven with 200 iterations: `base=3, +200 returns -> 3`. This is the loop `test_object_args` explicitly withheld because the leak made it unrunnable.
- [x] 3.2 Drop-then-free test: dropping an owned wrapper drives the reference down when nothing else holds it.
  - `TestReturnDroppedReleases`: `base=3, held=4, dropped=3` — the finalizer releases exactly the one transferred reference.
- [x] 3.3 Idempotence test: explicit `Unref()` then forced GC performs no second release.
  - `TestReturnUnrefIdempotence`: `base=3, after Unref=3, after GC=3`.
- [x] 3.4 Borrow-safety test: a dropped borrowing wrapper leaves the true owner's count unchanged.
  - `TestBorrowedRefNeverReleases`: `base=3, after dropping borrower=3`.
- [x] 3.5 Wire the 200-call `get_shape()` loop into `test/demo/main.gd` and assert a clean engine exit.
  - `test_return_ownership` added and called from `_ready`. The stale comment in `test_object_args` explaining why the loop could not run is replaced with a pointer to the new test.
- [x] 3.6 (added during implementation) Null-return test: `TestNullReturnSchedulesNoRelease` runs 50 `get_shape()` calls against an emptied collision shape and survives collection, covering the `ObjectArgPtr` guard in `NewRefTransfer`.

## 4. Verify
- [x] 4.1 `go build ./...` — clean. `go vet ./pkg/... ./test/pkg/...` — clean.
- [x] 4.2 `make generate && make build` — regenerated 661 owning factories, no hand edits.
- [x] 4.3 `GODOT=/home/pcting/bin/godot make test` — **1085 assertions, 0 failures, 0 leaked instances**, three consecutive identical runs.
  - The task's expected "known pre-existing `to_string` failure" is gone; `fix-to-string-virtual-dispatch` landed first.
- [x] 4.4 Falsification. Removing the finalizer from `NewRefTransfer` reproduces the original defect exactly: `base=3, +200 returns -> 203`, `FAIL: 200 references retained`, two `Leaked instance` markers, `make` exit 2. Restored gives 1085/0 and exit 0.
