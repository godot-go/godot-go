## Purpose

Ensures all Godot built-in vector and matrix types (Vector2, Vector2i, Vector3, Vector3i, Rect2, Rect2i, Vector4, Vector4i, Transform2D, Plane, Quaternion, AABB, Basis, Transform3D, Projection) decode and encode correctly as method arguments across both the ptrcall and varcall call paths.

### Requirement: Vector-type varcall argument decoding
The system SHALL decode all vector built-in types from the Variant call path without panicking, producing the correct Go built-in value for each argument.

#### Scenario: Rect2 argument via Variant call
- **WHEN** a Go method receives a `Rect2` argument via a Variant call
- **THEN** the argument is decoded to the Go `Rect2` value with the same content and the call does not panic.

#### Scenario: Transform2D argument via Variant call
- **WHEN** a Go method receives a `Transform2D` argument via a Variant call
- **THEN** the argument is decoded to the Go `Transform2D` value and the call does not panic.

#### Scenario: Plane argument via Variant call
- **WHEN** a Go method receives a `Plane` argument via a Variant call
- **THEN** the argument is decoded to the Go `Plane` value and the call does not panic.

#### Scenario: Quaternion argument via Variant call
- **WHEN** a Go method receives a `Quaternion` argument via a Variant call
- **THEN** the argument is decoded to the Go `Quaternion` value and the call does not panic.

#### Scenario: AABB argument via Variant call
- **WHEN** a Go method receives an `AABB` argument via a Variant call
- **THEN** the argument is decoded to the Go `AABB` value and the call does not panic.

#### Scenario: Basis argument via Variant call
- **WHEN** a Go method receives a `Basis` argument via a Variant call
- **THEN** the argument is decoded to the Go `Basis` value and the call does not panic.

#### Scenario: Transform3D argument via Variant call
- **WHEN** a Go method receives a `Transform3D` argument via a Variant call
- **THEN** the argument is decoded to the Go `Transform3D` value and the call does not panic.

#### Scenario: Projection argument via Variant call
- **WHEN** a Go method receives a `Projection` argument via a Variant call
- **THEN** the argument is decoded to the Go `Projection` value and the call does not panic.

### Requirement: Vector-type ptrcall argument decoding
The system SHALL decode all vector built-in types from the ptrcall path without panicking, producing the correct Go built-in value for each argument.

#### Scenario: All vector types via ptrcall
- **WHEN** a Go method receives any vector built-in type argument via a ptrcall
- **THEN** the argument is decoded to the corresponding Go value and the call does not panic.

### Requirement: Vector-type round-trip fidelity
The system SHALL preserve vector argument values through both call paths so that a Go method returning its received vector argument produces the original value on the GDScript side.

#### Scenario: Ptrcall round-trip
- **WHEN** a Go method returns the vector value it received via a ptrcall
- **THEN** the value in GDScript equals the original.

#### Scenario: Variant call round-trip
- **WHEN** a Go method returns the vector value it received via a Variant call
- **THEN** the value in GDScript equals the original.
