# instance-binding-lookup Specification

## Purpose
Governs how the GDExtension instance-binding slot is read and decoded so that any object crossing Godot→Go — a plain engine-class instance or a user-defined extension-class instance — resolves to its Go wrapper, and a lookup that cannot resolve fails with a clear typed error instead of a crash.

## Requirements

### Requirement: Binding Getter Decodes The Registered Shape
The instance-binding getter SHALL interpret the binding-slot value returned by `object_get_instance_binding` — crossed as `uintptr_t` and presented as a `cgo.Handle` — using the same shape every registration site stores, and SHALL NOT cast the value to a Go interface pointer and dereference it.

#### Scenario: Returned binding is decoded as a handle
- **WHEN** the lookup receives a non-null binding value from the engine
- **THEN** it SHALL resolve the value via `cgo.Handle` and return the `Object` it wraps
- **AND** it SHALL NOT dereference the value as `*Object`

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

### Requirement: Binding Handles Have Exactly One Release Owner
Every `cgo.Handle` stored in an instance-binding slot SHALL be deleted exactly once, by the engine's free notification for that binding shape, and no other path SHALL delete it.

#### Scenario: Engine-class binding handle freed with the binding
- **WHEN** the engine frees an instance binding that was created by the engine-class binding create callback (the object is being destroyed or the binding is explicitly released)
- **THEN** the binding free callback for that shape SHALL delete the handle stored in the binding slot
- **AND** the wrapper reachable through that handle becomes collectable by Go garbage collection

#### Scenario: User-defined class binding slot is not deleted independently
- **WHEN** the engine invokes the binding free callback for this extension's own library token on a user-defined extension class instance
- **THEN** the callback SHALL NOT delete the binding slot value, because that slot holds the same handle already owned and deleted by the instance-free callback for the same object

#### Scenario: Lookup after release fails typed
- **WHEN** a binding lookup receives a slot whose handle has been deleted
- **THEN** the lookup SHALL return the existing typed binding-lookup error naming the dead handle
- **AND** the process SHALL NOT panic or crash

### Requirement: Handle Slot Values Cross the C Boundary as Integers
A packed `cgo.Handle` value that this extension stores in a binding, extension-instance, or userdata slot SHALL cross the C boundary only through integer-typed storage (`uintptr_t` C shim parameters and results, `cgo.Handle`/`uintptr` Go values), never through GC pointer-typed slots that remain live across a garbage-collection stack scan.

#### Scenario: Collection during a handle-carrying crossing does not fault
- **WHEN** a garbage-collection cycle scans stacks while a goroutine is parked in a C call whose argument or result carries a packed handle value
- **THEN** the runtime SHALL NOT abort with "invalid pointer found on stack"
- **AND** the instance-churn and binding-churn fixtures SHALL complete with zero runtime fatal errors

#### Scenario: Binding getter reinterprets the slot value as a handle
- **WHEN** a generated binding-getter wrapper returns a binding-slot value
- **THEN** the C shim SHALL have returned it as `uintptr_t` and the Go wrapper SHALL present it as a `cgo.Handle`
- **AND** no `unsafe.Pointer` reinterpretation of the small integer value SHALL appear on the API surface

### Requirement: Generated object returns resolve to the runtime-class wrapper
A generated method that returns a non-reference-counted engine object SHALL return the wrapper bound for the returned object's runtime class when one resolves and satisfies the method's declared return type, not a new wrapper typed to that declared class. A reference-counted return keeps its declared-class owned wrapper, because that wrapper owns the reference the return slot transfers.

#### Scenario: Narrowing a returned object succeeds
- **WHEN** a generated method declares a base-class return such as `Node` and the engine returns an object whose runtime class is `Label`
- **THEN** the returned wrapper stands for the `Label` runtime class
- **AND** a Go type assertion of the returned value to `Label` succeeds
- **AND** the caller obtains it with that assertion alone, without calling a library helper or re-resolving the binding
- **AND** the method's declared return type is unchanged, so existing callers compile as written

#### Scenario: Repeated returns yield the same wrapper
- **WHEN** two generated calls return the same engine object
- **THEN** both calls return the identical Go wrapper value

#### Scenario: Widening still works on a returned object
- **WHEN** a generated method returns an object whose runtime class is a subclass of the declared return class
- **THEN** the returned wrapper satisfies the declared return type
- **AND** it also satisfies every ancestor class interface

#### Scenario: Null engine return stays a null wrapper
- **WHEN** the generated method's return slot holds a null object pointer
- **THEN** the returned wrapper carries no engine object and no engine binding lookup is issued

### Requirement: Unresolvable return binding falls back to the declared class
When no wrapper can be resolved through the instance binding for a returned engine object — the class has no registered binding callbacks, or the engine declines to create a binding — or the resolved wrapper does not satisfy the declared return type, the generated method SHALL return a wrapper built over its declared return class instead of failing the call.

#### Scenario: Class without binding callbacks still returns a usable wrapper
- **WHEN** a generated object return resolves to an engine object whose class has no registered binding callbacks
- **THEN** the method returns a wrapper over the declared return class
- **AND** the call does not panic or return a nil wrapper in place of a non-null object

#### Scenario: Wrapper that fails the declared type is not returned
- **WHEN** the instance binding resolves to a wrapper that is not an instance of the method's declared return class
- **THEN** the method returns a wrapper over the declared return class
- **AND** the caller never receives a wrapper that fails the declared return type

#### Scenario: Returned fallback wrapper still reaches the engine object
- **WHEN** a caller invokes an engine method through a fallback wrapper
- **THEN** the call reaches the object the return slot carried
