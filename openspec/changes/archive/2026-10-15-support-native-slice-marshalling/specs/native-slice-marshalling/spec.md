# Spec Delta

## Purpose

Lets Go method signatures use native slices (`[]byte`, `[]string`, `[]int32`, vector and color slices) wherever the engine uses `Packed*Array` containers, with defined mapping, copy semantics, ownership, and rejection behavior.

## ADDED Requirements

### Requirement: Slice arguments decode from packed arrays on both call paths
A bound Go method declaring a supported native slice parameter SHALL receive a deep-copied Go slice when the engine passes the corresponding `Packed*K` value, via both the ptrcall and the varcall path, and the decoded slice SHALL be Go-owned memory independent of the engine container.

#### Scenario: Ptrcall argument decode
- **WHEN** GDScript calls a bound method directly with a typed `PackedInt32Array` variable and the Go parameter is `[]int32`
- **THEN** the Go method receives a slice with the same length and element values, and the call does not panic

#### Scenario: Varcall argument decode
- **WHEN** GDScript invokes the same method through `example.call(...)` with a `PackedStringArray`
- **THEN** the Go method receives a `[]string` with equal element contents

#### Scenario: Nil argument is rejected at the call layer
- **WHEN** GDScript passes `null` where a packed-parameter is declared, direct or through `call()`
- **THEN** the binder's strict conversion gate rejects the call before the Go method runs (Nil has no entry in the engine's strict conversion table for packed types; the engine itself forwards raw Variants unvalidated, so the binding must enforce this), and the rejection is contained in the suite so later assertions still execute

### Requirement: Slice returns encode to packed arrays
A bound Go method returning a supported native slice SHALL hand GDScript a `Packed*K` value with the same length and element values, on both direct and `call()` dispatch. A nil returned slice SHALL produce an empty packed array rather than null.

#### Scenario: Typed direct-call return
- **WHEN** a bound Go method returns `[]float64{1.5, -2.25}` and GDScript stores the result in a `PackedFloat64Array`
- **THEN** the GDScript array equals the returned elements

#### Scenario: Nil slice return
- **WHEN** a bound Go method returns a nil `[]byte`
- **THEN** GDScript receives an empty `PackedByteArray` (`.is_empty()` is true, not `null`)

### Requirement: Registration advertises packed variant types and rejects unmappable elements
A bound method's slice parameters and return SHALL be registered with the `PACKED_*` variant type matching the element type (never the element's scalar type), and binding a slice whose element type has no packed counterpart SHALL fail at bind time with an error naming the element type.

#### Scenario: Slice parameter registers as packed type
- **WHEN** a method with a `[]int32` parameter is bound
- **THEN** the parameter's registered variant type is `PACKED_INT32_ARRAY`, and a typed GDScript `PackedInt32Array` argument passes static checking

#### Scenario: Unmappable element rejected at bind time
- **WHEN** a method with a `[]bool` parameter (no packed bool array exists in the engine) is bound
- **THEN** binding fails with an error naming `bool` instead of silently registering a wrong variant type

#### Scenario: Non-convertible argument rejected at the binder gate
- **WHEN** GDScript passes an `int` where a `PackedByteArray`-typed parameter is declared
- **THEN** the binder's strict conversion gate rejects the call before the Go method runs, and the failure is surfaced to GDScript

### Requirement: Round-trip fidelity across all packed element kinds
Element values SHALL survive a GDScript → Go → GDScript round trip unchanged for every supported pairing: bytes, `int32`, `int64` (full 64-bit range including negatives), `float32`, `float64`, `Vector2`, `Vector3`, `Vector4`, `Color`, and UTF-8 strings including non-ASCII text, at both small and bulk (≥10,000 element) sizes.

#### Scenario: UTF-8 string array round trip
- **WHEN** GDScript passes a `PackedStringArray` containing non-ASCII entries such as `"héllo"` and `"日本"` to a `[]string` parameter and the Go method returns it unchanged
- **THEN** the returned array equals the original element-for-element

#### Scenario: Bulk byte round trip
- **WHEN** GDScript passes a 10,000-element `PackedByteArray` with distinct values to a `[]byte` parameter and the Go method returns it unchanged
- **THEN** the returned array equals the original

### Requirement: Decoded slices are independent of their engine arrays
A decoded slice SHALL remain valid after the call completes (safe to retain across frames and release through ordinary Go GC, with no engine release call required), and mutations on either side after decoding SHALL NOT be visible on the other side.

#### Scenario: Retention across frames
- **WHEN** a Go method stores a decoded `[]byte` argument and reads it on a later frame
- **THEN** the contents are intact, no cgocheck violation is reported, and the suite reports no leaked engine objects

#### Scenario: No write-back
- **WHEN** a Go method mutates its decoded slice and returns, and GDScript then reads the original array
- **THEN** the original array still holds the original values
