# Tasks

## 1. Class-name resolution helper (pkg/core)

- [ ] 1.1 Add internal helper `godotClassNameForObjectType(t reflect.Type) (string, bool)` in `pkg/core` implementing the D2 resolution order: `GDNativeConstructors` plain name → `Internal.GDRegisteredGDClasses` plain name → `Ref`-prefix validated against `GDClassRefConstructors` → generic object interfaces (`Object`, `Ref`, `GDClass`, `GDExtensionClass`) fall back to `"Object"`.
- [ ] 1.2 Unit-test the helper in `pkg/core` (seed the registries directly with fixture entries): plain engine name, extension class name, `RefXxx` → `Xxx`, bare `Ref`/`Object` → `"Object"`, and an unresolvable OBJECT type returning `false`.

## 2. Wire resolution into the bind metadata

- [ ] 2.1 In `NewGoMethodMetadata`'s argument loop (`pkg/core/method_bind.go` ~lines 243–257), when `variantTypes[i] == GDEXTENSION_VARIANT_TYPE_OBJECT`, build `argPropClassNameStringNames[i]` from the resolved class name; on resolution failure `log.Panic` with the method name and unresolved Go type. Non-OBJECT arguments keep `className` unchanged.
- [ ] 2.2 In the return `PropertyInfo` construction (~lines 219–229), apply the same resolution when `returnType == GDEXTENSION_VARIANT_TYPE_OBJECT`; non-object returns keep `className` unchanged.
- [ ] 2.3 Extend `pkg/core` unit tests for `NewGoMethodMetadata`: object-typed arg (`Node`) advertises `"Node"`; refcounted arg (`RefShape2D`-shaped fixture) advertises the pointee class; object return advertises its own class; a scalar (`int64`) argument's `PropertyInfo` is unchanged from the pre-fix expectation; an unresolvable OBJECT argument panics at bind.

## 3. Demo and test-project coverage

- [ ] 3.1 In `test/pkg/example.go`, add bound test methods for typed GDScript round trips: a method taking a base class (`Shape2D`) that is called with a subclass instance (`CircleShape2D`), and a method returning an engine class (e.g. `Node`) whose advertised return class is asserted from GDScript via `get_method_list()` or typed assignment.
- [ ] 3.2 In `test/demo/main.gd`, replace the untyped object `var` declarations in `test_object_args` with statically-typed declarations (`var cs: CollisionShape2D = ...`, `var circle: CircleShape2D = ...`, etc.), wire in the new subclass-polymorphism call, and delete the stale workaround comment block (~lines 558–566) that documented the metadata bug.
- [ ] 3.3 Add a negative fixture script (e.g. `test/demo/type_rejection.gd`, not loaded by `main.gd`) that passes a wrong-class typed variable (`var img: Image` into a `CollisionShape2D` parameter) and verify with `GODOT=/home/pcting/bin/godot` check-only mode (or editor) that the engine reports the expected parse error; record the verification command in the PR description.

## 4. Verification

- [ ] 4.1 `go build ./...` and `go vet ./pkg/... ./test/pkg/...` clean; `go test ./pkg/core/` green.
- [ ] 4.2 `make generate` produces no diff (confirms no codegen drift; no template changes are expected).
- [ ] 4.3 `GODOT=/home/pcting/bin/godot make test` green, ignoring the known startup warnings (`get_variant_get_internal_ptr_func`, `callable_custom_get_user_data`, `classdb_register_extension_class_5`/`6`) and with no `Parse Error: Invalid argument` lines in the output.
