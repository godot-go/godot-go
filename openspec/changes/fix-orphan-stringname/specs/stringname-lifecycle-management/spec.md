## ADDED Requirements

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

### Requirement: Elimination of Orphaned Warnings
The implementation MUST eliminate all "orphaned StringName" warnings from the project's test output.

#### Scenario: Clean Test Output
- **WHEN** `make test` is executed
- **THEN** the output contains zero occurrences of "orphaned StringName" warnings.
