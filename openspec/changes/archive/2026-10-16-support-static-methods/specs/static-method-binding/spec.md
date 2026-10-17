# Spec Delta

## Purpose

Lets Go code expose package functions as Godot class static methods that
GDScript and the engine can call without an object instance, with the same
marshalling, defaults, and error reporting as instance method binds.

## ADDED Requirements

### Requirement: Go package functions bind as class static methods
`core.ClassDBBindStaticMethod` SHALL register a Go package function against
the registered class of the provided type, flagged with the engine's static
method flag, so that GDScript can call the method through the class path
without any instance existing.

#### Scenario: Static call through the class path
- **WHEN** GDScript invokes the bound static method as `Example.<name>(args)`
  with no `Example` instance alive
- **THEN** the Go function executes with the supplied arguments and its
  return value is delivered to the caller

#### Scenario: Typed and dynamic invocations both dispatch
- **WHEN** the same bound static method is invoked once from statically
  typed GDScript expressions and once through dynamic/vararg call paths
- **THEN** both invocations execute the same Go function with correctly
  decoded arguments

#### Scenario: Default arguments on a static method
- **WHEN** GDScript calls the static method omitting trailing arguments
  that were registered with default values
- **THEN** the Go function receives the registered default values

### Requirement: Null-instance dispatch for static binds
The varcall and ptrcall dispatch into Go SHALL treat a null instance
pointer as valid when the invoked bind is static, executing the bound
function without receiver injection, and SHALL remain fatal on a null
instance for non-static binds.

#### Scenario: Engine invokes a static bind
- **WHEN** the engine calls a static bind and passes a null instance
  pointer
- **THEN** the bound function runs and returns its value instead of the
  process aborting with a null-instance fault

#### Scenario: Null instance on an instance bind stays fatal
- **WHEN** the engine calls a non-static bind with a null instance pointer
- **THEN** the binding reports a fatal internal fault as it does today

### Requirement: Static marshalling matches instance marshalling
Static binds SHALL reuse the existing argument decoding, argument type
metadata, return-value conversion and ownership, and caller-error reporting
behaviors unchanged; the only difference from an instance bind is the
absence of a receiver.

#### Scenario: Static method takes an object argument
- **WHEN** GDScript passes an Object to a bound static method
- **THEN** the Go function receives a usable reference to that object via
  the standard object-argument decoding

#### Scenario: Static method returns an object
- **WHEN** a bound static method returns an Object-derived value
- **THEN** the value crosses to GDScript under the same ownership rules as
  an instance-method return

#### Scenario: Argument count mismatch on a static method
- **WHEN** GDScript calls a bound static method with the wrong number of
  arguments
- **THEN** the engine receives a call error through the standard
  error-reporting channel rather than a process abort

### Requirement: Static bind-time validation
The binding API SHALL reject invalid static registrations fatally at bind
time: a non-callable value as the function, and a method name that already
exists on the class (from an instance, virtual, or prior static bind).

#### Scenario: Non-function value supplied
- **WHEN** `ClassDBBindStaticMethod` is called with a value that is not a
  Go function
- **THEN** binding fails fatally naming the class and method

#### Scenario: Duplicate method name
- **WHEN** a static bind uses a Godot method name already registered on
  the same class
- **THEN** binding fails fatally and the previous registration is left
  intact

#### Scenario: Static introspection sees the method
- **WHEN** `ClassDB.class_has_method` queries the class for the bound
  static method with no instance
- **THEN** the query reports the method as present on the class
