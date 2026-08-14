## MODIFIED Requirements

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
