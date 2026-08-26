## Purpose

Lets Go classes registered as GDExtension classes receive Godot notifications: the engine's `notification_func` creation-info callback dispatches to a Go-implemented `_notification` virtual so classes can react to lifecycle, visibility, theme, and editor events (`NOTIFICATION_*`).

## ADDED Requirements

### Requirement: Notification Dispatches To The Class Virtual
When Godot delivers a notification to an instance of a Go-registered class, the binding SHALL dispatch to that class's registered `_notification` virtual, passing the notification code and the reversed flag.

#### Scenario: Registered virtual receives the notification code
- **WHEN** a `NOTIFICATION_*` event is delivered to an instance whose `_notification` virtual is registered
- **THEN** the Go method executes with the matching notification code

### Requirement: Unimplemented Notifications Are Silent No-Ops
A Go-registered class that does not register `_notification` SHALL accept notifications without error, log output, or side effects.

#### Scenario: Missing virtual absorbs high-frequency events
- **WHEN** repeated notifications are delivered to an instance of a class that never registered `_notification`
- **THEN** delivery completes silently with no observable behavior or per-call logging
