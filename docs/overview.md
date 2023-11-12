# Overview

This will be a living doc which will provide an overview of key concepts in the godot-go bidnings.

# GDScript Language Feature Mapping

There are a lot of language features supported by GDScript that does not map cleanly to Go.

## Default Parameter Value

Default parameters are currently not supported.

## Class Inheritance

Go does not support classical class inheritance. Instead, composition with struct embedding is used in it's place. Lets take a look at the following example user-defined class:

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

Go does not natively support virtual functions or struct methods. Insteead, a method name prefix convention will be implemented. The current implementation ignores all virtual methods on existing Godot classes.

```go
func (e *Example) V_Ready() { ... }

...

// register the function with Godot
ClassDBBindMethodVirtual(t, "V_Ready", "_ready", nil, nil)
```

__(NOT YET IMPLEMENTED)__ The eventual best practice will be the following example:

```go
func (e *Example) V_Example_Ready() { ... }

...

// register the function with Godot
ClassDBBindMethodVirtual(t, "V_Example_Ready", "_ready", nil, nil)
```

* `V_` denotes this this is a virtual function.
* `Example_` matches the name of the class. godot-go should panic if the registered method does not follow this pattern.
* `Ready` matches `_ready` gdscript method.

## Default Argument Values

Go does not support default parameter values. Default argument will show up in the godocs comments, but it will not be implemented directly in the code.

## Static Methods

Go does not support static methods in structs. Registering static methods is not supported.

## Static Variables

Go does not support static variables in structs. __(NOT YET IMPLEMENTED)__ Global variables can be registered as gdscript static variables.

## Packed Arrays

Works fine and partially tested in the tests.

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
| `Array` | `Array` | __(NOT YET IMPLEMENTED)__ `[]Variant`. |
| `PackedByteArray` | `PackedByteArray` | __(NOT YET IMPLEMENTED)__ `[]byte`. |
| `PackedInt32Array` | `PackedInt32Array` | __(NOT YET IMPLEMENTED)__ `[]int32`. |
| `PackedInt64Array` | `PackedInt64Array` | __(NOT YET IMPLEMENTED)__ `[]int64`. |
| `PackedFloat32Array` | `PackedFloat32Array` | __(NOT YET IMPLEMENTED)__ `[]float32`. |
| `PackedFloat64Array` | `PackedFloat64Array` | __(NOT YET IMPLEMENTED)__ `[]float64`. |
| `PackedStringArray` | `PackedStringArray` | __(NOT YET IMPLEMENTED)__ `[]string`. |
| `PackedVector2Array` | `PackedVector2Array` | __(NOT YET IMPLEMENTED)__ `[]Vector2`. |
| `PackedVector3Array` | `PackedVector3Array` | __(NOT YET IMPLEMENTED)__ `[]Vector3`. |
| `PackedColorArray` | `PackedColorArray` | __(NOT YET IMPLEMENTED)__ `[]color`. |
| `Dictionary` | `Dictionary` | No additioanl work needed. |
| `Signal` | `Signal` | No additioanl work needed. |
| `Callable` | `Callable` | No additioanl work needed. |
