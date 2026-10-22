# Tasks

## 1. Variant mapping witnesses

- [x] 1.1 Extend `pkg/core/variant_reflect_type_test.go` with a table covering bool, every integer width, both float widths, native string, each Godot value type, containers, object wrappers, and a nil type; verify with `go test ./pkg/core -run TestReflectTypeToGDExtensionVariantType`.
- [x] 1.2 Add rejection cases for map, channel, function, `uintptr`, complex, unsafe pointer, unknown struct, unknown interface, concrete reference struct/pointer, `[]bool`, and `[4]float32`, asserting each panics; verify with the same `go test ./pkg/core` invocation.
- [x] 1.3 Cover `Variant` in the mapping table (`VARIANT_MAX`) and assert from the demo that a bound Variant return advertises nil plus the NIL_IS_VARIANT usage flag while a Variant parameter advertises nil and still dispatches; verify with `go test ./pkg/core` and `GODOT=/path/to/godot make test`.
- [x] 1.4 Confirm the existing 32-type matrix audit (`test/demo/main.gd` `test_variant_matrix_audit`) still covers every mapped type, adding any missing row; verify with `GODOT=/path/to/godot make test`.

## 2. Wrapper and cast witnesses

- [x] 2.1 Extend `pkg/builtin/wrapped_nil_ptr_test.go` with nil-receiver and ownerless cases for the object-pointer accessors, `IsNil`, and the base no-op `Destroy`; verify with `go test ./pkg/builtin`.
- [x] 2.2 Add a demo method in `test/pkg/example.go` that casts a `Control` instance to `Control` and to `Example`, asserts wrapper identity with the original pointer, casts to an incompatible class and asserts nil, and checks the inheritance predicate for `Control`, `Node`, and `Image`; verify the method compiles with `go build ./...`.
- [x] 2.3 Bind and call that method from `test/demo/main.gd` and assert its result; verify with `GODOT=/path/to/godot make test`.

## 3. Property access and vararg witnesses

- [x] 3.1 Add GDScript read/write round-trips for `dproperty_0`, `dproperty_1`, and `dproperty_2` in `test/demo/main.gd`, so the `_get`/`_set` handled path is asserted directly rather than only through the property list; verify with `GODOT=/path/to/godot make test`.
- [x] 3.2 Add a variadic demo method returning `(int, bool)`, bind it with `ClassDBBindMethodVarargs`, and assert from GDScript that only the integer is delivered; verify with `GODOT=/path/to/godot make test`.
- [x] 3.3 Add a package function declared `...Variant` bound with `ClassDBBindStaticMethod`, call it as `Example.<name>(...)` from GDScript at two argument counts, and assert both results; verify with `GODOT=/path/to/godot make test`.

## 4. Registration and bootstrap witnesses

- [x] 4.1 Add `pkg/core` unit tests for the bind-time gates: variadic Go method without the flag, vararg flag on a non-variadic method, more than two returns, a non-bool second return, and static-plus-virtual; verify with `go test ./pkg/core -run 'TestBind|TestStatic'`.
- [x] 4.2 Add a `pkg/core` unit test that seeds a `ClassInfo` with an existing method name and an existing virtual name and asserts the pre-allocation guard rejects both, then asserts an unregistered class is rejected; verify with `go test ./pkg/core`.
- [x] 4.3 Add a demo method that reads the registration registry for `Example`, asserts it is present, records the scene initialisation level, and constructs an instance which the object encoder accepts; bind and assert it from GDScript; verify with `GODOT=/path/to/godot make test`.

## 5. Documentation

- [x] 5.1 Add a "Binding substrate contracts" table to `docs/overview.md` mapping each of the six new capabilities to its spec path and to the public entry points it governs. The table links the post-archive locations (`openspec/specs/<capability>/spec.md`), so verify the sources exist with `ls openspec/changes/document-binding-substrate/specs/<capability>/spec.md`.
- [x] 5.2 Re-read each new spec against the code and correct any scenario that cannot be observed as written; verify `openspec validate "document-binding-substrate" --strict` passes with no warnings.

## 6. Integration verification

- [x] 6.1 Run `openspec validate "document-binding-substrate" --strict` and confirm the change is valid with all six capability deltas present.
- [x] 6.2 Run `go test ./pkg/core ./pkg/builtin` and `go build ./...`, and confirm they pass; the `pkg/builtin` run skips `TestRefAcquireUnrefReleasesExactlyOnce` and `TestRefAcquireFinalizerReleasesWhenUnreachable`, which crash without a running engine (see design.md, Verification Baseline).
- [x] 6.3 Run `GODOT=/path/to/godot make test` and confirm both the runtime suite and the editor-plugin gate pass with no new startup warnings.

## Workflow follow-up

- Review the six specs against the boundary table in `design.md` before archiving.
- Archive the change; confirm the main spec count moves from 41 to 47 and each new spec carries its Purpose.
