## Purpose

Refines `stringname-lifecycle-management` for the ptrcall return path: echoing a borrowed StringName/NodePath argument must not release the refcount Godot's caller still holds.

## MODIFIED Requirements

### Requirement: Return Path Refcount Cleanup
StringName and NodePath values created on the Go side and returned to Godot SHALL have their Go-side refcount contribution released after encoding into Godot's output buffer, preventing refcount leaks that manifest as orphan StringName warnings at process exit. The sole exception is an echoed borrow: when the returned value is byte-for-byte identical to a StringName/NodePath argument decoded as a borrow for this call, the Go-side source SHALL NOT be destroyed, because destroying it would release a reference Godot's caller still holds (the output buffer already received its own copy).

#### Scenario: Varcall return of Go-created StringName
- **WHEN** a Go method returns a StringName (created via `NewStringNameWithUtf8Chars` or equivalent) through the varcall return path
- **AND** the StringName is encoded into a Godot Variant via `GDExtensionVariantPtrFromReflectValue`
- **THEN** the Go-side StringName SHALL be destroyed after encoding, ensuring the Godot Variant holds the sole remaining refcount

#### Scenario: Ptrcall return of Go-created StringName
- **WHEN** a Go method returns a StringName through the ptrcall return path via `GDExtensionTypePtrFromReflectValue`
- **THEN** the StringName SHALL be copied into Godot's output buffer using the GDExtension copy constructor and the Go-side copy SHALL be destroyed, ensuring proper ownership transfer

#### Scenario: Ptrcall echo of a borrowed argument is not destroyed
- **WHEN** a Go method returns the same StringName or NodePath argument it received via ptrcall
- **THEN** the output buffer receives a fresh copy but the borrowed source is left undestroyed, and repeated calls produce no orphaned-reference warnings

## ADDED Requirements

### Requirement: Ptrcall Borrow Echo Detection Covers Refcounted Basic Types
The ptrcall return path SHALL classify whether a returned `StringName` or `NodePath` is a byte-for-byte echo of one of the call's borrow-decoded arguments before deciding to destroy the source value, using the same content-based detection applied to container types.

#### Scenario: Detection distinguishes echoes from fresh values
- **WHEN** a Go method returns either the borrowed argument itself or a newly constructed value whose contents differ from every argument
- **THEN** the former is classified as an echo (source not destroyed) and the latter is not (source destroyed after copy-constructor encoding)

#### Scenario: Round-trips are clean in both call styles
- **WHEN** GDScript callers pass StringName/NodePath arguments to echo methods and to methods returning differently-valued results, via ptrcall and via varcall
- **THEN** returned values have the expected contents and no orphaned-reference or double-free symptoms appear across repeated calls
