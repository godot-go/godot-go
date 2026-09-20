# method-default-arguments (delta)

## Purpose

Governs how a Go-bound method's default argument values map to its parameters and are applied when a caller omits trailing arguments, across both the engine varcall path and direct Go calls to the exported call entry.

## ADDED Requirements

### Requirement: Bound Defaults Map To Trailing Parameters

A bound method's default-value array SHALL correspond to the **last** N parameters of the signature, where N is the number of bound defaults. When a caller omits a trailing argument that has a bound default, the bound value SHALL be delivered to that parameter by position, matching Godot's trailing-default convention.

#### Scenario: Partial default fills the correct trailing parameter

- **WHEN** a method declares two parameters and binds a default for only the last one, and a caller supplies just the first argument
- **THEN** the first parameter receives the caller's value and the second receives the bound default
- **AND** the bound method body runs

#### Scenario: All-defaulted binding keeps identical behavior

- **WHEN** a method binds a default for every parameter and a caller supplies zero, some, or all arguments
- **THEN** each omitted trailing parameter receives its own bound default and each supplied parameter receives the caller's value, exactly as before this change

### Requirement: Short Call Satisfiability Tracks The Default Count

A call that omits no more trailing arguments than there are bound defaults SHALL be satisfiable and dispatch with the defaults applied. A call that omits more trailing arguments than there are bound defaults SHALL NOT be satisfiable.

#### Scenario: Omitting exactly the defaulted trailing arguments succeeds

- **WHEN** a method declares three parameters with one bound trailing default and a caller supplies the two leading arguments
- **THEN** the call is satisfiable and dispatches with the third parameter set to its bound default

#### Scenario: Omitting beyond the available defaults is unsatisfiable

- **WHEN** a method declares three parameters with one bound trailing default and a caller supplies only one argument
- **THEN** the call is not satisfiable (reported as a too-few arity error by the call-error capability)

### Requirement: Binding More Defaults Than Parameters Is Rejected

Registering a method whose bound default count exceeds its declared parameter count SHALL be rejected at bind time with a clear diagnostic, rather than producing a binding with unreachable or misaligned defaults.

#### Scenario: Excess defaults rejected at registration

- **WHEN** a binding is registered with more default values than the method declares parameters
- **THEN** registration fails with a diagnostic naming the offending method

### Requirement: Direct Go Callers Are Validated

A direct call to the exported call entry that leaves a parameter slot unfilled — bypassing the varcall path's prior arity validation — SHALL be reported as a fatal misuse identifying the bound method and the unfilled slot, rather than silently delivering a zero-value argument. The engine varcall path, which validates before dispatch, SHALL NOT trigger this report.

#### Scenario: Direct call with too few arguments is reported

- **WHEN** a Go caller invokes the exported call entry with too few arguments to fill the signature even after applying defaults
- **THEN** the call fails with a diagnostic naming the bound method and the unfilled slot index
- **AND** the zero-value-argument silent fill no longer occurs

#### Scenario: Varcall path does not trigger the direct-call report

- **WHEN** the engine varcall path dispatches a call that passed arity validation
- **THEN** no unfilled-slot report is raised, because every slot is fillable
