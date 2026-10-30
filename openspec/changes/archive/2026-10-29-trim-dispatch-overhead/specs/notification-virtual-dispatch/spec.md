# Spec Delta

## MODIFIED Requirements

### Requirement: Notification Dispatch Survives Teardown Deliveries

When Godot delivers a notification to an instance whose Go wrapper has begun being freed — observed as a pre-delete delivery on a destructing extension instance — the binding SHALL absorb it silently. Absorption SHALL NOT depend on the class name the engine reports for the instance.

#### Scenario: Pre-delete delivery under the parent class name is absorbed

- **WHEN** a notification arrives for a destructing instance — the delivery the engine presents under a parent class name — whose wrapper has begun being freed
- **THEN** delivery completes silently with no panic and no per-call logging

#### Scenario: A live instance keeps receiving notifications while the extension shuts down

- **WHEN** a notification arrives for an instance whose wrapper has not begun being freed
- **THEN** the class's `_notification` virtual runs as registered
