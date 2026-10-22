# Spec Delta

## Purpose

Defines the contract for turning a Go type into a Godot extension class: the
naming and embedding invariants checked before registration, the creation-info
callbacks installed for it, the bookkeeping that makes it constructible and
encodable, the method namespace rules, and the teardown performed when it is
unregistered.

## ADDED Requirements

### Requirement: Registered class identity must match the Go type's shape

A class SHALL be registrable only when its reported Godot class name equals the
Go struct's name and its first embedded field is the parent class's generated
`Impl` type, because construction and inherited-method resolution both rely on
that shape.

#### Scenario: Matching name and embedding

- **WHEN** a Go struct named `Player` embeds `CharacterBody2DImpl` and reports `Player` as its class and `CharacterBody2D` as its parent
- **THEN** registration proceeds

#### Scenario: Class name differs from struct name

- **WHEN** a Go struct named `Player` reports `PlayerCharacter` as its class name
- **THEN** registration fails before any engine call

#### Scenario: First embedded field is not the parent's Impl

- **WHEN** a `Player` struct whose first field is not `CharacterBody2DImpl` reports `CharacterBody2D` as its parent
- **THEN** registration fails before any engine call

#### Scenario: Parent equals the class

- **WHEN** a type reports the same name as both class and parent
- **THEN** registration fails before any engine call

### Requirement: The parent class must already be known

A class SHALL be registrable only when its parent has engine constructors
registered, since instance creation constructs the parent object first.

#### Scenario: Unknown parent

- **WHEN** a class reports a parent for which no engine constructor exists
- **THEN** registration fails before the engine registration call

### Requirement: Registration installs a concrete, exposed, non-virtual class

A registered class SHALL be presented to the engine as concrete, exposed, and
neither virtual nor runtime, with a generated icon path of none and the complete
set of creation-info callbacks this binding uses for property access, property
listing, revert, validation, notification, string conversion, instance
creation/free, and virtual dispatch.

#### Scenario: Registered class is script-visible

- **WHEN** a class is registered and GDScript asks the engine for it
- **THEN** the class is visible as a concrete, instantiable extension class

#### Scenario: Creation routes through the binding

- **WHEN** the engine instantiates the registered class or queries its properties, notifications, or string form
- **THEN** each callback reaches the binding rather than a nil slot

### Requirement: Registration is recorded at the level being initialised

The registration SHALL record the initialisation level that is currently being
delivered, defaulting to the core level before any level is delivered, so that
level-scoped teardown can find it.

#### Scenario: Registered during a scene initializer

- **WHEN** a class is registered while the scene level is being delivered
- **THEN** its registration records the scene level

### Requirement: Registration makes the class constructible, encodable, and binding-resolvable

A successful registration SHALL install the class's Go constructor, its argument
encoder, and its instance-binding callbacks, and SHALL add it to the registration
registry before the engine registration call, so that construction, argument
encoding, and binding lookups work immediately.

#### Scenario: Create after register

- **WHEN** a registered class is instantiated from Go immediately after its registration call
- **THEN** the instance is constructed through the registered constructor and appears in the registration registry

#### Scenario: Encode after register

- **WHEN** an instance of a registered class is passed as an object argument
- **THEN** the object encoder installed at registration handles it

### Requirement: The method namespace rejects collisions before engine registration

Binding a method or virtual SHALL fail when its GDScript name is already bound as
a method or as a virtual on that class, and SHALL fail when the class is not
registered, before any engine-visible metadata is allocated.

#### Scenario: Duplicate method name

- **WHEN** a second method is bound under a GDScript name already used on the class
- **THEN** the bind fails and no engine method is registered

#### Scenario: Method name already used by a virtual

- **WHEN** a method is bound under a name already registered as a virtual
- **THEN** the bind fails

#### Scenario: Unregistered class

- **WHEN** a method is bound for a class with no registration
- **THEN** the bind fails loudly

### Requirement: Unregistration releases class metadata exactly once

Unregistration SHALL remove the class from the registry, ask the engine to
unregister it, release every virtual and method metadata object, and then release
the class-owned property backing store, retained signal names, and cached class
and parent names exactly once; unregistering an unknown class SHALL fail loudly.

#### Scenario: Unregister a registered class

- **WHEN** a registered class is unregistered
- **THEN** the engine no longer reports it
- **AND** its property strings, signal names, and cached names are released once

#### Scenario: Unregister twice

- **WHEN** an already-unregistered class is unregistered again
- **THEN** the call fails loudly rather than double-freeing its metadata
