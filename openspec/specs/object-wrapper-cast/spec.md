## Purpose

Defines the base value every Go wrapper of a Godot object provides — owner
pointer, nil reporting, and destruction — and how one wrapper is cast to another
engine class through the engine's class tags.

## Requirements

### Requirement: A wrapper exposes its engine owner and a nil-safe object pointer

Every object-carrying wrapper SHALL hold the engine object it stands for and
SHALL make that owner readable. The object-pointer accessors SHALL return a null
pointer for a nil wrapper or a wrapper with no owner, so callers can pass a nil or
typed-nil wrapper into a call slot and the engine reads null.

#### Scenario: Nil wrapper passed as an argument

- **WHEN** a nil wrapper is passed where an object argument is expected
- **THEN** the call slot contains a null object pointer

#### Scenario: Base wrapper with no owner

- **WHEN** a base wrapper was never given an owner
- **THEN** its object-pointer accessors return null and its nil check reports true

### Requirement: The type-pointer accessors expose the owner slot

The type-pointer accessors SHALL address the wrapper's stored owner slot so the
engine can read or write through it; they require a non-nil wrapper and are not
part of the nil-safe contract.

#### Scenario: Engine reads through the owner slot

- **WHEN** a bound call passes a wrapper through a type-pointer slot
- **THEN** the engine reads the stored owner pointer from that slot

### Requirement: The base wrapper's destruction is a no-op

Destroying a base wrapper SHALL do nothing; reference release and engine
destruction belong to the generated class implementations and to the instance
lifecycle, not to the base wrapper.

#### Scenario: Base wrapper destroyed

- **WHEN** `Destroy` is called on a base wrapper
- **THEN** no engine call is made and the owner pointer is not changed

### Requirement: A class-instance wrapper delegates identity to its wrapped instance

The wrapper used for registered class instances SHALL report the class name, the
parent class name, and the owner by delegating to the instance it wraps, and its
object-pointer accessors SHALL be nil-safe for a nil wrapper or a wrapper with no
wrapped instance. Its nil check follows the wrapped instance and therefore is not
nil-safe when no instance is carried at all.

#### Scenario: Registered instance reports identity

- **WHEN** a property or notification callback holds the class-instance wrapper
- **THEN** asking for its class and parent yields the registered class's names

#### Scenario: Wrapper with no instance

- **WHEN** the class-instance wrapper carries no instance
- **THEN** its object-pointer accessors return null rather than panicking

### Requirement: Casting uses the engine's class tag and resolves through the instance binding

A cast of a non-nil wrapper SHALL look up the target class's engine class tag,
perform an inheritance-aware engine cast, and resolve the cast object through this
library's instance binding. A failed cast SHALL yield nil; a successful cast whose
binding cannot be resolved SHALL yield nil with a warning; a null wrapper SHALL
yield nil.

#### Scenario: Cast to a valid class

- **WHEN** a wrapper for a `Node` that is a `Sprite2D` is cast to `Sprite2D`
- **THEN** the result is the same Go wrapper that stands for that engine object

#### Scenario: Cast of an object not yet wrapped

- **WHEN** a cast succeeds and no wrapper is bound to the cast object yet
- **THEN** the engine creates the binding from the callbacks registered for that object's class

#### Scenario: Cast to an incompatible class

- **WHEN** a wrapper for an `Image` is cast to a `Node`
- **THEN** the result is nil

#### Scenario: Null wrapper

- **WHEN** a nil wrapper is cast to any class
- **THEN** the result is nil without an engine call

#### Scenario: Target class unknown to the engine

- **WHEN** a wrapper is cast to a class name for which the engine publishes no class tag
- **THEN** the cast fails loudly instead of returning a wrong wrapper

### Requirement: Inheritance checks accept the class and its subclasses

An inheritance check SHALL report true when the object is an instance of the named
class or of a subclass of it, using the engine's class tag and cast; a null object,
empty class name, unknown class, or failed cast SHALL report false.

#### Scenario: Subclass satisfies the check

- **WHEN** a `Sprite2D` instance is checked against `Node2D`
- **THEN** the check reports true

#### Scenario: Unrelated class fails the check

- **WHEN** an `Image` instance is checked against `Node`
- **THEN** the check reports false

#### Scenario: Null or unknown input

- **WHEN** the object is null, the class name is empty, or the class name is unknown
- **THEN** the check reports false without an engine call that can crash
