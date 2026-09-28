# Spec Delta

## Purpose

Defines how the binding answers Godot's `to_string` creation-info callback (`GDExtensionClassToString`): report validity through the out parameter, dispatch to a user-registered `to_string` virtual when one exists, and encode the delivered string so Godot renders it correctly.

## ADDED Requirements

### Requirement: Validity Is Reported Through The Out Parameter
When the binding supplies a string to Godot's `to_string` callback, it SHALL write the validity flag through the `r_is_valid` out pointer so the engine accepts the supplied string instead of falling back to the default `Object::to_string()`.

#### Scenario: Engine accepts the supplied string
- **WHEN** Godot invokes the `to_string` callback and the binding produces a string
- **THEN** `*r_is_valid` is set to true and the string written to the out pointer is what `to_string()` returns in GDScript

#### Scenario: GDScript sees the binding's string, not the engine default
- **WHEN** GDScript calls `to_string()` on an instance of a Go-registered class
- **THEN** the returned string is not the engine's default `ClassName:<ClassName#id>` form

### Requirement: User-Registered to_string Virtual Is Dispatched
When a Go-registered class has bound a `to_string` virtual (registered in the class's virtual method map), the binding SHALL invoke that Go method and deliver its returned string to Godot.

#### Scenario: Override's return value reaches Godot
- **WHEN** GDScript calls `to_string()` on an instance whose class registered a `to_string` virtual
- **THEN** the Go virtual executes and the string it returned is exactly what GDScript receives

### Requirement: Classes Without An Override Produce A Sensible Default String
When a Go-registered class provides no `to_string` virtual, the binding SHALL still supply a non-empty string identifying the class and instance, reported as valid.

#### Scenario: Default format for classes with no override
- **WHEN** Godot requests `to_string` for an instance of a class that never registered a `to_string` virtual
- **THEN** the binding supplies the non-empty default `[ GDExtension::<class> <--> Instance ID:<id> ]` with validity set to true

### Requirement: Supplied String Is UTF-8 Encoded
The binding SHALL construct the out Godot `String` from the Go string's UTF-8 bytes via `string_new_with_utf8_chars`, so any non-ASCII content renders correctly in Godot.

#### Scenario: Non-ASCII override renders correctly
- **WHEN** a user `to_string` virtual returns a Go string containing non-ASCII characters (e.g. `héllo — 世界`)
- **THEN** GDScript receives the same characters, not Latin-1-mangled byte substitutions
