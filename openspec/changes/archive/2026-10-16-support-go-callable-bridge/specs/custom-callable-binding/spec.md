# Spec Delta

## Purpose

Lets Go code hand engine-native `Callable` objects backed by Go closures to
GDScript, with defined invocation, identity, failure-reporting, and cleanup
behavior.

## ADDED Requirements

### Requirement: Go closure becomes an engine Callable
The system SHALL expose `core.NewGoCallable` which accepts a Go function of
signature `func(args []builtin.Variant) (builtin.Variant, error)` and returns
a `builtin.Callable` that is valid (`is_valid()` true) and usable anywhere a
GDScript `Callable` is accepted, including as a return value from a bound Go
class method.

#### Scenario: Create and hand to GDScript
- **WHEN** a bound Go method calls `NewGoCallable` and returns the result to
  GDScript
- **THEN** GDScript receives a Callable whose `is_valid()` returns true

#### Scenario: Invoke the callable from GDScript
- **WHEN** GDScript calls `cb.call(3, 4)` on a Callable created from a Go
  closure that returns the sum of its integer arguments
- **THEN** the closure receives both arguments and `cb.call` evaluates to the
  sum

### Requirement: Call arguments are borrowed for the call duration
The system SHALL pass the GDScript-supplied arguments to the Go closure in
order, regardless of count, as Variant values that remain valid until the
invocation returns. A value copied by the closure during the invocation MUST
remain valid after it returns.

#### Scenario: Arbitrary argument count reaches the closure
- **WHEN** GDScript invokes a Go Callable with five arguments
- **THEN** the closure receives five arguments in the supplied order

#### Scenario: Copy stored beyond the call stays valid
- **WHEN** a closure copies a String argument during the call and GDScript
  later invokes the closure again to read the stored copy
- **THEN** the stored copy reads back equal to the original argument

### Requirement: Go Callable identity is per-closure
Each `NewGoCallable` invocation SHALL produce a Callable distinct from every
other Go Callable, and copies of a Callable SHALL remain equal to the
original, so signal connection dedup follows Callable identity.

#### Scenario: Same callable connected twice dedups
- **WHEN** GDScript connects the same Go Callable to one signal twice
- **THEN** the signal reports a single connection and the closure body runs
  once per emission

#### Scenario: Distinct callables are distinct
- **WHEN** GDScript connects two Go Callable values created by separate
  `NewGoCallable` invocations to the same signal
- **THEN** two connections exist and one emission runs both closures

#### Scenario: Deferred call reuses the same path
- **WHEN** GDScript calls `cb.call_deferred(...)`
- **THEN** the same Go closure runs on the main thread with the same argument
  semantics and its return value is discarded

### Requirement: Go Callable string representation
The system SHALL report a custom string representation for Go Callables via
the engine's to-string mechanism, marked valid so the engine uses it instead
of a generic fallback.

#### Scenario: str(cb) shows the Go label
- **WHEN** GDScript evaluates `str(cb)` on a Go Callable
- **THEN** the result is the Go Callable label rather than an unconfigured
  fallback rendering

### Requirement: Closure failures are reported, never crash the engine
When a Go closure returns a non-nil error, the system SHALL log the error,
signal the invocation as failed through the engine call-error channel so the
GDScript call raises a script error, and keep the engine running. When a Go
closure panics, the system SHALL recover the panic, log it, and report the
invocation failed in the same way.

#### Scenario: Closure returns an error
- **WHEN** GDScript calls a Go Callable whose closure returns an error
- **THEN** the call fails with a script-side call error, the error is logged,
  and the process continues

#### Scenario: Closure panics
- **WHEN** GDScript calls a Go Callable whose closure panics
- **THEN** the panic is recovered and logged, the call fails, and subsequent
  calls on the same Callable still work

### Requirement: Reentrant engine calls from closures
The system SHALL support a Go closure invoking engine functionality (for
example calling methods on captured engine objects) while running inside a
Callable invocation, including nested calls back through Go-bound methods.

#### Scenario: Closure calls back into the engine
- **WHEN** a Go Callable closure calls a bound Go class method on a captured
  object during invocation
- **THEN** the nested call completes and its result is observable from the
  outer call

### Requirement: Registry entries are released deterministically
The system SHALL release each Go Callable's registry entry when the engine
destroys the last reference to it, without waiting for Go garbage collection,
so no registry entry outlives the Callable that owns it and the test-suite
leak gate stays clean.

#### Scenario: Dropping all references frees the entry
- **WHEN** GDScript drops the last reference to a Go Callable and garbage
  collection runs
- **THEN** the registry entry count decreases without any further Go-side
  action

#### Scenario: Shutdown with live and dead callables
- **WHEN** the test project shuts down after creating, invoking, and dropping
  Go Callables
- **THEN** no ObjectDB leak is reported and the run passes the green gate
