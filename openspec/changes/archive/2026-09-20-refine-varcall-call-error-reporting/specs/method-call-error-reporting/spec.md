# method-call-error-reporting (delta)

## MODIFIED Requirements

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

## ADDED Requirements

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
