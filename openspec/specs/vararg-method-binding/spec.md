## Purpose

Defines how a Go variadic method becomes a Godot vararg method: the bind-time
conditions that must agree, how the engine sees the signature, and how the
caller's raw arguments reach the Go function.

## Requirements

### Requirement: A variadic Go signature and the vararg flag must agree

A method SHALL be bindable as variadic only when the Go method is declared
variadic and the bind requests the vararg flag. A declarative mismatch in either
direction SHALL fail at bind time before engine registration.

#### Scenario: Variadic Go method with the flag

- **WHEN** a Go method declared with a trailing `...Variant` parameter is bound with the vararg flag
- **THEN** the bind succeeds and the engine sees a vararg method

#### Scenario: Variadic Go method without the flag

- **WHEN** a variadic Go method is bound as an ordinary method
- **THEN** the bind fails before engine registration

#### Scenario: Flag without a variadic Go method

- **WHEN** a non-variadic Go method is bound with the vararg flag
- **THEN** the bind fails before engine registration

### Requirement: Variadic binds register no named arguments

A variadic bind SHALL advertise zero named argument entries to the engine, so
that a GDScript call supplying no arguments parses and dispatches, and the
caller's arguments are delivered to the Go function as the variadic tail.

#### Scenario: Zero-argument call parses

- **WHEN** GDScript calls a vararg method with no arguments
- **THEN** the call parses and the Go method receives an empty variadic slice

#### Scenario: Excess arguments are accepted

- **WHEN** GDScript calls the vararg method with more arguments than any named parameter
- **THEN** the call reaches the Go method with every supplied argument in order

### Requirement: Variadic binds reject bound default arguments

A variadic bind SHALL NOT accept default argument values, because the engine
computes a required-argument count from the registered defaults and a variadic
signature has no fixed slots for them to fill.

#### Scenario: Defaults supplied with a vararg bind

- **WHEN** a variadic bind supplies one or more default values
- **THEN** the bind fails before engine registration

### Requirement: Variadic dispatch passes raw Variants without per-argument type validation

Each call SHALL collect the caller's arguments as raw borrowed Variants, skip
argument-type conversion checks, and pass them unchanged to the Go function
through variadic invocation. Arity validation SHALL accept any supplied count.

#### Scenario: Mixed argument kinds

- **WHEN** a caller supplies an integer, a string, and an object to the same variadic call
- **THEN** the Go method receives all three as Variants in order

#### Scenario: No arguments

- **WHEN** a caller supplies no arguments
- **THEN** no type validation runs and the variadic slice is empty

### Requirement: Dispatch omits the receiver for static binds

An instance variadic bind SHALL receive the dispatch instance before the
variadic slice; a static variadic bind SHALL receive only the variadic slice, in
the same order and with the same values as the caller supplied.

#### Scenario: Instance variadic bind

- **WHEN** the engine calls a variadic method on an instance
- **THEN** the Go receiver is the instance and the slice holds the caller's arguments

#### Scenario: Static variadic bind

- **WHEN** the engine calls a variadic static method with a null instance
- **THEN** the Go function receives only the variadic slice

### Requirement: Variadic returns are delivered as a Variant

A variadic bind that returns nothing SHALL yield a nil Variant to the engine; a
bind that returns a value SHALL encode its first return value into the call's
return Variant. A second declared return value is not delivered to the engine.

#### Scenario: Void vararg method

- **WHEN** a vararg Go method returns no value
- **THEN** the engine receives a nil return

#### Scenario: Value-returning vararg method

- **WHEN** a vararg Go method returns a Variant
- **THEN** the engine receives that Variant

#### Scenario: Two return values declared

- **WHEN** a vararg Go method returns a value and a bool
- **THEN** the first value is delivered and the bool does not reach the engine
