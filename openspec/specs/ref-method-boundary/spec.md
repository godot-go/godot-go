# ref-method-boundary Specification

## Purpose

Defines how reference-counted `Ref` values cross the boundary of reflectively bound Go methods on both dispatch paths: which Go signatures may express a reference, how arguments decode (class keying, ownership, nil), how mismatches fail, and what returns hand back to Godot.

## Requirements

### Requirement: Supported Ref signatures at bind time

Binding a method SHALL accept parameters and returns of the generated `RefXxx` interfaces and of bare `builtin.Ref`, and SHALL reject at bind time, with a message naming the offending Go type, the concrete reference forms: `RefXxxImpl` structs, `*RefXxxImpl` pointers, and `*RefBase[T]` pointers.

#### Scenario: Interface Ref forms bind successfully

- **WHEN** a method declares a `RefResource` parameter and a bare `builtin.Ref` return
- **THEN** registration succeeds and the method is callable from GDScript

#### Scenario: Concrete struct Ref form rejected at bind time

- **WHEN** a method declares a `RefResourceImpl` struct parameter
- **THEN** registration fails with an error naming that type, and no method is registered

#### Scenario: RefBase pointer form rejected at bind time

- **WHEN** a method declares a `*RefBase[Resource]` parameter
- **THEN** registration fails with an error naming that type, and no method is registered

### Requirement: Ref arguments decode keyed to the declared class

A `RefXxx` parameter SHALL decode an object argument of the declared class or any subclass into a wrapper of the declared `RefXxx` type (subclass sliced to the declared class). A bare `builtin.Ref` parameter SHALL accept any reference-counted object. Decoding SHALL behave identically through varcall and ptrcall.

#### Scenario: Subclass sliced to declared Ref class

- **WHEN** GDScript passes a custom `Resource` subclass instance to a `RefResource` parameter
- **THEN** the Go method runs and receives a `RefResource` value whose object is the passed instance

#### Scenario: Bare Ref accepts any refcounted object

- **WHEN** GDScript passes a `RectangleShape2D` to a bare `builtin.Ref` parameter
- **THEN** the Go method runs, and the value is valid with a non-nil referenced object

#### Scenario: Same decode through call()

- **WHEN** the same arguments arrive through `obj.call(...)` instead of a direct call
- **THEN** the decoded values satisfy the same conditions

### Requirement: Decoded Ref arguments own a reference

A Ref value decoded from a method argument SHALL hold its own engine reference, released deterministically when the call returns, so a Go method MAY use the argument for the duration of the call; a method that stores the argument past the call SHALL copy it (e.g. `NewRefAcquire`) to retain ownership. The suite SHALL report no leaked objects attributable to decoded arguments.

#### Scenario: Stored argument keeps the object alive

- **WHEN** a bound method copies its `RefResource` argument (via `NewRefAcquire`) and stores the copy, and a later frame reads the stored value
- **THEN** the referenced object is still alive and its contents are readable
- **AND** the suite shutdown reports no leaked engine object from this path

### Requirement: Nil reference arguments yield the zero Ref

A nil argument arriving where a `RefXxx` or bare `Ref` parameter is declared SHALL decode to the zero value of the declared type on both dispatch paths, with no engine call against a null pointer and no panic.

#### Scenario: Nil through varcall

- **WHEN** GDScript invokes the method through `call()` passing `null` for a `RefResource` parameter
- **THEN** the Go method runs and the parameter is invalid (`IsValid()` reports false)

#### Scenario: Null through ptrcall

- **WHEN** GDScript makes a direct typed call with a null `Resource` variable
- **THEN** the Go method runs and the parameter is the zero `RefResource`

### Requirement: Ref mismatch fails as a contained call error on varcall

When an argument arrives through varcall that is a non-reference-counted object, or an object whose class does not inherit the declared pointee class, the call SHALL be rejected before the Go method body runs, SHALL be reported to the engine without aborting the process, and later calls in the same session SHALL proceed normally.

#### Scenario: Non-refcounted object where Ref declared

- **WHEN** GDScript calls through `call()` passing a `Node` where a `RefResource` parameter is declared
- **THEN** the Go method body does not run, the failure is surfaced to GDScript, and the process survives with subsequent assertions still executing

#### Scenario: Unrelated refcounted class

- **WHEN** GDScript calls through `call()` passing an unrelated refcounted object (e.g. an `AudioStreamPlayback`) where `RefResource` is declared
- **THEN** the call is rejected in the same contained manner

### Requirement: Ref returns hand the same object to Godot

Returning a valid `RefXxx` or bare `Ref` value from a bound method SHALL give GDScript a reference to the same engine object (matching instance id), without altering the reference count beyond what the engine's own receiving code does. Returning a nil or invalid Ref SHALL give GDScript `null`.

#### Scenario: Ref echo preserves identity

- **WHEN** GDScript passes a `Resource` into a bound method that returns its `RefResource` parameter unchanged
- **THEN** the returned object has the same instance id as the input and is usable in GDScript

#### Scenario: Invalid Ref returns null

- **WHEN** a bound method returns a zero-value `RefResource`
- **THEN** GDScript receives `null` and no engine call dereferences a null pointer
