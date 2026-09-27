# Tasks

## 1. Nil-safe object pointer accessors

- [ ] 1.1 Harden `WrappedImpl.AsGDExtensionObjectPtr()` and `AsGDExtensionConstObjectPtr()` in `pkg/builtin/wrapped.go` to return a nil pointer on a nil receiver instead of panicking; verify with a Go test that calls the accessor on a nil `*WrappedImpl` and on a typed-nil `Node` interface without a panic, and that a non-nil receiver still returns the owner pointer unchanged.
- [ ] 1.2 Apply the same nil-receiver hardening to `WrappedClassInstance.AsGDExtensionObjectPtr()` and `AsGDExtensionConstObjectPtr()`; verify the nil-receiver test passes for this family too and existing callers are unaffected.

## 2. Template: engine-class argument encoding

- [ ] 2.1 In `cmd/generate/gdclassimpl/classes.go.tmpl`, replace the engine-class branch body with a local `GDExtensionObjectPtr` bound from `AsGDExtensionObjectPtr()` plus address-of on that local, and a nil-interface guard; verify by rendering the template for `Node.AddChild` and reading the emitted Go to confirm the slot points at a pointer-valued cell.
- [ ] 2.2 Route `Ref`-typed arguments through `ToObject().AsGDExtensionObjectPtr()` (not `Ptr()`, which is absent from the `Ref` interface); verify by rendering `CollisionShape2D.SetShape` and confirming the emitted expression and that no `Reference()` call is introduced.
- [ ] 2.3 Confirm the value-type branch is byte-for-byte unchanged; verify by rendering a value-typed method such as `Node2D.SetPosition` and diffing against the pre-change output.

## 3. Regenerate and review

- [ ] 3.1 Run `make generate` and verify it completes cleanly and `go build ./...` succeeds.
- [ ] 3.2 Review the regenerated diff in `pkg/gdclassimpl/classes.gen.go`; verify only object-argument bodies changed — no return-path, value-type, or signature lines moved — using a filtered diff plus spot checks across several classes.
- [ ] 3.3 Grep the regenerated output to confirm the commented-out `.Reference()` was not reinstated and that no object-argument site still passes `unsafe.Pointer(&<interface>)` directly.

## 4. Tests

- [ ] 4.1 Add an object-argument identity round-trip test (Godot passes an object into a Go method, Go passes it into another Godot call); verify the instance id Godot sees on the second call matches the first, and confirm the test fails against the pre-fix bindings before it passes against the fix.
- [ ] 4.2 Add a node-spawn test: create a `CharacterBody2D` from Go, `AddChild` it to a live tree, assert the child count and identity; verify the call returns normally with no SIGSEGV — this is the scenario that crashes today.
- [ ] 4.3 Add a `SetShape` test: create a `CircleShape2D`, set its radius, assign it through the `Ref` argument, and read the radius back through the collision shape; verify the value round-trips.
- [ ] 4.4 Add nil-argument tests covering a nil interface, a typed-nil pointer, and an invalid `Ref` (`IsValid()` false); verify Godot receives a null object for each and no Go panic occurs.
- [ ] 4.5 Add a reference-count stability test: pass the same refcounted object as an argument across many calls and assert the reference count does not grow; verify it also catches a regression if `ToObject()` were later changed to acquire a reference.

## 5. Integration verification

- [ ] 5.1 Run `make test` and verify the full suite is green with no new failures or crashes.
- [ ] 5.2 Verify coverage of the affected surface: confirm the generated methods that take engine-class arguments (794 across 164 types per the proposal's count) now compile and exercise the new encoding, and that a representative sample spanning plain classes, `Ref[T]` arguments, and a user-defined `RefCounted` subclass all pass.
