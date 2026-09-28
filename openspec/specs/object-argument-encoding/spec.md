## Purpose

Defines how a Go-side engine-class object — a plain wrapped class instance or a reference-counted smart pointer — is presented to Godot when it is passed as a call argument, so object-taking calls behave identically from Go and from Godot script.

## Requirements

### Requirement: Engine-Class Arguments Reach Godot As The Same Object

When a Go method passes a wrapped engine-class instance as an argument to a Godot call, Godot SHALL observe that same object, not a corrupted or unrelated pointer. Identity SHALL be preserved and observable from the receiving side.

#### Scenario: Node passed as an argument is the node Godot receives

- **WHEN** Go calls `AddChild` with a `CharacterBody2D` it holds
- **THEN** the child added to the tree is that same body, with the same object id
- **AND** the call returns normally with no crash

#### Scenario: Argument identity survives a round trip

- **WHEN** Godot calls a Go method with an object argument and that Go method passes the object back into another Godot call
- **THEN** the object Godot sees in the second call has the same instance id as the one it passed in

#### Scenario: Distinct arguments stay distinct

- **WHEN** Go passes two different objects as two arguments of the same call
- **THEN** Godot receives two distinct objects, each matching the argument Go intended, with no aliasing between them

### Requirement: Smart-Pointer Arguments Reach Godot As The Referenced Object

When a reference-counted pointer is passed as an argument, Godot SHALL receive the referenced object rather than the pointer wrapper's own storage.

#### Scenario: Shape assigned through a reference argument

- **WHEN** Go calls `SetShape` passing a reference to a `CircleShape2D` it created
- **THEN** the collision shape reports that circle as its shape, with the radius that was set

#### Scenario: Reference argument to a user-defined class

- **WHEN** Go passes a reference to a user-defined `RefCounted` subclass as an argument
- **THEN** Godot receives that object and can read its properties

### Requirement: Nil And Invalid Object Arguments Encode As Null

A nil interface value or an invalid (released or empty) reference passed as an object argument SHALL encode as a null object reference. The call SHALL NOT panic on the Go side and Godot SHALL see a null argument.

#### Scenario: Nil interface argument

- **WHEN** Go passes a nil value of an engine-class interface type as an argument
- **THEN** Godot receives a null object for that argument and no panic occurs

#### Scenario: Invalid reference argument

- **WHEN** Go passes a reference whose `IsValid()` reports false as an argument
- **THEN** Godot receives a null object for that argument and no panic occurs

### Requirement: Passing An Object Argument Does Not Change Its Lifetime

Passing an object as an argument SHALL NOT transfer ownership and SHALL NOT alter the object's reference count beyond what Godot's own receiving code does. Repeated calls SHALL NOT accumulate references or release ones still in use.

#### Scenario: Repeated object arguments do not grow the reference count

- **WHEN** Go passes the same refcounted object as an argument across many calls and never retains it
- **THEN** the object's reference count does not grow with call count

#### Scenario: Object remains usable after being passed

- **WHEN** Go passes an object as an argument and then continues to use that object
- **THEN** the object is still valid and its reference count reflects only references Go or Godot deliberately holds

### Requirement: Object Argument Encoding Is Uniform Across The Generated Surface

Every generated method that accepts an engine-class argument SHALL use the same argument encoding, so no object-taking call is a special case. Coverage SHALL span every engine-class argument type in the API, not only the types exercised by examples.

#### Scenario: Previously crashing calls become usable

- **WHEN** Go calls any generated method whose signature takes an engine-class argument, including `Node.AddChild` and `CollisionShape2D.SetShape`
- **THEN** the call completes without a memory-safety failure

#### Scenario: Value-type arguments are unaffected

- **WHEN** Go calls a generated method whose arguments are value types such as numbers, vectors, or matrices
- **THEN** those arguments encode exactly as before this change
