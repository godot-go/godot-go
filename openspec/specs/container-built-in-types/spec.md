## Purpose

Ensures all container built-in types (`Array`, `PackedByteArray`, `PackedInt32Array`, `PackedInt64Array`, `PackedFloat32Array`, `PackedFloat64Array`, `PackedStringArray`, `PackedVector2Array`, `PackedVector3Array`, `PackedColorArray`) decode and encode correctly as method arguments and return values across both the ptrcall and varcall call paths.

## Requirements

### Requirement: Varcall container argument decoding
The system SHALL decode all container built-in types from the Variant call path without panicking, producing the correct Go built-in value for each argument.

#### Scenario: Array argument via varcall
- **WHEN** a Go method receives an `Array` argument via a varcall
- **THEN** the argument is decoded to the Go `Array` value and the call does not panic.

#### Scenario: PackedByteArray argument via varcall
- **WHEN** a Go method receives a `PackedByteArray` argument via a varcall
- **THEN** the argument is decoded to the Go `PackedByteArray` value and the call does not panic.

#### Scenario: PackedInt32Array argument via varcall
- **WHEN** a Go method receives a `PackedInt32Array` argument via a varcall
- **THEN** the argument is decoded to the Go `PackedInt32Array` value and the call does not panic.

#### Scenario: PackedInt64Array argument via varcall
- **WHEN** a Go method receives a `PackedInt64Array` argument via a varcall
- **THEN** the argument is decoded to the Go `PackedInt64Array` value and the call does not panic.

#### Scenario: PackedFloat32Array argument via varcall
- **WHEN** a Go method receives a `PackedFloat32Array` argument via a varcall
- **THEN** the argument is decoded to the Go `PackedFloat32Array` value and the call does not panic.

#### Scenario: PackedFloat64Array argument via varcall
- **WHEN** a Go method receives a `PackedFloat64Array` argument via a varcall
- **THEN** the argument is decoded to the Go `PackedFloat64Array` value and the call does not panic.

#### Scenario: PackedStringArray argument via varcall
- **WHEN** a Go method receives a `PackedStringArray` argument via a varcall
- **THEN** the argument is decoded to the Go `PackedStringArray` value and the call does not panic.

#### Scenario: PackedVector2Array argument via varcall
- **WHEN** a Go method receives a `PackedVector2Array` argument via a varcall
- **THEN** the argument is decoded to the Go `PackedVector2Array` value and the call does not panic.

#### Scenario: PackedVector3Array argument via varcall
- **WHEN** a Go method receives a `PackedVector3Array` argument via a varcall
- **THEN** the argument is decoded to the Go `PackedVector3Array` value and the call does not panic.

#### Scenario: PackedColorArray argument via varcall
- **WHEN** a Go method receives a `PackedColorArray` argument via a varcall
- **THEN** the argument is decoded to the Go `PackedColorArray` value and the call does not panic.

### Requirement: Ptrcall container argument decoding
The system SHALL decode all container built-in types from the ptrcall path without panicking, producing the correct Go built-in value for each argument.

#### Scenario: All container types via ptrcall
- **WHEN** a Go method receives any container built-in type argument via a ptrcall
- **THEN** the argument is decoded to the corresponding Go value and the call does not panic.

#### Scenario: PackedVector4Array argument via ptrcall
- **WHEN** a Go method receives a `PackedVector4Array` argument via a ptrcall
- **THEN** the argument is decoded to the Go `PackedVector4Array` value and the call does not panic.

### Requirement: Container-type round-trip fidelity
The system SHALL preserve container argument values through both call paths so that a Go method returning its received container argument produces the original value on the GDScript side.

#### Scenario: Ptrcall round-trip
- **WHEN** a Go method returns the container value it received via a ptrcall
- **THEN** the value in GDScript equals the original.

#### Scenario: Varcall round-trip
- **WHEN** a Go method returns the container value it received via a varcall
- **THEN** the value in GDScript equals the original.

#### Scenario: PackedVector4Array ptrcall round-trip
- **WHEN** a Go method returns the `PackedVector4Array` value it received via a ptrcall
- **THEN** the value in GDScript equals the original.

#### Scenario: PackedVector4Array varcall round-trip
- **WHEN** a Go method returns the `PackedVector4Array` value it received via a varcall
- **THEN** the value in GDScript equals the original.
