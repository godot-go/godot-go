# class-property-list Delta

## ADDED Requirements

### Requirement: List-Declared Properties Survive Repeated Engine Queries

Every `get_property_list()` query against a live instance of a registered class SHALL report the class's list-declared custom properties with identical name, type, class_name, hint, hint_string, and usage, regardless of how many times the query has run before.

#### Scenario: Second read matches the first

- **WHEN** GDScript calls `get_property_list()` on an instance twice in a row
- **THEN** both reads contain the class's list-declared custom property
- **AND** the two entries report identical metadata

#### Scenario: Metadata stays valid after unrelated engine work

- **WHEN** a property-list query is followed by garbage collection, signal dispatch, and other engine activity, and then the query runs again
- **THEN** the custom property entry still reports valid, uncorrupted name and hint data

### Requirement: Exposed Property Memory Is Released By The Paired Free Callback

The memory backing a returned property array SHALL be released by the paired property-list free callback, and the extension SHALL NOT retain the exposed array beyond that callback. Repeated query/free cycles SHALL NOT grow extension memory retention.

#### Scenario: Query churn stays flat

- **WHEN** a test drives thousands of `get_property_list()` queries against an instance and forces Go garbage collection between samples
- **THEN** live Go heap after collection stays within a constant band regardless of iteration count

#### Scenario: Queries after free still succeed

- **WHEN** the engine copies a returned array and invokes the paired free callback, then issues a new query
- **THEN** the new query returns the full property set again rather than an empty list or stale data

### Requirement: Property String Data Lives With The Class Registration

The string data behind exposed property names, class names, and hint strings SHALL remain valid for the whole class registration lifetime, across any sequence of queries and free callbacks, and SHALL be destroyed exactly once when the class registration is torn down — independent of whether any free callback already ran.

#### Scenario: No double destroy regardless of query timing

- **WHEN** a class registration is destroyed after zero queries, after one query, and after many queries
- **THEN** each teardown frees its property strings exactly once with no crash or double-free warning

#### Scenario: Late query still reports the hint string

- **WHEN** a query runs long after several free callbacks for earlier queries completed
- **THEN** hint strings and class names in the result read back exactly as registered
