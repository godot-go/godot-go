# Spec Delta

## Purpose

Governs how a Go-defined class publishes its metadata to the engine class database:
signals and their parameter descriptions, integer constants, enum grouping and
bitfield-ness, and property groups and subgroups, including the name collisions that
must be gated before the engine is asked.

## ADDED Requirements

### Requirement: A registered signal publishes one parameter description per declared parameter

A signal registration SHALL deliver exactly one parameter description per declared
parameter carrying that parameter's variant type and name, with no usage hint, and the
descriptions SHALL be complete by the time the engine call is made so that the
registration's temporaries may be released immediately afterwards.

#### Scenario: Signal with two parameters

- **WHEN** a class registers a signal declaring a string parameter and an integer parameter
- **THEN** the engine receives two parameter descriptions in declaration order with those variant types and names
- **AND** the descriptions carry no usage hint

#### Scenario: Signal with no parameters

- **WHEN** a class registers a signal declaring no parameters
- **THEN** the engine receives no parameter descriptions and the signal is registered

#### Scenario: Registration temporaries are not retained

- **WHEN** a signal with parameters has been registered
- **THEN** the parameter name temporaries are released and the run reports no orphaned name warnings

### Requirement: A signal name already carried by the class or an ancestor is rejected

A signal registration SHALL be rejected before the engine is asked when the signal name
is already carried by the class, including names it inherits from its ancestors. The
rejection SHALL be fatal and its diagnostic SHALL name both the class and the signal.

#### Scenario: Same name registered twice on one class

- **WHEN** a class registers a signal under a name it has already registered
- **THEN** the request fails fatally naming the class and that signal
- **AND** the engine is not asked to register the duplicate

#### Scenario: Name inherited from an engine class

- **WHEN** a class derived from an engine class registers a signal under a name that class already carries by inheritance
- **THEN** the request fails fatally naming the class and that signal
- **AND** no partially registered signal is left behind

#### Scenario: Rejected name is not consumed

- **WHEN** a signal registration is rejected as a duplicate
- **THEN** the class does not retain the rejected name as one it owns

#### Scenario: Distinct name is unaffected

- **WHEN** a class registers a signal under a name neither it nor its ancestors carry
- **THEN** the signal is registered with the engine

### Requirement: A parameter description carries the declaring class as its class name

Each parameter description of a registered signal SHALL carry the declaring class's own
name in its class-name field, whatever the parameter's variant type, and the hint string
field SHALL be empty.

#### Scenario: Non-object parameter

- **WHEN** a signal declares a string parameter
- **THEN** that parameter's description reports the declaring class in its class-name field

#### Scenario: Author supplies no class qualification

- **WHEN** a signal parameter is declared with only a variant type and a name
- **THEN** no hint string is sent for it

### Requirement: A constant name is unique within its class

A class SHALL expose each integer constant name at most once, regardless of which enum
the constant is grouped under, and a repeated constant name SHALL be rejected before the
engine is asked.

#### Scenario: Same value name in two enums

- **WHEN** a class binds a constant name inside one enum and then binds the same name inside a second enum
- **THEN** the second binding fails fatally naming the class
- **AND** the first binding is the one the engine keeps

#### Scenario: Constant with no enum

- **WHEN** a class binds a constant with no enum name
- **THEN** the constant is readable from scripting without an enum qualification

### Requirement: An enum groups the constants that name it

Constants bound under the same non-empty enum name SHALL appear in the class's metadata
as one enum member holding those constants. The engine-visible enum name SHALL be the
portion of the supplied enum name following its first separator, so a qualified name is
accepted and shortened rather than rejected.

#### Scenario: Two values of one enum

- **WHEN** a class binds two constants under the same enum name
- **THEN** the class reports one enum member containing both constants
- **AND** each constant reports that enum as its owning enum

#### Scenario: Qualified enum name

- **WHEN** a class binds a constant under an enum name qualified with the class name
- **THEN** the enum is registered under the trailing portion of that name

#### Scenario: Enum-less constant

- **WHEN** a class binds only constants with no enum name
- **THEN** the class reports no enum member for them

### Requirement: A bitfield flag records its enum as a bitfield

A bitfield flag SHALL be bound under the enum name its caller supplies, and that enum
SHALL be reported as a bitfield. The supplied enum name SHALL reach the engine.

#### Scenario: Flag registered with its enum name

- **WHEN** a class binds two flags under one bitfield enum name
- **THEN** the class reports that enum as present
- **AND** the class reports that enum as a bitfield
- **AND** each flag reports that enum as its owning enum

#### Scenario: Flags stay readable unqualified

- **WHEN** flags have been bound under a bitfield enum name
- **THEN** each flag remains readable without the enum qualification, with the value it was bound with

### Requirement: Property groups and subgroups are published in registration order

A group or subgroup registration SHALL deliver its display name and its prefix to the
engine, and placement depends on registration order: a group or subgroup affects the
properties registered after it whose names begin with that prefix.

#### Scenario: Subgroup inside a group

- **WHEN** a class registers a group, then a subgroup with a longer prefix, then a property whose name begins with that longer prefix
- **THEN** the engine reports the property under the subgroup of that group

#### Scenario: Property outside any prefix

- **WHEN** a property's name begins with none of the registered prefixes
- **THEN** the property is reported ungrouped

### Requirement: Metadata requests against an unregistered class fail loudly

A signal, constant, enum, bitfield, or group request naming a class that has no
registration SHALL fail loudly rather than dereferencing a missing registration, and the
engine SHALL receive no metadata for a class it does not know.

#### Scenario: Signal on an unknown class

- **WHEN** a signal is registered for a class that was never registered
- **THEN** the request fails loudly and the engine is not called

#### Scenario: Constant on an unknown class

- **WHEN** a constant is bound for a class that was never registered
- **THEN** the request fails loudly and the engine is not called

#### Scenario: Group on an unknown class

- **WHEN** a property group is registered for a class that was never registered
- **THEN** the request fails loudly and the engine is not called
