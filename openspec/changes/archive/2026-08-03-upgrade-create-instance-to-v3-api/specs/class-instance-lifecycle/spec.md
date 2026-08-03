## Purpose

Defines the lifecycle contract for GDExtension class instances created from Go, using the `CreateInstance2` callback and `cgo.Handle`-based instance tracking.

## ADDED Requirements

### Requirement: Class Instance Creation Uses CreateInstance2 Callback
The system SHALL provide a `GoCallback_ClassCreationInfoCreateInstance2` callback with the signature `(unsafe.Pointer, GDExtensionBool) GDExtensionObjectPtr`, matching the `GDExtensionClassCreateInstance2` C type.

#### Scenario: GDClass instance constructed via CreateInstance2 callback
- **WHEN** Godot invokes the class creation callback
- **THEN** the Go-side callback `GoCallback_ClassCreationInfoCreateInstance2` SHALL be called with the class userdata pointer and a `p_notify_postinitialize` boolean
- **AND** the callback SHALL return a `GDExtensionObjectPtr` to the newly created Godot object

#### Scenario: Class registration uses CreationInfo4 with CreateInstance2 function
- **WHEN** `ClassDBRegisterClass` is called for a Go-defined GDClass
- **THEN** the `GDExtensionClassCreationInfo4` struct SHALL be populated with `cgo_classcreationinfo_createinstance2` cast as `GDExtensionClassCreateInstance2`
- **AND** the class SHALL be registered via `GDExtensionInterfaceClassdbRegisterExtensionClass4`

### Requirement: C String Userdata for Class Callbacks
The system SHALL pass a C string (class name) as the `class_userdata` in `GDExtensionClassCreationInfo4`, allocated via `C.CString`.

#### Scenario: C string passed as creation callback userdata
- **WHEN** a class is registered with `ClassDBRegisterClass`
- **THEN** the `class_userdata` field SHALL be a `C.CString` of the class name
- **AND** the `CreateInstance2` callback SHALL recover the class name via `C.GoString`

### Requirement: cgo.Handle-Based Instance Retrieval
The system SHALL use `cgo.Handle` values to retrieve GDExtension class instances in per-instance callbacks (`FreeInstance`, `GetPropertyList`, etc.). The Handle SHALL NOT be deleted after retrieval — this matches origin/main behavior and prevents stale-pointer crashes in Godot.

#### Scenario: Instance retrieved via cgo.Handle in FreeInstance callback
- **WHEN** Godot calls the `GDExtensionClassFreeInstance` callback with a non-nil `p_instance` pointer
- **THEN** the system SHALL retrieve the `WrappedClassInstance` from the `cgo.Handle` stored in `p_instance`
- **AND** the Handle SHALL NOT be deleted (Godot still holds the raw pointer internally)

#### Scenario: FreeInstance returns early on nil pointer
- **WHEN** Godot calls the `GDExtensionClassFreeInstance` callback with a nil `p_instance`
- **THEN** the system SHALL return immediately without performing any cleanup

### Requirement: Re-Entrant FreeInstance Guard
The system SHALL track created instances in `Internal.GDClassInstances` and guard `FreeInstance` against re-entrant calls using a SyncMap existence check.

#### Scenario: Re-entrant FreeInstance is detected
- **WHEN** `FreeInstance` is called re-entrantly through `Destroy()` (releasing the last Godot reference)
- **THEN** the second invocation SHALL fail the `Internal.GDClassInstances` lookup and panic with a clear error message
- **AND** the first invocation's `defer Destroy()` fires after the guard clears, preventing double C++ destructor invocation

#### Scenario: Normal FreeInstance flow
- **WHEN** `FreeInstance` is called for a tracked instance
- **THEN** the guard SHALL check `Internal.GDClassInstances`, delete the entry, then invoke `Destroy()` via defer

### Requirement: SetConstructInfo Matches godot-cpp Pattern
The system SHALL provide a `SetConstructInfo` function that sets the extension instance handle and instance binding on a newly constructed Godot object, matching the godot-cpp `Wrapped::_set_construct_info` pattern.

#### Scenario: SetConstructInfo binds instance to Godot object
- **WHEN** `SetConstructInfo` is called with a Wrapped instance, extension class name, and binding callbacks
- **THEN** `GDExtensionInterfaceObjectSetInstance` SHALL be called to associate the instance handle with the Godot object
- **AND** `GDExtensionInterfaceObjectSetInstanceBinding` SHALL be called to register the instance binding callbacks

### Requirement: String-Based CreateGDClassInstance2
The system SHALL provide a `CreateGDClassInstance2(tn string) GDClass` function that creates instances by class name using reflection and `ClassdbConstructObject2`.

#### Scenario: CreateGDClassInstance2 creates instance by name
- **WHEN** `CreateGDClassInstance2("Example")` is called
- **THEN** the system SHALL look up the `ClassInfo` by name, construct the native Godot object via `ClassdbConstructObject2`, create the Go instance via reflection, and call `WrappedPostInitialize` to bind it
