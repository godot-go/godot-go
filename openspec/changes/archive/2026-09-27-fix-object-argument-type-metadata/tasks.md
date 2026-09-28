# Tasks

## 1. Class-name resolution helper (pkg/core)

- [x] 1.1 `godotClassNameForObjectType(t reflect.Type) (string, bool)` in `pkg/core/object_class_name.go`, implementing the D2 order: `GDNativeConstructors` → `Internal.GDRegisteredGDClasses` → `Ref`-prefix validated against `GDClassRefConstructors` → generic object interfaces fall back to `"Object"`.
  - **Pointer dereference added during implementation.** A pointer type's `Name()` is empty, and concrete Go class implementations *are* pointers (`*ExampleRef`), so every such return looked unresolvable. The helper now walks `t = t.Elem()` while `t.Kind() == reflect.Ptr`. This was discovered by the guard in 2.1 firing on a real bind, not by reading the design.
- [x] 1.2 Unit tests in `pkg/core/object_class_name_test.go`, seeding the registries directly (they are empty until extension init runs). Nine cases, all green: plain engine name, registered extension class, `RefXxx` → `Xxx`, a `Ref`-prefixed class name that must **not** be stripped, the four generic interfaces → `"Object"`, bare `Ref` must not yield the empty string, unknown type reports unresolved, pointer-to-registered-class, pointer-to-engine-class.
  - The registries cannot be swapped out — package-level vars of another package are read-only — so they are seeded and unseeded in place. `Internal.GDRegisteredGDClasses` is nil until init, so the test owns an instance for its duration, the same way `classdb_to_string_test.go` already does.

## 2. Wire resolution into the bind metadata

- [x] 2.1 Arguments: `objectClassForMetadata(...)` resolves OBJECT-variant arguments and leaves every other variant on the owning class unchanged.
- [x] 2.2 Return: same helper applied to the return `PropertyInfo`.
- [x] 2.3 **Deviation from the task as written.** The task asked for `pkg/core` unit tests of `NewGoMethodMetadata` itself. That function builds `StringName`s and so needs a live engine, which a plain `go test` does not have. The resolution logic — the part that can be wrong — is covered by 1.2; the wiring is covered in-engine by 3.2's typed declarations and by the falsification in 4.4, which shows the typed call sites are genuinely gated by the advertised class.
- [x] 2.4 **Spec revision: unresolvable OBJECT types no longer panic.** The spec required a bind-time panic. It fired on a legitimate case: `ReturnEmptyRef` returns `*pkg.ExampleRef`, a Go-defined `RefCounted` whose `RegisterClassExampleRef()` is commented out at `test/pkg/lib.go:20`, so the type is a valid object that is simply not a registered Godot class. Panicking broke a working bind. The requirement now specifies advertising `"Object"` with a warning: truthful, since every engine object is an `Object`, so no correctly typed call is rejected, and the unresolved case is still surfaced rather than silently absorbed.

## 3. Demo and test-project coverage

- [x] 3.1 New bound methods in `test/pkg/object_arg_tests.go`: `TestObjectArgBaseClassParam(shape Shape2D)` called with a `CircleShape2D`, and `TestObjectArgReturnNode(node Node) Node` exercised by a typed GDScript return assignment.
- [x] 3.2 `test/demo/main.gd`: every object `var` in `test_object_args` is now statically typed (`var cs: CollisionShape2D = ...`, `var child: Node = ...`, etc.), the subclass call is wired in, and the stale workaround comment block that documented this bug is replaced with a note that the static types are themselves part of the assertion.
- [x] 3.3 Negative fixture `test/demo/type_rejection.gd`, not loaded by `main.gd`, passing a typed `Image` into the `CollisionShape2D`/`RefShape2D` parameters of `test_object_arg_set_shape`. Driven by `make check_type_rejection`, which asserts the engine rejects it and that **both** expected mismatches appear:
  - `argument 1 should be "CollisionShape2D" but is "Image"`
  - `argument 2 should be "Shape2D" but is "Image"` — the Ref pointee class, exactly as the spec requires
  - The target fails if the fixture ever parses cleanly, so a regression to `Object`-everything is caught rather than silently weakening the typing.

## 4. Verification

- [x] 4.1 `go build ./...` clean; `go vet ./pkg/... ./test/pkg/...` clean; `go test ./pkg/...` green.
- [x] 4.2 `make generate` produces no diff — no template changes, no codegen drift.
- [x] 4.3 `GODOT=/home/pcting/bin/godot make test` — **1093 assertions, 0 failures, 0 leaked instances**, exit 0, two consecutive identical runs, with no `Parse Error: Invalid argument` lines.
- [x] 4.4 Falsification. Restoring the old behaviour (advertise the owning class for OBJECT args) reproduces the original rejection across every typed call site: `argument 1 should be "Example" but is "CollisionShape2D"`, `... should be "Example" but is "CircleShape2D"`, including the new `test_object_arg_base_class_param`. `make` exit 2. This proves the typed declarations are genuinely gated by the metadata and are not passing for some unrelated reason.
- [x] 4.5 Dependency on `fix-ptrcall-object-arg-decode` confirmed resolved. This change was stashed when it turned out that landing it alone converted a parse-time rejection into a runtime `unsupported interface type` crash. With that change landed, the typed object calls parse **and** decode.
