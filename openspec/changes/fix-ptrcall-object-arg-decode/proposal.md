# Proposal

## Why

The ptrcall argument decoder cannot decode an object-typed Go interface. Any Godot→Go call that arrives through ptrcall with an object argument panics:

```
panic: unsupported interface type {"arg_index": 0, "type": "builtin.Node"}
  at pkg/core/method_bind_reflect.go:652
```

Root cause, proven rather than inferred. The interface branch dispatches with

```go
v := reflect.Zero(t)
inst := v.Interface()
switch inst.(type) {
case Object:   // the correct logic lives here
```

For an interface `t`, `reflect.Zero(t)` is the zero **interface**, so `.Interface()` returns a nil `interface{}`. A Go type switch matches a concrete case only against a non-nil dynamic type, so `case Object:` can never be taken and control falls through to `default`, which handles only `Ref`-implementing types and then panics. Verified in isolation: `reflect.Zero[Node]().Interface() == nil` is `true`, `matches(Object)` is `false`, while `reflect.TypeOf[Node]().Implements(Object)` is `true`.

The correct logic already exists inside that unreachable `case Object:` — read the runtime class name from the engine, look up the constructor, build the wrapper from the owner pointer. It has simply never executed. The same file already uses the working form elsewhere (`switch { case t.Implements(gdClassType): … }` at ~1076), so the broken shape is the anomaly, not the convention.

**Why this has stayed hidden:** the object-argument metadata defect makes every typed GDScript object call fail at parse time, so no typed object argument ever reached ptrcall, and untyped calls arrive as varargs through the varcall decoder instead. This is therefore a prerequisite for `fix-object-argument-type-metadata`: landing that change alone converts a parse-time rejection into a runtime crash. Isolated by disabling both new test methods added while investigating — the pre-existing `var child: Node` plus `test_object_arg_add_child(child)` still crashes.

## What Changes

- Replace the zero-value type switch in the `reflect.Interface` case with the explicit `t.Implements(...)` dispatch already used elsewhere in this file: `gdObjectType` first, then `refType`, then the panic.
- Guard the object branch against a null object pointer, which the unreachable code never had to consider: a null argument yields the zero value of the declared interface rather than a `get_class_name` call on null.
- Preserve the existing `Ref` branch behaviour unchanged, including its class-name derivation.

## Capabilities

### New Capabilities

- `ptrcall-object-argument-decode`: how an object-typed Go interface parameter is decoded when Godot dispatches a call through ptrcall, covering engine classes, subclass polymorphism, and null.

### Modified Capabilities

(none)

## Impact

- `pkg/core/method_bind_reflect.go` — the `reflect.Interface` case of `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`.
- `test/pkg/*` and `test/demo/main.gd` — typed object-argument tests that currently crash.
- Ordering: must land **before** `fix-object-argument-type-metadata`, whose spec scenario "the Go method receives the `CircleShape2D` instance" is unverifiable until this is fixed.

## Non-goals

- No change to the instance-binding shape (`cgo.Handle` vs `&inst`) or to user-defined extension classes arriving as object arguments. That is `fix-user-defined-class-object-arg-decode`, which owns `Variant.ToObject` and the binding callbacks.
- No change to `PropertyInfo.class_name` advertisement. That is `fix-object-argument-type-metadata`.
- No change to outbound Go→Godot object-argument encoding (landed in `e727e39`).
- No hand-editing of `*.gen.*` files.
