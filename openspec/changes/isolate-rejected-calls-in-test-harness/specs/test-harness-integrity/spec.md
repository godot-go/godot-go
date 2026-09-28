# Spec Delta

## ADDED Requirements

### Requirement: A deliberately-rejected call shall not strand what follows it

The test driver SHALL provide a primitive for making a call that is expected to
be rejected by the engine, such that the script-error abandonment the engine
produces is contained within the primitive and the calling function continues to
execute. Assertions and resource cleanup written after the primitive SHALL run.

#### Scenario: Cleanup after an expected rejection still runs

- **WHEN** a test makes a call through the primitive and then frees the objects
  it created
- **THEN** the frees execute and the run reports no leaked engine objects

#### Scenario: Assertions after an expected rejection still execute

- **WHEN** a test writes assertions after the primitive
- **THEN** those assertions are evaluated and counted, rather than silently
  skipped while the suite still reports green

#### Scenario: Inline rejection without the primitive is still a hazard worth naming

- **WHEN** a deliberately-rejected call is written inline with code following it
  in the same function
- **THEN** the following code does not run, and the driver's documentation names
  this as the reason the primitive exists

### Requirement: A rejection that stops being a rejection shall be detected

The primitive SHALL report whether the call actually completed. A call expected
to be rejected that instead succeeds SHALL be recorded as a test failure rather
than passing quietly.

#### Scenario: The engine stops rejecting and the test notices

- **WHEN** a call wired through the primitive completes successfully because the
  rejection was removed or the argument became acceptable
- **THEN** the test fails, rather than the suite losing coverage without saying so

#### Scenario: A genuine rejection is reported as such

- **WHEN** the engine rejects the call
- **THEN** the primitive reports rejection and the caller's subsequent
  assertions about side effects are meaningful

### Requirement: The primitive's containment shall be pinned by a self-check

The behaviour the primitive relies on -- that a script error unwinds only the
frame in which it occurs -- SHALL be verified by an executable check, so that a
change in that behaviour fails loudly instead of leaving the primitive silently
containing nothing.

#### Scenario: Whole-stack unwinding is caught

- **WHEN** a Godot version changes script-error handling to unwind the whole
  call stack
- **THEN** the self-check fails and identifies the primitive as the cause, rather
  than the suite quietly stopping mid-run
