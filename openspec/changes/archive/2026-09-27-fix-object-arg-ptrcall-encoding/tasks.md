# Tasks

## 1. Nil-safe object pointer accessors

- [x] 1.1 Harden `WrappedImpl.AsGDExtensionObjectPtr()` and `AsGDExtensionConstObjectPtr()` in `pkg/builtin/wrapped.go` to return a nil pointer on a nil receiver instead of panicking, and add the shared `ObjectArgPtr(Wrapped)` helper in `pkg/builtin/object_arg.go` that covers the nil interface, a typed nil (via `reflect`), and a live object; verify with a Go test covering all three shapes plus nil `Owner`, and that a non-nil receiver still returns the owner pointer unchanged. *(Typed-nil coverage moved into the helper: a promoted method on a value-embedded Impl faults while Go addresses the embedded field, before any check inside the accessor can run — see design.md.)*
- [x] 1.2 Apply the same nil-receiver hardening to `WrappedClassInstance.AsGDExtensionObjectPtr()` and `AsGDExtensionConstObjectPtr()`, and add `RefArgPtr(Ref)` gating `ToObject()` behind `IsValid()`; verify the nil-receiver tests pass for this family too, that a typed-nil and invalid `Ref` both yield null, and that existing callers are unaffected.

## 2. Template: engine-class argument encoding

- [x] 2.1 In `cmd/generate/gdclassimpl/classes.go.tmpl`, replace the engine-class branch body with a local `GDExtensionObjectPtr` bound from the shared helper plus address-of on that local; verify by rendering `Node.AddChild` and reading the emitted Go to confirm the slot points at a pointer-valued cell. *(Rendered as `var argObj_0 GDExtensionObjectPtr; argObj_0 = ObjectArgPtr(node); argPtrSlice[0] = ...&argObj_0`.)*
- [x] 2.2 Route `Ref`-typed arguments through `RefArgPtr` (which gates `ToObject()` behind `IsValid()`), using the same `IsRefcountedClassName` predicate the signature already uses; verify by rendering `CollisionShape2D.SetShape` and confirming `argObj_0 = RefArgPtr(shape)` and that no `Reference()` call is introduced.
- [x] 2.3 Confirm the value-type branch is byte-for-byte unchanged; verified that `Node2D.SetPosition` does not appear in the regenerated diff at all and still emits `unsafe.Pointer(&position)`.

## 3. Regenerate and review

- [x] 3.1 Run `make generate` and verify it completes cleanly and `go build ./...` succeeds. *(Both exit 0.)*
- [x] 3.2 Review the regenerated diff in `pkg/gdclassimpl/classes.gen.go`; verified only object-argument bodies changed. Added lines are exclusively the 3-line pattern (901 decls, 171 `ObjectArgPtr`, 730 `RefArgPtr`, 901 slot assignments); removed lines are exclusively 901 old `unsafe.Pointer(&arg)` sites with **zero** other removals — no return-path, value-type, or signature lines moved.
- [x] 3.3 Confirm no `.Reference()` call was reinstated (only a warning comment mentions it) and that no object-argument site still passes an interface address. Completeness cross-checked against `extension_api.json`: 903 engine-class arguments, of which 2 sit on vararg methods that take the varcall branch (out of scope per the proposal), leaving 901 — exactly matching the 730 + 171 generated.

## 4. Tests

- [x] 4.1 Object-argument identity round-trip: `TestObjectArgIdentity` takes two Godot-supplied nodes, wires them with `AddChild` in Go, and the caller asserts `kid.get_parent() == parent` and `parent.get_child_count() == 1` in Godot. **Regression proof:** reverting the template to the buggy version, regenerating, and re-running crashes with `signal 11` at the first object-argument call (`main.gd:570`), so the tests do fail against pre-fix bindings.
- [x] 4.2 Node-spawn / `AddChild`: `TestObjectArgAddChild` passes a Godot-owned node into Go, which calls `AddChild` and asserts the child count grew by one and the parent identity matches. *Deviation:* the node is created in GDScript rather than in Go — Go has no engine-singleton instantiation path here (`ClassDB` is not reachable via `GetSingleton`), and a wrapper built over a nil owner would itself encode as null. The call shape that segfaulted is unchanged.
- [x] 4.3 `SetShape` round trip: `TestObjectArgSetShape` passes a `Ref<Shape2D>` through the `Ref` argument and the caller asserts `cs.shape == circle` in Godot. *Deviation:* readback is asserted in GDScript, not by reading a radius through Go. The generated object-**return** path allocates a `Shape2DImpl` wrapper and pins it on every `GetShape()` call, which leaves the resource reachable at shutdown and trips the leak check — isolating it showed `GetShape()` (not `SetShape`) causes the leak, since 51 `SetShape` calls with no `GetShape` leak nothing.
- [x] 4.4 Nil-argument tests: Go unit tests in `pkg/builtin/wrapped_nil_ptr_test.go` cover nil interface, typed nil, and invalid `Ref` against `ObjectArgPtr`/`RefArgPtr`; engine-side `TestObjectArgSetShapeTypedNil` and `TestObjectArgSetShapeInvalidRef` assert Godot really cleared the shape, and `TestObjectArgNilPlainObject` covers a nil non-refcounted argument via `set_owner(null)`. No Go panic in any shape.
- [x] 4.5 Reference-count stability: `TestObjectArgRefcountStability` passes the same `Ref` 50 times and asserts the count is unchanged from baseline. A reinstated `Reference()` or a `ToObject()` that acquired a reference would show as +50 and fail.

## 5. Integration verification

- [x] 5.1 `make test`: **1077 passes, 1 failure, no leaks, no crashes.** The single failure is pre-existing and unrelated — `test_suite` line 28 expects godot-cpp's `to_string` format (`[ GDExtension::Example <--> Instance ID:N ]`) but the engine returns `Example:<Example#N>`. Confirmed present with all new tests disabled (baseline 1062 passes, same 1 failure).
- [x] 5.2 Affected-surface coverage: all 901 ptrcall object-argument sites compile and use the new encoding (730 `RefArgPtr` + 171 `ObjectArgPtr`), spanning plain engine classes (`Node`), `Ref[T]` arguments (`RefShape2D`), and refcounted resources. *Not covered:* a user-defined class passed as a plain-object argument. That case crashes in the **inbound varcall decoder**, not the encoding under change — `getObjectInstanceBinding` dereferences a garbage pointer returned by `ObjectGetInstanceBinding` at `pkg/builtin/variant.go:115` (`addr=0x76`). It is a separate pre-existing defect and was left out of scope rather than patched here.
