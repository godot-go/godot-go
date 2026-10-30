## Purpose

Pins what engine→Go dispatch is allowed to cost per call — the ceremony around a handler, not the handler itself — and defines how the test harness detects that cost drifting upward or its measurements disappearing.

## Requirements

### Requirement: Class Callback Dispatch Identifies Its Target Without An Engine Round Trip

When the engine invokes a class callback — notification, property get, property set, property validation, property list, or revert query — the binding SHALL identify the target class from state recorded for that instance, without calling into the engine to learn the instance's class name and without copying a class name out of engine memory.

#### Scenario: Notification delivery asks the engine nothing about the class

- **WHEN** the engine delivers a notification to an instance of a registered Go class
- **THEN** dispatch resolves the target and runs without the engine being asked for the instance's class name and without a class name converted from an engine string

#### Scenario: Resolution change is behavior-preserving

- **WHEN** a notification, property get, property set, property validation, property list, or revert query is delivered to a live instance of a registered class
- **THEN** the same handler runs with the same arguments and the same return contract as before resolution was changed

### Requirement: Teardown Deliveries Decline On Wrapper State

A class callback delivered to an instance whose Go wrapper has begun being freed SHALL decline without invoking the user handler, without panic, and without per-call logging. Declining SHALL NOT depend on the class name the engine reports for that instance.

#### Scenario: Notification arriving during destruction is absorbed

- **WHEN** a notification is delivered to an instance whose wrapper has begun being freed
- **THEN** the delivery completes silently with no handler call, no panic, and no per-call log entry

#### Scenario: Revert query during destruction answers negatively

- **WHEN** a revert query is routed to an instance whose wrapper has begun being freed
- **THEN** the query completes reporting no revert support, with no panic and no per-call log entry

### Requirement: Dispatch Shapes Carry Allocation Budgets That Fail On Regression

Every dispatch shape covered by the cost baseline SHALL carry a per-call allocation budget, and the test target SHALL fail when a shape exceeds its budget. Budgets SHALL tolerate measurement noise through a floor and a margin, and each run SHALL report the measured figure beside its budget.

#### Scenario: A compiled dispatch shape holds its zero budget

- **WHEN** the baseline measures a dispatch shape that performs no per-call reflection and no per-call allocation
- **THEN** the measured figure is zero and the target passes

#### Scenario: A reintroduced allocation fails the target

- **WHEN** a dispatch shape that was holding its budget starts allocating per call
- **THEN** the target fails naming the shape and both the measured figure and the budget

#### Scenario: A budget that no longer measures anything is detected

- **WHEN** a budgeted dispatch shape stops being measured, so its assertion would pass vacuously
- **THEN** the target fails rather than reporting a clean run

### Requirement: Dispatch Timing Is Recorded Without Gating

The runtime suite SHALL record per-shape dispatch timing with the engine running and print the figures with the run's diagnostic counters. The output gate SHALL require that the recording is present and SHALL NOT fail on the value of a timing.

#### Scenario: A clean run reports timings

- **WHEN** the runtime suite completes successfully
- **THEN** the log carries one timing record per covered dispatch shape

#### Scenario: Slower dispatch does not fail the gate on its own

- **WHEN** dispatch timings rise while every allocation budget still holds
- **THEN** the gate passes, leaving the figures to be read
