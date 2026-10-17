# class-instance-lifecycle Delta

## MODIFIED Requirements

### Requirement: cgo.Handle-Based Instance Retrieval
The system SHALL use `cgo.Handle` values to retrieve GDExtension class instances in per-instance callbacks (`FreeInstance`, `GetPropertyList`, etc.). The Handle SHALL be deleted by the instance-free callback after the wrapped object's `Destroy()` has returned, releasing the wrapped Go instance graph; it SHALL NOT be deleted at any earlier point, so callbacks issued during destruction still resolve.

#### Scenario: Instance retrieved via cgo.Handle in FreeInstance callback
- **WHEN** Godot calls the `GDExtensionClassFreeInstance` callback with a non-nil `p_instance` pointer
- **THEN** the system SHALL retrieve the `WrappedClassInstance` from the `cgo.Handle` stored in `p_instance`
- **AND** the system SHALL invoke `Destroy()` before performing the handle deletion
- **AND** the Handle SHALL be deleted after `Destroy()` returns, so re-entrant callbacks issued during `Destroy()` still observe a live handle

#### Scenario: No engine callback resolves the deleted handle
- **WHEN** `FreeInstance` returns and the engine proceeds with object destruction
- **THEN** the system SHALL NOT be required to resolve the deleted handle through any per-instance callback, because the engine clears its stored extension-instance pointer immediately after `free_instance` returns
- **AND** a lookup that nonetheless receives a binding slot holding the deleted value SHALL fail with the existing typed binding-lookup error rather than a crash

#### Scenario: FreeInstance returns early on nil pointer
- **WHEN** Godot calls the `GDExtensionClassFreeInstance` callback with a nil `p_instance`
- **THEN** the system SHALL return immediately without performing any cleanup

### Requirement: SetConstructInfo Matches godot-cpp Pattern
The system SHALL provide a `SetConstructInfo` function that sets the extension instance handle and instance binding on a newly constructed Godot object, matching the godot-cpp `Wrapped::_set_construct_info` pattern. The stored instance handle SHALL carry a `*WrappedClassInstance` so the shared instance-free release path can resolve it, and the function SHALL NOT pin the callbacks argument, or any other per-instance value, on program-lifetime storage.

#### Scenario: SetConstructInfo binds instance to Godot object
- **WHEN** `SetConstructInfo` is called with a Wrapped instance, extension class name, and binding callbacks
- **THEN** `GDExtensionInterfaceObjectSetInstance` SHALL be called to associate the `*WrappedClassInstance` handle with the Godot object
- **AND** `GDExtensionInterfaceObjectSetInstanceBinding` SHALL be called to register the instance binding callbacks
- **AND** the engine's copy of the callbacks struct SHALL be the sole storage retaining the callback pointers beyond the call

## ADDED Requirements

### Requirement: Instance Construction Adds No Program-Lifetime Retention
Construction paths for extension class instances SHALL NOT retain per-instance Go values on program-lifetime storage (the package pinner). Reachability of the wrapped instance for the life of the Godot object SHALL be provided solely by the `cgo.Handle`, which is released by the instance-free callback.

#### Scenario: Post-initialization pins nothing per instance
- **WHEN** `WrappedPostInitialize` binds a newly created user-defined instance to its Godot object
- **THEN** no `Pin` call on program-lifetime storage SHALL occur for that instance's `obj`, `owner`, `inst`, class-name pointer, or callbacks copy
- **AND** the instance remains resolvable through its handle for every subsequent engine callback

#### Scenario: Freed instances stop retaining Go memory
- **WHEN** the demo creates and frees a large batch of user-defined extension class instances through the engine and forces Go garbage collection
- **THEN** live Go heap after collection stays within a constant band regardless of batch size
- **AND** weak references to wrappers from a completed batch report empty once the engine has freed the corresponding objects
