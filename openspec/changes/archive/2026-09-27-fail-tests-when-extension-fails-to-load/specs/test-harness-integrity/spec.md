# Spec Delta

## Purpose

Governs the integrity of the headless test target (`make test`): it must guarantee a current extension library and must fail loudly, with a distinct actionable message, whenever the harness itself did not genuinely run the test suite — while preserving existing leak detection and passing genuinely clean runs.

## ADDED Requirements

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
