## Purpose

Ensures the basic built-in types (`bool`, integer and float widths, Go `string`, and the Godot `String`, `StringName`, `NodePath` types) decode and encode correctly as method arguments and return values across both the ptr call and Variant call paths.

## Requirements

### Requirement: Scalar argument decoding
The system SHALL decode `bool`, integer (`int`, `uint`, and fixed-width variants), floating-point (`float32`, `float64`), and Go native `string` method arguments from both the ptr call and Variant call paths without panicking and with values that match what Godot passed.

#### Scenario: Ptr call scalar round-trip
- **WHEN** a Go method receives `bool`, `int64`, `float64`, and Go `string` arguments via a ptr call
- **THEN** the Go values match the values passed from GDScript and the call does not panic.

#### Scenario: Variant call scalar round-trip
- **WHEN** a Go method receives `bool`, `int64`, `float64`, and Go `string` arguments via a Variant call
- **THEN** the Go values match the values passed from GDScript and the call does not panic.

### Requirement: Godot string-family argument decoding
The system SHALL decode Godot `String`, `StringName`, `NodePath`, and `Array` method arguments in both the ptr call and Variant call paths without panicking, producing the corresponding Go built-in value.

#### Scenario: String argument via ptr call
- **WHEN** a Go method receives a Godot `String` argument via a ptr call
- **THEN** the argument is converted to the Go `String` value with the same content and the call does not panic.

#### Scenario: StringName argument via Variant call
- **WHEN** a Go method receives a `StringName` argument via a Variant call
- **THEN** the argument is converted to the Go `StringName` value and the call does not panic.

#### Scenario: NodePath argument via ptr call
- **WHEN** a Go method receives a `NodePath` argument via a ptr call
- **THEN** the argument is converted to the Go `NodePath` value and the call does not panic.

#### Scenario: Array argument via Variant call
- **WHEN** a Go method receives an `Array` argument via a Variant call
- **THEN** the argument is converted to the Go `Array` value with the same content and the call does not panic.

### Requirement: Unsigned integer argument fidelity
The system SHALL decode unsigned integer method arguments in the Variant call path without sign-wrapping, so `uint64` values above `MaxInt64` are preserved exactly.

#### Scenario: Large uint64 argument
- **WHEN** a Go method receives a `uint64` argument whose value is greater than `MaxInt64` via a Variant call
- **THEN** the Go value equals the value passed from GDScript.

### Requirement: Narrow numeric return encoding
The system SHALL encode narrow integer and floating-point return values through the Variant call path by widening to the engine's `int64`/`float64` representation, without reading past the end of the Go value.

#### Scenario: Narrow numeric Variant return
- **WHEN** a Go method returns an `int8`, `uint16`, or `float32` value via a Variant call
- **THEN** Godot receives the correctly widened numeric value and no out-of-bounds read occurs.

### Requirement: Go string return lifecycle
The system SHALL release the temporary Godot `String` created when a Go `string` return value is encoded through the Variant call path, so repeated returns do not leak engine memory.

#### Scenario: Repeated string returns
- **WHEN** a Go method returns a Go `string` via a Variant call many times
- **THEN** the engine String memory allocated for each return is released after the call.
