# Spec Delta

## ADDED Requirements

### Requirement: Colour Codes Are Stripped On Every Supported Platform

The test gate SHALL remove ANSI colour sequences from the captured run log using only POSIX-portable shell facilities, so that parsing the driver summary produces the same verdict under GNU and BSD userlands.

#### Scenario: Clean run captured on a BSD userland

- **WHEN** the log wraps the summary counts in colour escapes and the gate runs with BSD `sed`
- **THEN** the gate parses `PASSES` and `FAILURES` and exits zero, rather than failing as a missing driver summary

#### Scenario: Clean run captured on a GNU userland

- **WHEN** the same colour-wrapped log is parsed with GNU `sed`
- **THEN** the gate reaches the same verdict as on the BSD userland

#### Scenario: A real failure survives the portable strip

- **WHEN** a colour-wrapped log reports `FAILURES: 1`
- **THEN** the gate still parses the count and fails the target with the driver-reported failure count
