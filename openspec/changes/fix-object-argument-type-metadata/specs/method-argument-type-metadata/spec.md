# Spec Delta

## Purpose

Governs the Godot class name a Go-bound method advertises in the `PropertyInfo.class_name` of its object-typed parameters and return value, so GDScript static type checking accepts correct argument classes (including subclasses) and rejects wrong ones.

## ADDED Requirements

### Requirement: Object-Typed Parameter Advertises Its Own Godot Class

When a bound method's parameter has the OBJECT variant type, the parameter's `PropertyInfo.class_name` SHALL be the Godot class of the parameter's own Go type — not the registering (owning) class. A plain engine-class Go interface `Node` SHALL advertise `"Node"`; a smart-pointer Go interface `RefShape2D` SHALL advertise `"Shape2D"`.

#### Scenario: Plain engine-class parameter advertises its class

- **WHEN** a method bound on class `Example` declares a parameter of Go interface type `Node`
- **THEN** that parameter's registered `PropertyInfo.class_name` is `"Node"` and its variant type is OBJECT

#### Scenario: Refcounted smart-pointer parameter advertises the pointee class

- **WHEN** a method bound on class `Example` declares a parameter of Go interface type `RefShape2D`
- **THEN** that parameter's registered `PropertyInfo.class_name` is `"Shape2D"`, not `"RefShape2D"` and not `"Example"`

### Requirement: Correctly Typed GDScript Arguments Pass Static Checking

A GDScript call that passes a statically-typed variable whose class matches the advertised parameter class SHALL pass the engine's parse-time type check and dispatch to the Go method.

#### Scenario: Typed variable of the declared class is accepted

- **WHEN** GDScript declares `var cs: CollisionShape2D = CollisionShape2D.new()` and calls a bound method whose parameter advertises class `CollisionShape2D`
- **THEN** the script parses without error and the Go method receives the object

### Requirement: Wrong-Class Typed GDScript Arguments Are Rejected By The Engine

A GDScript call that passes a statically-typed variable whose class is not compatible with the advertised parameter class SHALL be rejected by the engine with a parse error, and the Go method SHALL NOT be invoked.

#### Scenario: Wrong class fails static checking

- **WHEN** GDScript declares `var img: Image = Image.new()` and calls a bound method whose parameter advertises class `CollisionShape2D`
- **THEN** the engine reports a parse error naming the expected and actual classes and the call never reaches Go

### Requirement: Subclass Instances Are Accepted Where Godot Allows Polymorphism

A GDScript call that passes a statically-typed variable of a subclass of the advertised parameter class SHALL pass static checking, matching Godot's native polymorphic acceptance.

#### Scenario: Subclass satisfies the base-class parameter

- **WHEN** GDScript declares `var circle: CircleShape2D = CircleShape2D.new()` and calls a bound method whose parameter advertises the base class `Shape2D`
- **THEN** the script parses without error and the Go method receives the `CircleShape2D` instance

### Requirement: Value-Type Parameter Metadata Is Unchanged

Parameters whose variant type is not OBJECT (bools, ints, floats, strings, vectors, containers, `Variant`) SHALL keep the `PropertyInfo` metadata produced today; this change SHALL NOT alter their class name, name, hint, or variant type.

#### Scenario: Scalar parameter keeps existing metadata

- **WHEN** a method bound on class `Example` declares an `int64` parameter
- **THEN** the parameter's registered `PropertyInfo` is byte-for-byte equivalent to the pre-change registration (variant type INT, same class name and hint as before)

### Requirement: Object-Typed Return Advertises Its Own Godot Class

When a bound method's return has the OBJECT variant type, the return `PropertyInfo.class_name` SHALL be the Godot class of the method's Go return type, resolved the same way as parameters.

#### Scenario: Node-returning method advertises Node

- **WHEN** a method bound on class `Example` declares a Go return type of interface `Node`
- **THEN** the return `PropertyInfo.class_name` is `"Node"`, not `"Example"`

#### Scenario: Ref-returning method advertises the pointee class

- **WHEN** a method bound on class `Example` declares a Go return type of interface `RefImage`
- **THEN** the return `PropertyInfo.class_name` is `"Image"`

### Requirement: Unresolvable Object Types Fail Loudly At Bind Time

If a parameter or return resolves to the OBJECT variant type but its Go type name matches no registered Godot class (in `GDNativeConstructors`, `GDClassRefConstructors`, or the registered extension classes), and it is not a generic object interface that falls back to `"Object"`, the bind SHALL fail with a panic identifying the method and the unresolvable Go type, rather than silently advertising the owning class.

#### Scenario: Unknown object-typed argument panics at bind

- **WHEN** a method is bound with an OBJECT-typed parameter whose Go interface name resolves to no registered class and is not a generic object interface
- **THEN** binding panics with the method name and the unresolved Go type name, and no method registration occurs
