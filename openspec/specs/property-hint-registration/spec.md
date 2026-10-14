## Purpose

Defines the metadata contract for properties registered by Go-defined
GDExtension classes: hint, hint string, and usage flags delivered to the
engine's class database, class-name derivation from resource-type hints, and
pointer ownership during registration.

## Requirements

### Requirement: Property registration carries hint metadata
The system SHALL deliver a registered property's hint, hint string, and
usage flags to the engine class database exactly as declared by the Go
registration call, so that reflection reports them unchanged.

#### Scenario: Range hint reports unchanged
- **WHEN** a float property is registered with hint `PROPERTY_HINT_RANGE`
  and hint string `0:100:1`
- **THEN** the class property list reports hint `RANGE` with hint string
  `0:100:1`

#### Scenario: Enum hint enables dropdown
- **WHEN** an int property is registered with hint `PROPERTY_HINT_ENUM` and
  a comma-separated hint string of names
- **THEN** the class property list reports hint `ENUM` with that exact hint
  string

#### Scenario: Non-default usage flags report unchanged
- **WHEN** a property is registered with usage limited to storage
  (`PROPERTY_USAGE_STORAGE` only)
- **THEN** the class property list reports storage-only usage and the
  property is not exposed as editor-visible

### Requirement: Resource-type hint binds class name from hint string
When a property is registered with `PROPERTY_HINT_RESOURCE_TYPE`, the
registered property's `class_name` SHALL be the class named by the hint
string. When the hint is not a resource-type hint, `class_name` SHALL be an
explicitly declared class qualifier if present, and otherwise the owning
class name.

#### Scenario: Resource property narrows to hinted type
- **WHEN** a `RESOURCE`-typed property is registered with hint
  `PROPERTY_HINT_RESOURCE_TYPE` and hint string `Texture2D`
- **THEN** the class property list reports `class_name` `Texture2D`, not the
  owning class name

#### Scenario: Qualifier applies when no resource-type hint
- **WHEN** an `OBJECT`-typed property is registered with hint
  `PROPERTY_HINT_NONE` and an explicit class qualifier
- **THEN** the class property list reports the qualifier as `class_name`

#### Scenario: No hint and no qualifier keeps owning class
- **WHEN** a property is registered without any hint or class qualifier
- **THEN** the class property list reports the owning class name as
  `class_name`, matching prior behavior

### Requirement: Hint-free registration preserves existing defaults
The existing hint-free registration convenience SHALL keep registering
properties with hint `PROPERTY_HINT_NONE`, an empty hint string, and
`PROPERTY_USAGE_DEFAULT`, leaving current callers' observable behavior
unchanged.

#### Scenario: Legacy-style registration reports defaults
- **WHEN** a property is registered through the hint-free convenience
- **THEN** the class property list reports hint `NONE`, empty hint string,
  and default usage

### Requirement: Registration owns string temporaries
The system SHALL own the conversion and lifetime of property name, hint
string, and class qualifier values for the duration of the engine
registration call, releasing them once the engine has copied them. Callers
MUST NOT be required to pre-allocate or manage Godot string objects, and
engine-stored property metadata MUST remain valid after the call returns.

#### Scenario: Metadata stays readable after registration
- **WHEN** a property is registered with a Go-native hint string and the
  registration call has returned
- **THEN** reading the class property list still yields the declared hint
  string without crashes or corrupted text
