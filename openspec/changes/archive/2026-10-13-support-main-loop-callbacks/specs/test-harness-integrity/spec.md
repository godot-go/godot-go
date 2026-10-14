# Spec Delta

## ADDED Requirements

### Requirement: Main-loop shutdown marker is mandatory

The test gate SHALL require the main-loop shutdown marker emitted by the test
extension in the run output. Its absence SHALL fail the target with a message
naming the missing shutdown callback marker, distinct from every other
failure mode's message.

#### Scenario: Clean summary but no shutdown marker

- **WHEN** Godot exits with a clean driver summary and no leak markers but the
  log lacks the shutdown marker
- **THEN** the target exits non-zero with a message naming the missing
  main-loop shutdown marker

#### Scenario: Marker present does not fail the gate

- **WHEN** the log contains the shutdown marker
- **THEN** the gate does not fail on the basis of the marker

## MODIFIED Requirements

### Requirement: A Genuinely Clean Run Passes

The test target SHALL exit zero when the extension loaded, the driver summary banner is present with a positive total test count, the driver-reported failure count is zero, the main-loop shutdown marker is present, and no leak markers are present.

#### Scenario: All integrity checks satisfied

- **WHEN** the log shows the extension loaded, `PASSES: 1077`, `FAILURES: 0`, the shutdown marker, and no leak markers
- **THEN** the target exits zero
