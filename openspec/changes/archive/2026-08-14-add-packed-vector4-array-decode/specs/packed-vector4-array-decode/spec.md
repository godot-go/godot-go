## Purpose

Ensures `PackedVector4Array` decodes and encodes correctly as a method argument and return value across both the ptrcall and varcall call paths, matching the existing support for `PackedVector3Array` and other container built-in types.

## ADDED Requirements

### Requirement: PackedVector4Array varcall argument decoding
The system SHALL decode `PackedVector4Array` arguments from the Variant call path without panicking, producing the correct Go `PackedVector4Array` value.

#### Scenario: PackedVector4Array argument via varcall
- **WHEN** a Go method receives a `PackedVector4Array` argument via a varcall
- **THEN** the argument is decoded to the Go `PackedVector4Array` value and the call does not panic.

### Requirement: PackedVector4Array ptrcall argument decoding
The system SHALL decode `PackedVector4Array` arguments from the ptrcall path without panicking, producing the correct Go `PackedVector4Array` value via an owned copy constructor.

#### Scenario: PackedVector4Array argument via ptrcall
- **WHEN** a Go method receives a `PackedVector4Array` argument via a ptrcall
- **THEN** the argument is decoded to the Go `PackedVector4Array` value and the call does not panic.

### Requirement: PackedVector4Array round-trip fidelity
The system SHALL preserve `PackedVector4Array` argument values through both call paths so that a Go method returning its received value produces the original value on the GDScript side.

#### Scenario: Ptrcall round-trip
- **WHEN** a Go method returns the `PackedVector4Array` value it received via a ptrcall
- **THEN** the value in GDScript equals the original.

#### Scenario: Varcall round-trip
- **WHEN** a Go method returns the `PackedVector4Array` value it received via a varcall
- **THEN** the value in GDScript equals the original.
