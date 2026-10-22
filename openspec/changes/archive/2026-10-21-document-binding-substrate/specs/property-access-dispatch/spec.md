# Spec Delta

## Purpose

Defines how the engine's property get, set, and validation requests reach a Go
class: which virtual implementations answer them, how the handled flag is
reported, and what lifetime the delivered value has.

## ADDED Requirements

### Requirement: Property get dispatches to the class's get virtual

When the engine asks a registered class for a property by name, the binding SHALL
invoke that class's bound `_get` virtual with the instance and the property name,
and SHALL report the result through the engine's handled flag and value slot.

#### Scenario: Handler handles the property

- **WHEN** the engine reads a property whose name the `_get` implementation recognises
- **THEN** the handler returns a Variant and true
- **AND** the engine receives that value and treats the property as handled

#### Scenario: Handler declines the property

- **WHEN** the `_get` implementation sees a name it does not own and returns false
- **THEN** the engine treats the property as unhandled and no value is written

### Requirement: A get handler returns a Variant and a handled flag

A `_get` implementation SHALL return exactly two values, the first being a
Variant and the second a bool. A handler whose first return is not a Variant is an
authoring error and SHALL be reported fatally rather than delivering a malformed
value.

#### Scenario: Wrong first return type

- **WHEN** a bound `_get` returns a `string` and a bool
- **THEN** the request fails loudly and the engine receives no value

#### Scenario: Two-value handler

- **WHEN** a bound `_get` returns a Variant and a bool
- **THEN** the request completes without an authoring error

### Requirement: Property set dispatches a name and a borrowed value

When the engine writes a property, the binding SHALL invoke the class's bound
`_set` virtual with the instance, the property name, and the delivered value as a
borrowed Variant that is valid only for the duration of that call, and SHALL
report the handler's bool as the handled flag.

#### Scenario: Handler accepts the value

- **WHEN** the `_set` implementation stores the value during the call and returns true
- **THEN** the engine treats the property as handled

#### Scenario: Handler declines the value

- **WHEN** the `_set` implementation returns false
- **THEN** the engine treats the property as unhandled

#### Scenario: Handler retains the value

- **WHEN** a `_set` implementation wants to keep the delivered value beyond the call
- **THEN** it must take an owned copy, because the delivered Variant borrows engine storage that is not valid afterwards

### Requirement: Missing accessors decline instead of failing

A registered class with no bound `_get` or `_set` SHALL report the property
request as unhandled, so the engine can fall back to its own storage semantics
without an error.

#### Scenario: Class binds only `_set`

- **WHEN** the engine reads a property on a class with no `_get`
- **THEN** the read is reported unhandled and no error is raised

#### Scenario: Class binds neither

- **WHEN** the engine reads or writes a property on a class with neither accessor
- **THEN** both requests are reported unhandled

### Requirement: Property validation uses the registration-supplied validator

When the engine validates a property on a registered class, the binding SHALL
invoke the validator supplied at registration with the engine's property metadata
and report the property as handled; a class registered without a validator SHALL
report the property as not handled.

#### Scenario: Validator registered

- **WHEN** the engine validates a property on a class registered with a validator
- **THEN** the validator receives the engine's property metadata once

#### Scenario: No validator registered

- **WHEN** the engine validates a property on a class registered without one
- **THEN** the request is reported not handled and no callback runs

### Requirement: An unresolvable class fails a get, set, or validation request

A property get, set, or validation request whose instance cannot be resolved to a
registered class SHALL fail loudly through the library's fatal-error channel rather
than dereferencing a missing registration. Revert and notification requests take
the same missing registration as a deliberate decline, because an instance being
torn down reports a parent class name that was never registered.

#### Scenario: Unknown class registration

- **WHEN** a get, set, or validation callback receives an instance whose class has no registration
- **THEN** the request fails loudly and the engine receives no fabricated value

#### Scenario: Revert on a torn-down instance

- **WHEN** a revert or notification callback receives an instance whose reported class has no registration
- **THEN** the request declines without an error
