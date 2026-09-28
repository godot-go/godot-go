# Spec Delta

## Purpose

Governs how a varcall reacts when an argument clears variant-level validation but cannot be decoded into the bound Go parameter type, so that a caller which supplies a real-but-wrong object is rejected through the engine's call-error channel instead of terminating the host process.

## ADDED Requirements

### Requirement: Decode Failure Is Reported As An Argument Error Not A Crash

When a varcall argument passes variant-level validation but cannot be decoded into the bound Go parameter type, the system SHALL reject the call by writing `INVALID_ARGUMENT` into the engine's call-error slot with the `argument` field set to the offending zero-based index, and SHALL return without invoking the bound method. The process SHALL remain alive.

#### Scenario: Object of the wrong class is rejected, not fatal

- **WHEN** a caller passes a `Label` to a bound method whose parameter is declared as `Sprite2D`
- **THEN** the engine receives an `INVALID_ARGUMENT` call error whose `argument` is that parameter's index
- **AND** the bound Go method body does not run
- **AND** the process remains alive

#### Scenario: Unresolvable object is rejected, not fatal

- **WHEN** a caller passes an object whose instance binding cannot be resolved to the declared parameter type
- **THEN** the engine receives an `INVALID_ARGUMENT` call error identifying that argument index
- **AND** the bound Go method body does not run
- **AND** the process remains alive

#### Scenario: First failing argument wins

- **WHEN** a call has more than one undecodable argument
- **THEN** the rejection reports the lowest-numbered failing argument index

### Requirement: Decode Failure Carries Position And Reason

A decode rejection SHALL surface a typed failure that names the failing argument index and a human-readable reason identifying the parameter type involved, so the rejection is diagnosable without reproducing the caller.

#### Scenario: Failure identifies the parameter type

- **WHEN** a decode failure is reported for an argument
- **THEN** the surfaced failure carries the argument index and a reason naming the declared parameter type that could not be satisfied

### Requirement: Rejected Decode Leaves A Defined Return

When a varcall is rejected at the decode stage, the engine's return slot SHALL hold a defined nil value rather than indeterminate memory, so the caller observes a well-formed failed call.

#### Scenario: Return slot initialised on decode rejection

- **WHEN** a varcall is rejected because an argument cannot be decoded
- **THEN** the engine's return value slot holds a defined nil value

### Requirement: Arguments Decoded Before A Failure Are Released

Decoding proceeds argument by argument, and earlier arguments may be decoded as owned copies. When a later argument fails, the system SHALL release every owned argument value already decoded in that call before returning the rejection, so a mid-call rejection leaks nothing.

#### Scenario: Container decoded before a failing object is released

- **WHEN** a call decodes an owned container argument successfully and then fails to decode a later object argument
- **THEN** the owned container is released exactly once as part of the rejection
- **AND** the rejection is still reported to the engine

#### Scenario: Reject path does not double-release

- **WHEN** a decode rejection unwinds previously decoded owned arguments
- **THEN** those values are not released a second time by the success-path cleanup, which does not run

### Requirement: Authoring Errors Remain Fatal

Failures a caller cannot induce SHALL stay fatal and SHALL NOT be converted into call errors. These include a missing constructor for a bound `Ref` parameter type and a Go parameter kind the decoder does not support at all.

#### Scenario: Missing Ref constructor is not downgraded to a call error

- **WHEN** a bound method declares a `Ref` parameter whose type has no registered constructor
- **THEN** the outcome remains a fatal authoring error rather than an `INVALID_ARGUMENT` rejection

#### Scenario: Unsupported parameter kind is not downgraded

- **WHEN** a bound method declares a parameter of a Go kind the varcall decoder does not handle
- **THEN** the outcome remains a fatal authoring error rather than an `INVALID_ARGUMENT` rejection

### Requirement: Decodable Calls Are Unaffected

Arguments that decode successfully SHALL marshal, dispatch, and encode their return exactly as before this capability existed. Adding the rejection path SHALL NOT change the behaviour of a well-formed call.

#### Scenario: Well-formed call unchanged

- **WHEN** a caller invokes a bound method with arguments that all decode into their declared parameter types
- **THEN** the bound method runs and its return is encoded as it was prior to this change

#### Scenario: Owned container cleanup still runs on success

- **WHEN** a well-formed call decodes owned container arguments and completes
- **THEN** those arguments are released after the return value has been encoded, as before this change
