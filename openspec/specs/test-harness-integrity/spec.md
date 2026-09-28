# test-harness-integrity Specification

## Purpose
Governs the integrity of the headless test target (`make test`): it must guarantee a current extension library and must fail loudly, with a distinct actionable message, whenever the harness itself did not genuinely run the test suite — while preserving existing leak detection and passing genuinely clean runs. It also governs the driver primitives used to induce engine-level failures, so that a deliberately rejected call cannot silently strand the assertions and cleanup written after it.

## Requirements

### Requirement: Test Target Guarantees A Current Extension Library

The test target SHALL ensure the test extension library is built and current before launching Godot, by depending on the build target.

#### Scenario: Run immediately after the library was cleaned

- **WHEN** `make generate` (which runs `clean` and deletes `test/demo/lib/libgodotgo-*`) has just completed and `make test` is invoked
- **THEN** the extension library is rebuilt before Godot starts, and the run executes against the current binary

### Requirement: Extension Load Failure Fails The Target With A Distinct Message

The test gate SHALL fail the target whenever the run output indicates the extension could not be loaded, and the emitted message SHALL explicitly identify an extension load failure, distinct from every other failure mode's message.

#### Scenario: Built library missing

- **WHEN** the shared library is absent and the output contains `Can't open dynamic library` or `GDExtension dynamic library not found`
- **THEN** the target exits non-zero and prints a message naming the extension load failure

#### Scenario: Library present but unloadable

- **WHEN** the library file exists but Godot fails to open or register it as a GDExtension
- **THEN** the target exits non-zero with the same extension-load-failure message

### Requirement: Test Script Parse Error Fails The Target

The test gate SHALL fail the target with a message identifying a test-script parse failure whenever the run output indicates the GDScript test driver failed to load or parse.

#### Scenario: Driver script fails to parse

- **WHEN** the output contains `Failed to load script` with `Parse error` for the test driver script
- **THEN** the target exits non-zero and prints a message naming the test-script parse error

### Requirement: Driver Summary Banner Is Mandatory

The test gate SHALL require the driver's `TESTS FINISHED` summary banner in the run output; its absence SHALL be treated as a harness failure with a distinct message.

#### Scenario: Driver never reached the summary

- **WHEN** Godot exits and the log contains no `TESTS FINISHED` banner
- **THEN** the target exits non-zero with a message that the test driver did not report a summary

### Requirement: Zero Tests Executed Fails The Target

The test gate SHALL parse the driver's `PASSES:` and `FAILURES:` counts and SHALL fail the target with a distinct message when their sum is zero.

#### Scenario: Summary reports no assertions ran

- **WHEN** the summary banner shows `PASSES: 0` and `FAILURES: 0`
- **THEN** the target exits non-zero with a message that zero tests executed

### Requirement: Driver-Reported Failures Fail The Target With The Count

The test gate SHALL fail the target with a message including the driver-reported failure count whenever the parsed `FAILURES` count is greater than zero.

#### Scenario: One assertion failed

- **WHEN** the summary banner shows `FAILURES: 1`
- **THEN** the target exits non-zero and the message states that the driver reported 1 failure

### Requirement: Leak Detection Is Preserved

The test gate SHALL preserve the existing leak check: output containing `ObjectDB instances were leaked` or `Leaked instance:` SHALL fail the target with the existing leaked-engine-objects message.

#### Scenario: Leaked instance reported

- **WHEN** the run output contains `Leaked instance:`
- **THEN** the target exits non-zero with the leaked-engine-objects message

### Requirement: Known Startup Warnings Do Not Fail The Target

The test gate SHALL NOT treat the known startup warnings for `get_variant_get_internal_ptr_func`, `callable_custom_get_user_data`, `classdb_register_extension_class_5`, or `classdb_register_extension_class_6` as a failure.

#### Scenario: Known warnings present in an otherwise clean run

- **WHEN** a run's log contains only those known startup warnings in addition to a clean driver summary
- **THEN** the gate does not fail on them

### Requirement: A Genuinely Clean Run Passes

The test target SHALL exit zero when the extension loaded, the driver summary banner is present with a positive total test count, the driver-reported failure count is zero, and no leak markers are present.

#### Scenario: All integrity checks satisfied

- **WHEN** the log shows the extension loaded, `PASSES: 1077`, `FAILURES: 0`, and no leak markers
- **THEN** the target exits zero

### Requirement: A deliberately-rejected call shall not strand what follows it

The test driver SHALL provide a primitive for making a call that is expected to be rejected by the engine, such that the calling function continues to execute. Assertions and resource cleanup written after the primitive SHALL run.

#### Scenario: Cleanup after an expected rejection still runs

- **WHEN** a test makes a call through the primitive and then frees the objects it created
- **THEN** the frees execute and the run reports no leaked engine objects

#### Scenario: Assertions after an expected rejection still execute

- **WHEN** a test writes assertions after the primitive
- **THEN** those assertions are evaluated and counted, rather than silently skipped while the suite still reports green

#### Scenario: Inline rejection without the primitive is still a hazard worth naming

- **WHEN** a deliberately-rejected call is written inline with code following it in the same function
- **THEN** the following code does not run, and the driver's documentation names this as the reason the primitive exists

### Requirement: A rejection that stops being a rejection shall be detected

The primitive SHALL report whether the call was rejected, so that a call expected to be rejected which instead succeeds is recorded as a test failure rather than passing quietly.

The report is derived from the call's return value, which constrains the primitive to value-returning targets: a void method yields the same null that a rejection does, so a void target cannot be judged this way and its caller must assert an observable side effect instead.

#### Scenario: The engine stops rejecting and the test notices

- **WHEN** a call wired through the primitive completes successfully because the rejection was removed or the argument became acceptable
- **THEN** the test fails, rather than the suite losing coverage without saying so

#### Scenario: A genuine rejection is reported as such

- **WHEN** the engine rejects the call
- **THEN** the primitive reports rejection and the caller's subsequent assertions about side effects are meaningful

#### Scenario: A mistyped method name is not mistaken for a rejection

- **WHEN** the primitive is given a method the target does not have
- **THEN** the test fails with a message identifying the missing method, rather than the typo reading as a rejection that worked

### Requirement: The rejection signal's basis shall be pinned by a self-check

The primitive infers rejection from the call yielding no value. That inference SHALL be pinned by an executable check covering both outcomes — one call that is rejected and one that is accepted — so a change in the engine's return behaviour fails the suite instead of silently inverting the report.

#### Scenario: An accepted call is pinned as not-rejected

- **WHEN** the primitive is pointed at a call the engine accepts
- **THEN** the suite fails, confirming the report tracks acceptance rather than defaulting to whatever was set beforehand

#### Scenario: A rejected call is pinned as rejected

- **WHEN** the primitive is pointed at a call the engine rejects
- **THEN** the suite passes, confirming the report tracks rejection
