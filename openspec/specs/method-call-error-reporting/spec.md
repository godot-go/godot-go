## Purpose

Governs how a varcall into a Go-bound method reacts to calls the caller cannot satisfy, so that a mismatched call is reported to the engine as a normal call error instead of terminating the host process.

## Requirements

### Requirement: Arity Mismatch Is Reported Instead Of Fatal

A varcall whose argument count the bound signature cannot satisfy SHALL be rejected by writing a call error into the engine's `GDExtensionCallError` slot and returning without invoking the bound method. A too-many-arguments call SHALL report `TOO_MANY_ARGUMENTS`; a call that cannot be satisfied even after applying trailing default arguments SHALL report `TOO_FEW_ARGUMENTS`. In both cases the `expected` field SHALL carry the **declared** argument count of the bound signature, matching godot-cpp's `call_with_variant_args_dv`, which reports the declared count for both arity errors. No argument-count mismatch SHALL terminate the process.

#### Scenario: Too many arguments

- **WHEN** a caller invokes a Go-bound method with more arguments than the signature declares
- **THEN** the engine receives a `TOO_MANY_ARGUMENTS` call error whose `expected` equals the declared argument count
- **AND** the bound Go method body does not run
- **AND** the process remains alive

#### Scenario: Too few arguments beyond available defaults

- **WHEN** a caller invokes a Go-bound method with fewer arguments than the signature requires after accounting for trailing default arguments
- **THEN** the engine receives a `TOO_FEW_ARGUMENTS` call error whose `expected` equals the declared argument count
- **AND** the bound Go method body does not run
- **AND** the process remains alive

#### Scenario: Defaults satisfy a short call

- **WHEN** a caller omits only arguments that have bound default values
- **THEN** the call succeeds using those defaults and is not reported as an arity error

### Requirement: Type Mismatch Is Reported With Argument Position

A varcall whose argument cannot be converted to the bound Go parameter type SHALL be rejected by writing `INVALID_ARGUMENT` into the `GDExtensionCallError` slot, setting the `argument` field to the offending zero-based index and the `expected` field to the expected variant type, and returning without invoking the bound method. Type conversion on the reject path SHALL NOT terminate the process.

#### Scenario: Wrong type for a parameter

- **WHEN** a caller passes a value whose variant type cannot convert to a bound parameter's Go type
- **THEN** the engine receives an `INVALID_ARGUMENT` call error identifying the failing argument index and expected type
- **AND** the bound Go method body does not run
- **AND** the process remains alive

### Requirement: Rejected Calls Leave A Defined Return

When a varcall is rejected before invoking the bound method, the callback SHALL still leave the engine's return slot initialized (a nil/void value) rather than uninitialized, so the engine observes a well-formed failed call.

#### Scenario: Return slot initialized on rejection

- **WHEN** a varcall is rejected for arity or type
- **THEN** the engine's return value slot holds a defined nil value rather than indeterminate memory

### Requirement: Internal Faults Remain Fatal While Caller Errors Do Not

Only faults the caller cannot induce — absent method user data or a null target instance — SHALL remain fatal. Any mismatch a caller can induce by choosing arguments or argument count SHALL be reported as a call error, never as a process-terminating fault.

#### Scenario: Caller-induced error is non-fatal

- **WHEN** a caller induces an argument count or type mismatch through its call
- **THEN** the outcome is a reported call error, not a process termination

### Requirement: Accepted Calls Are Unaffected

Introducing error reporting SHALL NOT change the dispatch, argument decoding, or return encoding of a call that satisfies the bound signature. A well-formed call behaves exactly as before this capability existed.

#### Scenario: Well-formed call unchanged

- **WHEN** a caller invokes a Go-bound method with the correct arity and argument types
- **THEN** the bound method runs and its return is encoded as it was prior to this change

### Requirement: Rejected Calls Emit A Debug Diagnostic

Every varcall rejection SHALL produce a debug-level log entry identifying the bound method, the reported error kind, and the `argument` and `expected` values written to the engine's call error slot, so a rejection is observable in the extension's own debug logs and not only through the engine's caller. The diagnostic SHALL NOT change the rejection behavior.

#### Scenario: Arity rejection is logged

- **WHEN** a varcall is rejected for a too-many or too-few argument count
- **THEN** a debug-level log entry is emitted naming the bound method and the error kind
- **AND** the entry carries the `expected` value written to the call error

#### Scenario: Type rejection is logged

- **WHEN** a varcall is rejected for a non-convertible argument type
- **THEN** a debug-level log entry is emitted naming the bound method and the error kind
- **AND** the entry carries the failing `argument` index and the `expected` variant type
