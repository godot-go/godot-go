## Purpose

Lets Go classes registered as GDExtension classes implement Godot's inspector revert contract: the engine asks a class instance whether a property can revert to its default and what that default is, and the class's Go methods answer. This makes `Object.property_can_revert()`/`Object.property_get_revert()` work for Go-registered classes and enables the editor inspector's revert-to-default affordance.

## Requirements

### Requirement: Revert Query Dispatches To The Class Virtual
When Godot queries whether a property of a Go-registered class instance can revert, the binding SHALL dispatch to that class's registered `_property_can_revert` virtual, passing the queried property name, and report the Go method's boolean answer back to Godot.

#### Scenario: Property with revert support reports revertable
- **WHEN** `property_can_revert` is called on an instance whose `_property_can_revert` virtual returns true for the given property name
- **THEN** the engine receives true

#### Scenario: Property without revert support reports non-revertable
- **WHEN** `property_can_revert` is called on an instance whose `_property_can_revert` virtual returns false for the given property name
- **THEN** the engine receives false

### Requirement: Revert Value Dispatch Returns The Default Value
When Godot requests the revert value for a property, the binding SHALL dispatch to the class's registered `_property_get_revert` virtual, passing the property name; when the Go method reports it handles the query, the returned value SHALL be delivered to Godot in the output parameter.

#### Scenario: Handled query delivers the default value
- **WHEN** `property_get_revert` is called with a property name whose default the `_property_get_revert` virtual supplies (reporting handled)
- **THEN** the engine receives the supplied value

### Requirement: Unimplemented Revert Virtuals Report No Revert Support
A Go-registered class that does not implement the revert virtuals SHALL behave as if no property can revert: both queries return a negative result rather than failing or crashing.

#### Scenario: Missing virtuals yield negative answers
- **WHEN** `property_can_revert` or `property_get_revert` is called on an instance of a class that did not register the corresponding virtual
- **THEN** the queries complete without error and report no revert support (false / no value)

### Requirement: Revert Callbacks Tolerate Unregistered Instances During Teardown
When Godot routes a revert query to an instance whose reported class is not a registered GDClass — which occurs when a destructing extension instance presents its parent class name instead — the binding SHALL answer negatively rather than panicking or logging.

#### Scenario: Query under a parent-class presentation reports no revert support
- **WHEN** a revert query is routed to an instance whose reported class fails the registered-class lookup (e.g. `"Control"` during destruction)
- **THEN** the query completes and reports no revert support
