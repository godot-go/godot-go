# Spec Delta

## ADDED Requirements

### Requirement: A deliberately-rejected call shall not strand what follows it

The test driver SHALL provide a primitive for making a call that is expected to
be rejected by the engine, such that the calling function continues to execute.
Assertions and resource cleanup written after the primitive SHALL run.

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

The primitive SHALL report whether the call was rejected, so that a call expected
to be rejected which instead succeeds is recorded as a test failure rather than
passing quietly.

The report is derived from the call's return value, which constrains the
primitive to value-returning targets: a void method yields the same null that a
rejection does, so a void target cannot be judged this way and its caller must
assert an observable side effect instead.

#### Scenario: The engine stops rejecting and the test notices

- **WHEN** a call wired through the primitive completes successfully because the
  rejection was removed or the argument became acceptable
- **THEN** the test fails, rather than the suite losing coverage without saying so

#### Scenario: A genuine rejection is reported as such

- **WHEN** the engine rejects the call
- **THEN** the primitive reports rejection and the caller's subsequent
  assertions about side effects are meaningful

#### Scenario: A mistyped method name is not mistaken for a rejection

- **WHEN** the primitive is given a method the target does not have
- **THEN** the test fails with a message identifying the missing method, rather
  than the typo reading as a rejection that worked

### Requirement: The rejection signal's basis shall be pinned by a self-check

The primitive infers rejection from the call yielding no value. That inference
SHALL be pinned by an executable check covering both outcomes -- one call that
is rejected and one that is accepted -- so a change in the engine's return
behaviour fails the suite instead of silently inverting the report.

#### Scenario: An accepted call is pinned as not-rejected

- **WHEN** the primitive is pointed at a call the engine accepts
- **THEN** the suite fails, confirming the report tracks acceptance rather than
  defaulting to whatever was set beforehand

#### Scenario: A rejected call is pinned as rejected

- **WHEN** the primitive is pointed at a call the engine rejects
- **THEN** the suite passes, confirming the report tracks rejection
