# Spec Delta

## MODIFIED Requirements

### Requirement: Revert Callbacks Tolerate Unregistered Instances During Teardown

When Godot routes a revert query to an instance that cannot be addressed — because its Go wrapper has begun being freed, or because its class has no registration — the binding SHALL answer negatively rather than panicking or logging. Answering negatively SHALL NOT depend on the class name the engine reports for the instance.

#### Scenario: Query under a parent-class presentation reports no revert support

- **WHEN** a revert query is routed to a destructing instance — the presentation the engine makes under a parent class name — whose wrapper has begun being freed
- **THEN** the query completes and reports no revert support

#### Scenario: Query for a class with no registration reports no revert support

- **WHEN** a revert query is routed to an instance whose class has no registration
- **THEN** the query completes and reports no revert support
