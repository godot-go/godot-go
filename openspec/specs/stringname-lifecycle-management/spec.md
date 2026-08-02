## Purpose

Manages the lifetime of Godot `StringName` handles when passed from Go to the GDExtension interface, eliminating Go-side orphaned-warning and double-free hazards during class registration.

## Requirements

### Requirement: C-Memory Allocation for StringNames
The system MUST provide a mechanism to allocate `StringName` data in C-managed memory (outside the Go heap) when preparing arguments for GDExtension interface calls that require `GDExtensionConstStringNamePtr`.

#### Scenario: Allocation for Property Registration
- **WHEN** a `StringName` is needed for a `GDExtensionPropertyInfo` struct
- **THEN** the `StringName` data is allocated in C memory, ensuring it doesn't move or get prematurely GC'd by Go.

### Requirement: Safe StringName Destruction
The system SHALL ensure that `StringName` objects allocated in C memory for interface calls are explicitly destroyed after the interface call returns, preventing memory leaks.

#### Scenario: Cleanup after Registration
- **WHEN** `ClassDBAddProperty` completes its interface call to Godot
- **THEN** all temporary C-allocated `StringName` objects used during the call are destroyed.

### Requirement: Elimination of Go-side Orphaned Warnings
The implementation MUST eliminate all orphaned StringName warnings attributable to Go-side lifecycle bugs (dangling pointers, premature destruction, GC movement) from the project's test output.

#### Scenario: Clean Test Output for Go-side Orphans
- **WHEN** `make test` is executed
- **THEN** the output contains no orphaned StringName warnings for classes, properties, signals, or methods registered from Go.

#### Scenario: No Orphaned StringName Warnings Remain
- **WHEN** `make test` exits
- **THEN** the output contains no orphaned StringName warnings at all, including for engine object instances (e.g. `Image`) passed into Go methods via `ObjectGetClassName`.

### Requirement: Return Path Refcount Cleanup
StringName and NodePath values created on the Go side and returned to
Godot SHALL have their Go-side refcount contribution released after
encoding into Godot's output buffer, preventing refcount leaks that
manifest as orphan StringName warnings at process exit.

#### Scenario: Varcall return of Go-created StringName
- **WHEN** a Go method returns a StringName (created via
  `NewStringNameWithUtf8Chars` or equivalent) through the varcall return
  path
- **AND** the StringName is encoded into a Godot Variant via
  `GDExtensionVariantPtrFromReflectValue`
- **THEN** the Go-side StringName SHALL be destroyed after encoding,
  ensuring the Godot Variant holds the sole remaining refcount

#### Scenario: Ptrcall return of Go-created StringName
- **WHEN** a Go method returns a StringName through the ptrcall return
  path via `GDExtensionTypePtrFromReflectValue`
- **THEN** the StringName SHALL be copied into Godot's output buffer
  using the GDExtension copy constructor and the Go-side copy SHALL be
  destroyed, ensuring proper ownership transfer

### Requirement: Argument Decoding Borrow Semantics
StringName and NodePath values received as arguments from Godot SHALL use
borrow semantics: the Go side SHALL NOT acquire an owned refcount that
survives the method call.

#### Scenario: Varcall decode of a StringName argument
- **WHEN** a Godot varcall passes a StringName argument to a Go method
- **AND** the argument is decoded via `convertVariantToGoTypeReflectValue`
- **THEN** the decoded StringName SHALL be created via copy constructor
  then immediately destroyed after byte-copying, leaving a borrowed value
  with no Go-side refcount ownership

#### Scenario: Ptrcall decode of a StringName argument
- **WHEN** a Godot ptrcall passes a StringName argument to a Go method
  via `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`
- **THEN** the StringName SHALL be decoded via byte-copy only, without
  calling the GDExtension copy constructor, preserving the Godot-side
  refcount unchanged
