# Tasks

## 1. Classify call sites and decide the fix
- [ ] 1.1 Inventory all 1,323 `NewRef[...]` call sites and classify owning vs borrowing (661 `IternalConstructor` = engine-transferred; 661 `AsRef` = borrowing; 3 nil-placeholder test uses); confirm counts against `pkg/gdclassimpl/classes.refs.gen.go` and record in design.md.
- [ ] 1.2 Cross-reference godot-cpp (`../godot-cpp`, or the engine at `../godot`) for `_gde_internal_constructor` / ptrcall return handling to confirm the return slot transfers a +1 to the receiver, and determine whether `NewXWithGodotOwnerObject` (`classes.go.tmpl:75`) receives a transferred reference or needs init semantics; record the answer under design.md Decision 4.
- [ ] 1.3 Confirm or revise the option (a) vs (b) decision from design.md Decision 1; if deviating from (b), record the rationale before implementing.

## 2. Implement the owning return path
- [ ] 2.1 Add the owning constructor to `pkg/builtin/ref_generic.go`: wraps without refcount change and installs `runtime.SetFinalizer(r, (*RefBase[T]).Unref)`; update `NewRef`'s doc comment to point transferred-reference callers at the owning constructor.
- [ ] 2.2 Update `cmd/generate/gdclassimpl/classes.refs.go.tmpl` so `NewRefXGDExtensionIternalConstructor` uses the owning constructor; leave `NewRefXAsRef` on borrowing `NewRef`.
- [ ] 2.3 Run `make generate`; verify the regenerated diff touches only generated outputs and matches the template change (no hand edits to `*.gen.*`).

## 3. Regression tests
- [ ] 3.1 Refcount-stability test: N repeated object returns (get_shape-style) with dropped wrappers keep the object's reference count constant across iterations.
- [ ] 3.2 Drop-then-free test: dropping a returned owned wrapper and forcing GC drives the underlying reference count to zero and frees the object when nothing else holds it.
- [ ] 3.3 Idempotence test: explicit `Unref()` on an owned wrapper followed by forced GC performs no second release (borrowed-owner count stays correct).
- [ ] 3.4 Borrow-safety test: a dropped `AsRef`-style borrowing wrapper, collected by GC, leaves the true owner's reference count unchanged.
- [ ] 3.5 Wire the 200-call `CollisionShape2D.get_shape()` loop from the repro into `test/demo/main.gd` and assert a clean engine exit (no `Leaked instance` / leaked-RID errors).

## 4. Verify
- [ ] 4.1 `go build ./...`
- [ ] 4.2 `make generate && make build` (`make generate` deletes the built library and `make test` does not rebuild it, so `make build` is required between them).
- [ ] 4.3 `GODOT=/home/pcting/bin/godot make test` — expect the 1077-pass baseline plus the new tests, with only the known pre-existing `to_string` failure at `test/demo/main.gd:28`; ignore the known startup warnings (`get_variant_get_internal_ptr_func`, `callable_custom_get_user_data`, `classdb_register_extension_class_5/6`).
