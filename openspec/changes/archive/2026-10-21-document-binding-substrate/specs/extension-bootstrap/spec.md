# Spec Delta

## Purpose

Defines how a Go GDExtension starts up and shuts down: which process-lifetime
state is prepared before the engine calls into user code, how an initialisation
level is chosen and dispatched, how the engine-class tables are seeded, and in
what order deinitialisation runs.

## ADDED Requirements

### Requirement: Binding start-up prepares process-lifetime state before the engine calls user code

On the extension entry point the binding SHALL construct the instance,
class-registration, and constructor registries, load the GDExtension interface,
install the level callbacks on the engine's initialization struct, bind the
Variant machinery, and seed the engine-class tables before reporting success.

#### Scenario: First load succeeds with usable Variant bindings

- **WHEN** the engine loads the extension for the first time and invokes its entry point
- **THEN** the entry point returns true
- **AND** the interface, the three registries, the Variant bindings, and the engine-class tables are all usable before any user initializer runs

#### Scenario: An unavailable interface function does not abort start-up

- **WHEN** a function named in the loaded header is unavailable on the running engine
- **THEN** start-up still completes and the missing slot is left unavailable for the guards that already handle it

### Requirement: The engine's minimum initialisation level is the lowest level with a registered initializer

The binding SHALL report to the engine the lowest initialisation level for which
a user initializer is registered, so that levels below it are never delivered.

#### Scenario: Only a scene initializer is registered

- **WHEN** a user registers an initializer for the scene level and no lower level
- **THEN** the initialization struct's minimum level is the scene level
- **AND** the engine delivers no core or servers level to this extension

#### Scenario: Several levels include core

- **WHEN** the user registers initializers for the editor level and the core level
- **THEN** the reported minimum level is core

### Requirement: Start-up fails when no initializer is registered

The binding SHALL refuse to complete start-up when no initialisation level has a
registered initializer, because an extension with no initializer would be
unloaded and reloaded without ever registering a class.

#### Scenario: No initializer registered

- **WHEN** the entry point runs with every level's initializer slot empty
- **THEN** start-up fails loudly instead of returning true

### Requirement: Each delivered level runs its initializer or terminator exactly once

The binding SHALL invoke the initializer registered for a delivered level, and on
deinitialisation SHALL invoke that level's terminator before any level-scoped
cleanup, treating a missing callback as a no-op.

#### Scenario: Level initializer runs

- **WHEN** the engine delivers the scene initialisation level and a scene initializer is registered
- **THEN** that initializer runs exactly once for that delivery

#### Scenario: Level without an initializer is a no-op

- **WHEN** the engine delivers a level whose initializer slot is empty
- **THEN** no user callback runs and no error is reported

#### Scenario: Terminator runs before level-scoped cleanup

- **WHEN** the editor level is deinitialised with both a terminator and registered editor plugins
- **THEN** the terminator runs before the editor-plugin registry is drained

### Requirement: Deinitialisation preserves process-lifetime registries

Deinitialisation SHALL NOT clear or rebuild the instance, registration, class
constructor, engine binding-callback, or Ref-constructor registries; they remain
valid for the life of the process.

#### Scenario: A level deinitialises while a later callback still resolves a class

- **WHEN** a class was registered during an earlier level and a later callback asks for its registration
- **THEN** the registration is still present after the earlier level's deinitialisation

#### Scenario: A shutdown callback resolves an engine class

- **WHEN** the main-loop shutdown hook asks for an engine class's binding callbacks or Ref constructor
- **THEN** the lookup succeeds
