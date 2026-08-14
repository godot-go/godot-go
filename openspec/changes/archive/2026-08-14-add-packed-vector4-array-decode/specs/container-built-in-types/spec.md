## MODIFIED Requirements

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
