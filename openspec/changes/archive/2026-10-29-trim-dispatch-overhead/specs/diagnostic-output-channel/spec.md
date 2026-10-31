# Spec Delta

## ADDED Requirements

### Requirement: Payloads Are Built Only When The Threshold Will Write The Entry

Work performed solely to describe an entry — formatting a value, converting a collection of arguments into text, or asking the engine to stringify a value — SHALL be performed only when the entry's threshold passes the configured entry threshold. An entry the threshold discards SHALL cost no more than the threshold comparison and its own message.

#### Scenario: Dispatch diagnostics at a suppressing threshold

- **WHEN** a bound method is dispatched while the configured entry threshold suppresses that entry
- **THEN** neither the call's arguments nor its return value are converted to text, and the engine is not asked to stringify any value belonging to the call

#### Scenario: Raising the threshold restores the entry unchanged

- **WHEN** the entry threshold is set to permit an entry that was previously suppressed
- **THEN** the entry is written carrying the same identification of the call, its arguments, and its return as before

#### Scenario: A value is described at most once per entry

- **WHEN** an entry describes the arguments and result of a single call
- **THEN** each argument and each returned value contributes at most one description
