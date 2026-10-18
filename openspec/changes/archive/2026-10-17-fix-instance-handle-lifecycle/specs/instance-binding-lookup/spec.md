# instance-binding-lookup Delta

## MODIFIED Requirements

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

## ADDED Requirements

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
