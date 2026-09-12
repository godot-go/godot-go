# Overview

This will be a living doc which will provide an overview of key concepts in the godot-go bindings.

# GDScript Language Feature Mapping

There are a lot of language features supported by GDScript that does not map cleanly to Go.

## Default Parameter Value

Default parameters are supported. Register methods with `ClassDBBindMethod` passing a `defaultValues []Variant` slice; GDScript callers may omit trailing arguments.

## Class Inheritance

Go does not support classical class inheritance. Instead, composition with struct embedding is used in its place. Lets take a look at the following example user-defined class:

```go
type PlayerCharacter struct {
	CharacterBody2DImpl
}

// interface test evidence
var _ CharacterBody2D = &PlayerCharacter{}
```

The user-defined `PlayerCharacter` class extends the `CharacterBody2D` interface by embedding the `CharacterBody2DImpl` struct. Following the definition of `CharacterBody2DImpl` we see the following definition:

```go
type CharacterBody2DImpl struct {
	PhysicsBody2DImpl
}
```

We see that `CharacterBody2DImpl` embeds the `PhysicsBody2DImpl` struct, which implements the `PhysicsBody2D` interface.

## Virtual Methods

Go does not natively support virtual functions or struct methods. Instead, a method naming convention is implemented: qualified methods named `V_<ClassName>_<MethodName>` are registered as Godot virtual methods.

```go
func (e *Example) V_Example_Ready() { ... }

...

// register the function with Godot
ClassDBBindMethodVirtual(t, "V_Example_Ready", "_ready", nil, nil)
```

* `V_` denotes that this is a virtual function.
* `Example` is the Go type whose embedding level implements the virtual.
* `Ready` matches the `_ready` gdscript method (the leading underscore is dropped and the rest is PascalCased).

### Qualified names are required

The deprecated flat `V_<MethodName>` form (`V_Ready`) panics at registration time with a diagnostic naming the expected qualified shape. Because Go struct embedding promotes methods across the whole chain, flat names allowed only one implementation slot per virtual: shadowing made base implementations unreachable, and diamond embedding made selectors ambiguous.

### Most-derived-wins resolution

`ClassDBBindMethodVirtual` resolves the implementation by walking the registering type's embedding chain, most-derived type first, looking for `V_<LevelTypeName>_<MethodName>`. The first match binds:

```go
type PlayerCharacter struct {
	ControlImpl
	CharacterBody2DBase // declares V_CharacterBody2DBase_GetMaximumSize
}

func (p *PlayerCharacter) V_PlayerCharacter_GetMaximumSize() Vector2 { ... }
```

An instance of `PlayerCharacter` dispatches to `V_PlayerCharacter_GetMaximumSize`; an instance of the base level dispatches to `V_CharacterBody2DBase_GetMaximumSize`.

### Explicit delegation replaces super-calls

Because each level's implementation has a distinct name, a derived implementation can extend rather than replace the base behavior by calling it through the promoted selector:

```go
func (p *PlayerCharacter) V_PlayerCharacter_GetMaximumSize() Vector2 {
	base := p.CharacterBody2DBase.V_CharacterBody2DBase_GetMaximumSize()
	return base.Add(Vector2{X: 100})
}
```

Virtuals are invoked through the GDExtension `get_virtual_call_data2` / `call_virtual_with_data` path. Virtuals that no level along the chain implements return `nil` call data so Godot falls back to its engine default; implemented virtuals are dispatched to the resolved Go method.

### Generated virtual surface catalog

`extension_api.json` declares 1437 virtual methods across 106 classes. Code generation emits a declaration-only interface for each class with Category-A virtuals (non-void return and a plain wrapper on the same class hierarchy — the subset whose result the engine consumes through that wrapper):

```go
// pkg/gdclassimpl/virtuals.gen.go
type ControlVirtuals interface {
	V_Control_GetMaximumSize() Vector2
	V_Control_GetMinimumSize() Vector2
	V_Control_GetTooltip(at_position Vector2) String
	V_Control_GetCursorShape(at_position Vector2) int32
}
```

These interfaces are a catalog: greppable, godoc-visible, zero runtime cost. Nothing is registered or dispatched from them, and satisfying one changes no behavior. The checked-in classification of the full virtual surface (`cmd/generate/gdclassimpl/virtual_census.json`) is a codegen test fixture whose criterion is documented in the file itself; a codegen test fails loudly if the generated surface and the census drift.

### Compile-time signature verification

Go interface satisfaction is all-or-nothing, and the class-wide `<Class>Virtuals` interfaces carry the Godot class-level names, so they do not serve per-override checking. Instead, verify one overridden virtual's exact signature with a per-method anonymous interface:

```go
// Compiles only if TestVirtualsConformance declares that exact method.
var _ interface{ V_TestVirtualsConformance_GetMinimumSize() Vector2 } = (*TestVirtualsConformance)(nil)
```

Take the parameter and return types from the generated catalog (for `_get_minimum_size`, `ControlVirtuals` says `Vector2`), name the method per the qualified convention, and a mismatch surfaces at build time instead of registration time. See `test/pkg/virtuals_conformance.go`.

### Delegation to engine defaults is impossible

Once a virtual is registered, an override of an engine GDVIRTUAL replaces its behavior wholesale: Godot resolves virtual presence once per object instance into a cached member pointer (`gdvirtual.gen.h` — absence caches as an INVALID sentinel) and never re-consults, so the engine-default body is unreachable from GDExtension for that instance. A plain-wrapper call from inside the override (e.g. a `_get_maximum_size` override calling `GetMaximumSize()`) re-enters the registered implementation forever:

| Reference binding | Override mechanism | Engine-default reach |
|---|---|---|
| godot-cpp | C++ inheritance; compile-time override detection via `if constexpr (!std::is_same_v<decltype(&B::_x), decltype(&T::_x)>)` | none |
| godot-rust | `I<Class>` capability traits; unimplemented trait methods are never registered | none ("No access to `super` methods", book) |

The constraint is pinned, not just documented: `TestDelegationRepro` (`test/pkg/delegation_repro.go`) registers a `_get_maximum_size` override that delegates through the plain wrapper, caps the re-entry depth, and panics with a `delegation recursion confirmed` diagnostic; `make test_delegation_trap` runs it in a separate godot invocation and fails unless the process aborts non-zero with that diagnostic. To extend engine or base behavior, delegate explicitly to a named level implementation (see Explicit delegation replaces super-calls above) — do not call the plain wrapper from inside an override.

## Default Argument Values

Go does not support default parameter values in its syntax. Default argument values are instead passed through the `defaultValues` parameter of `ClassDBBindMethod` (and the `ClassDBBindMethodVirtual`/`ClassDBBindMethodVarargs` variants). GDScript callers can then omit trailing arguments.

## Call Error Reporting

When GDScript calls a bound method through the dynamic varcall path, the engine hands the call to Go together with a `GDExtensionCallError` slot. godot-go validates the call against the bound signature **before** invoking the method and reports a mismatch through that slot rather than aborting:

- **Too many arguments** → `GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS` (`expected` = declared count). Variadic (`varargs`) methods accept any argument count and are never rejected here.
- **Too few arguments** → `GDEXTENSION_CALL_ERROR_TOO_FEW_ARGUMENTS`, only when the call cannot be satisfied even after applying bound default arguments. The required count is derived from the same rule used to fill defaults, so validation and default application never disagree.
- **Non-convertible argument type** → `GDEXTENSION_CALL_ERROR_INVALID_ARGUMENT` with the failing argument index and the expected variant type. Convertibility uses the engine's `variant_can_convert_strict` predicate (the same one godot-cpp uses), so Godot's own valid conversions — numeric/string inter-conversion, `nil` into an object parameter — still pass; only genuinely non-convertible pairings are rejected.

A rejected call leaves a nil return, never runs the bound Go method body, and keeps the process alive; the engine surfaces the call error to the caller as normal. Only faults a caller cannot induce (absent method user data, null instance) remain fatal. This applies to the varcall path only: the `ptrcall` fast path has no error slot in its GDExtension typedef.

## Static Methods

Go does not support static methods in structs. Registering static methods is not supported.

## Static Variables

Go does not support static variables in structs. __(NOT YET IMPLEMENTED)__ Global variables can be registered as gdscript static variables.

## Packed Arrays

The generated `Packed*Array` types work and are partially tested in the tests. Conversion to Go native slices (e.g. `[]Vector2`) is not yet implemented.

## Coroutines

Go does not support coroutines; this means we do not have access to `await` (or `yield`). Without a coroutine alternative, a cumbersome pattern of chaining method calls will be required. __(NOT YET IMPLEMENTED)__ Instead, we have goroutines to wrap `signal` and `callable`.

## Built-in Types

### Basic Built-in Types

| GDScript Type | Go Type | Description |
| --- | --- | --- |
| `null` | `nil` | |
| `bool` | `bool` | |
| `int` | `int64` | All method parameters that use variations of `uint` and `int` will be converted to `int64` before passing over the value to Godot. |
| `float` | `float64` | `float32` will convert to `float64` before passing over the value to Godot. |
| `String` | `String` | There are helper functions to convert to go native `string`. |
| `StringName` | `StringName` | There are helper functions to convert to go native `string`. |
| `NodePath` | `NodePath` | |

### Vector Built-in Types

| GDScript Type | Go Type |
| --- | --- |
| `Vector2` | `Vector2` |
| `Vector2i` | `Vector2i` |
| `Rect2` | `Rect2` |
| `Vector3` | `Vector3` |
| `Vector3i` | `Vector3i` |
| `Transform2D` | `Transform2D` |
| `Plane` | `Plane` |
| `Quaternion` | `Quaternion` |
| `AABB` | `AABB` |
| `Basis` | `Basis` |
| `Transform3D` | `Transform3D` |

### Engine built-in Types

| GDScript Type | Go Type |
| --- | --- |
| `Color` | `Color` |
| `RID` | `RID` |
| `Object` | `Object` |

### Container Built-in Types

| GDScript Type | Go Type | Description |
| --- | --- | --- |
| `Array` | `Array` | `[]Variant`. |
| `PackedByteArray` | `PackedByteArray` | `[]byte`. |
| `PackedInt32Array` | `PackedInt32Array` | `[]int32`. |
| `PackedInt64Array` | `PackedInt64Array` | `[]int64`. |
| `PackedFloat32Array` | `PackedFloat32Array` | `[]float32`. |
| `PackedFloat64Array` | `PackedFloat64Array` | `[]float64`. |
| `PackedStringArray` | `PackedStringArray` | `[]string`. |
| `PackedVector2Array` | `PackedVector2Array` | `[]Vector2`. |
| `PackedVector3Array` | `PackedVector3Array` | `[]Vector3`. |
| `PackedVector4Array` | `PackedVector4Array` | `[]Vector4`. |
| `PackedColorArray` | `PackedColorArray` | `[]Color`. |
| `Dictionary` | `Dictionary` | No additional work needed. |
| `Signal` | `Signal` | No additional work needed. |
| `Callable` | `Callable` | No additional work needed. |
