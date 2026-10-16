# Spec Delta: ptrcall-object-argument-decode

## MODIFIED Requirements

### Requirement: Subclass Instances Satisfy A Base-Class Parameter

An object argument arriving through ptrcall SHALL be resolved the same way varcall resolves objects: through the instance binding first, so a user-defined extension-class instance receives the Go wrapper registered for that instance, and otherwise through the engine-class constructor keyed by the argument's runtime class name. An argument whose runtime class is a subclass of the declared parameter's class SHALL decode successfully. Failures in this resolution remain loud failures: ptrcall has no error slot, and their reporting is governed by the fatal-panic policy change.

#### Scenario: Subclass passed where the base class is declared

- **WHEN** a bound method declares a parameter of Go interface type `Shape2D` and is called with a `CircleShape2D`
- **THEN** the decoder resolves the constructor for `CircleShape2D`, the Go method body runs, and the value it receives reports its class as `CircleShape2D`

#### Scenario: User-defined instance arrives on a direct typed call

- **WHEN** GDScript makes a direct typed call passing an instance of a user-defined extension class to a parameter declared as an engine class such as `Node`
- **THEN** the decoder returns the original Go wrapper for that instance, with the same Godot instance id, and the process does not crash

#### Scenario: Wrapper identity preserved across dispatch paths

- **WHEN** the same user-defined instance is passed once through `obj.call(...)` and once through a direct typed call
- **THEN** both decodes resolve to the identical Go wrapper value

## REMOVED Requirements

### Requirement: Ref-Typed Interface Parameters Keep Existing Behaviour

**Reason**: Reference crossing for bound methods is now specified end to end by the `ref-method-boundary` capability (supported signatures, declared-class keying, owned references, nil, mismatches) across both dispatch paths; the "keep existing behaviour" pin — written when the existing behavior included the declared-name-only lookup without bare `Ref` support and a borrowing wrapper — would freeze behavior this change deliberately supersedes.

**Migration**: See the `ref-method-boundary` requirements `Ref arguments decode keyed to the declared class`, `Decoded Ref arguments own a reference`, and `Nil reference arguments yield the zero Ref`.
