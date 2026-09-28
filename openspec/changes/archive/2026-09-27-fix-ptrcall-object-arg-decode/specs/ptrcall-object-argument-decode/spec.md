# Spec Delta

## Purpose

Governs how an object-typed Go interface parameter is decoded when Godot dispatches a call into Go through ptrcall, so engine-class objects, subclass instances and null arguments all resolve to a usable Go value instead of a panic.

## ADDED Requirements

### Requirement: Object-Typed Interface Parameter Decodes By Interface Check

When a bound method's parameter is a Go interface implementing `Object` and the call arrives through ptrcall, the decoder SHALL select the object branch by testing the declared type with `reflect.Type.Implements`, and SHALL resolve the argument to a Go wrapper around the supplied engine object.

#### Scenario: Plain engine-class parameter receives its object

- **WHEN** a bound method declares a parameter of Go interface type `Node` and Godot dispatches the call through ptrcall with a `Node` object pointer
- **THEN** the decoder produces a Go value satisfying `Node` that wraps the supplied engine object, and no panic occurs

#### Scenario: Dispatch does not depend on a zero value's dynamic type

- **WHEN** the decoder inspects a declared interface type to choose its branch
- **THEN** the choice is made from `reflect.Type.Implements` against the interface types, and no branch selection depends on the dynamic type of a zero value of that interface

### Requirement: Subclass Instances Satisfy A Base-Class Parameter

An object argument whose actual engine class is a subclass of the declared parameter's class SHALL decode successfully, using the argument's runtime class name rather than the declared parameter name to locate the constructor.

#### Scenario: Subclass passed where the base class is declared

- **WHEN** a bound method declares a parameter of Go interface type `Shape2D` and is called with a `CircleShape2D`
- **THEN** the decoder resolves the constructor for `CircleShape2D`, the Go method body runs, and the value it receives reports its class as `CircleShape2D`

### Requirement: Null Object Argument Yields The Zero Interface Value

When the engine supplies a null object pointer for an object-typed parameter, the decoder SHALL assign the zero value of the declared interface and SHALL NOT issue an engine call that dereferences the null pointer.

#### Scenario: Null passed to an object parameter

- **WHEN** a bound method declares a parameter of Go interface type `Node` and Godot passes a null object pointer
- **THEN** the Go method body runs with a nil `Node`, no `object_get_class_name` call is made against the null pointer, and no panic occurs

### Requirement: Ref-Typed Interface Parameters Keep Existing Behaviour

A parameter whose declared interface implements `Ref` SHALL continue to decode through the reference branch, deriving the pointee class name from the declared type and producing the corresponding `Ref` wrapper.

#### Scenario: Ref parameter still decodes after the dispatch change

- **WHEN** a bound method declares a parameter of Go interface type `RefShape2D` and the call arrives through ptrcall
- **THEN** the decoder resolves pointee class `Shape2D` and supplies a `RefShape2D` value, unchanged from current behaviour

### Requirement: Undecodable Interface Fails Loudly

An object-position parameter that implements neither `Object` nor `Ref` SHALL fail with a message naming the argument index and the offending type.

#### Scenario: Interface that is neither object nor reference

- **WHEN** the decoder reaches an interface parameter that implements neither `Object` nor `Ref`
- **THEN** the call fails with a message including the argument index and the Go type name
