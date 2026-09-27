# Spec Delta

## Purpose

Governs the lifetime of Go memory pinned for crossings between Go and Godot at the GDExtension boundary, so that scratch memory used by a single call is released when that call completes while values the engine retains beyond the call stay valid for as long as the engine holds them.

## ADDED Requirements

### Requirement: Call-Scoped Pins Are Released When The Call Completes
Every Go object pinned solely to satisfy a single native call — the return slot, the owner pointer cell, the argument pointer slice, and each argument slot — SHALL be unpinned before the Go function performing the call returns. Generated method bodies SHALL acquire a dedicated per-call pinner and release it with `defer Unpin()` at the top of the body.

#### Scenario: Return wrapper is unpinned after the call
- **WHEN** Go calls `CollisionShape2D.GetShape()` and the call returns normally
- **THEN** the `Shape2DImpl` return wrapper that was pinned during the call is no longer pinned once the Go method has returned

#### Scenario: Argument cells are unpinned after the call
- **WHEN** a generated method with three arguments completes its ptrcall
- **THEN** the argument slice backing array and all three argument cells are unpinned before the Go method returns

### Requirement: Repeated Calls Do Not Accumulate Retained Memory
Executing the same generated call N times SHALL NOT leave any object pinned after the calls complete, and heap retention SHALL NOT grow proportionally to N.

#### Scenario: Dropped return wrappers become collectable
- **WHEN** a test calls a generated object-returning method, drops the returned value, and forces garbage collection
- **THEN** a weak reference to the dropped return wrapper reports empty, proving the call left nothing pinned

#### Scenario: Frame-shaped loop stays flat
- **WHEN** a loop performs thousands of mixed generated calls and the test forces garbage collection between samples
- **THEN** live heap after collection stays within a constant band regardless of iteration count

### Requirement: Values That Must Outlive The Call Remain Valid
A Go value that C retains a reference to beyond the call that created it SHALL be pinned on a pinner whose lifetime matches that retention, and SHALL NOT be unpinned while any C-side reference remains.

#### Scenario: Method metadata survives unrelated per-call unpins
- **WHEN** a method is registered, pinning its `GoMethodMetadata` argument and return `PropertyInfo` `StringName` and hint `String` objects, and many unrelated generated calls subsequently run and unpin their own scratch
- **THEN** any later call or Godot reflection that reads those names observes valid, uncorrupted data

#### Scenario: Returned object stays usable after the call's pinner is released
- **WHEN** a generated method returns a reference-counted object to its Go caller and the call's per-call pinner has already run `Unpin()`
- **THEN** the caller can still use the returned object — read its properties and pass it into further Godot calls — with no corruption

### Requirement: Generated Call Paths Use The Scoped Pattern Uniformly
Every generated per-call body SHALL render the scoped pinner pattern, and no generated call body SHALL pin to a package-global pinner. Registration and initialization code MAY continue to use the package pinner for program-lifetime retention.

#### Scenario: Rendered body declares its own pinner
- **WHEN** `make generate` renders a generated method body
- **THEN** the body declares a local `runtime.Pinner` with `defer pinner.Unpin()` as its first statement and every `Pin` in the body targets that local pinner

#### Scenario: No global pins remain in generated call bodies
- **WHEN** the generated call-path sources are scanned after the change
- **THEN** no `pnr.Pin` call appears inside any generated method, constructor, utility-function, or variant-conversion body

### Requirement: The Engine Exit Leak Check Stays Clean
A test run that exercises object-returning generated calls SHALL shut the engine down with no leaked instances reported by the engine's exit-time leak check.

#### Scenario: GetShape loop exits clean
- **WHEN** the demo calls `GetShape()` in a loop and the project exits
- **THEN** Godot reports no `Leaked instance: CircleShape2D` and no leaked `GodotShape2D` RID allocations at exit

#### Scenario: Full suite keeps the leak gate green
- **WHEN** the full test suite runs after this change
- **THEN** the exit-time leak gate passes as it did before object-returning calls were exercised
