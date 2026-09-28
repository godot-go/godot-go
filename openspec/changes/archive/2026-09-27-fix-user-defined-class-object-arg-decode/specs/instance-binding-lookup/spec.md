# Spec Delta

## Purpose

Governs how the GDExtension instance-binding slot is read and decoded so that any object crossing Godot→Go — a plain engine-class instance or a user-defined extension-class instance — resolves to its Go wrapper, and a lookup that cannot resolve fails with a clear typed error instead of a crash.

## ADDED Requirements

### Requirement: Binding Getter Decodes The Registered Shape
The instance-binding getter SHALL interpret the `void*` returned by `object_get_instance_binding` using the same shape every registration site stores — a `cgo.Handle` value — and SHALL NOT cast the returned pointer to a Go interface pointer and dereference it.

#### Scenario: Returned binding is decoded as a handle
- **WHEN** the lookup receives a non-null binding pointer from the engine
- **THEN** it SHALL resolve the value via `cgo.Handle` and return the `Object` it wraps
- **AND** it SHALL NOT dereference the raw pointer as `*Object`

#### Scenario: Lookup token matches the registration token
- **WHEN** the getter calls `object_get_instance_binding`
- **THEN** it SHALL pass the same library token (`FFI.Token`) used by `object_set_instance_binding` at every registration site
- **AND** a binding registered under a different token SHALL NOT be resolved through this lookup

### Requirement: User-Defined Class Instance Resolves To Its Go Wrapper
When an instance of a user-defined extension class created by this extension is passed as an object argument, the varcall decode SHALL return the Go wrapper registered for that instance at construction, not a newly constructed object and not a crash.

#### Scenario: User-defined instance passed as a plain engine-class argument
- **WHEN** GDScript passes a `TestHierarchicalDerived` instance into a Go method whose parameter is a plain engine class such as `Node`
- **THEN** the decoded argument is the original Go wrapper for that instance, with the same Godot instance id
- **AND** the process does not segfault

#### Scenario: Wrapper identity preserved across calls
- **WHEN** the same user-defined instance is passed into two successive Go method calls
- **THEN** both calls resolve to the identical Go wrapper value

### Requirement: Engine-Class Instances Resolve As They Do Today
An engine-class object with no existing binding SHALL continue to resolve through the binding create callback, and that callback SHALL return the canonical `cgo.Handle` shape so the getter can decode it.

#### Scenario: Engine-class object argument still resolves after the shape change
- **WHEN** GDScript passes a plain engine-class object (e.g. a `Node2D` created in GDScript) into a Go method taking that engine class
- **THEN** the binding create callback produces a `cgo.Handle`-shaped binding
- **AND** the decode yields a working Go wrapper for that object, as today

### Requirement: Unresolvable Lookup Fails Typed
A lookup that cannot resolve to a live Go wrapper SHALL return a typed failure carrying the class name and the failure reason, and the varcall path SHALL report it through the existing varcall error reporting; it SHALL NOT panic, SIGSEGV, or silently return a dangling value.

#### Scenario: Null binding with no creatable callbacks
- **WHEN** the engine returns a null binding and the object's class has no registered binding callbacks, or the create callback yields no wrapper
- **THEN** the lookup returns a typed error naming the class
- **AND** the rejected varcall is reported to the engine without aborting the process

#### Scenario: Non-handle binding value is rejected
- **WHEN** the returned binding pointer cannot be resolved through `cgo.Handle` to a non-nil `Object`
- **THEN** the lookup returns a typed error instead of dereferencing the pointer
