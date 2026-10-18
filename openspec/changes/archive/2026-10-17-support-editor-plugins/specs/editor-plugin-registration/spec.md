# Spec Delta

## Purpose

Governs how a Go extension registers and unregisters editor plugin classes
with the running Godot editor, including misuse rejection outside editor
builds and automatic unload at editor shutdown.

## ADDED Requirements

### Requirement: Register an editor plugin class

The library SHALL provide a generic operation that adds an already
ClassDB-registered class descending from `EditorPlugin` as an active editor
plugin, identified by the class name carried by an instance argument. In an
editor build the engine SHALL instantiate the class and remove it again when
the matching removal operation is used.

#### Scenario: Plugin instantiated at editor startup

- **WHEN** the add operation is called during editor-level initialization with
  a registered `EditorPlugin`-derived class
- **THEN** the engine instantiates that class during editor startup (its Go
  constructor runs) and the plugin becomes an active editor plugin

#### Scenario: Explicit removal

- **WHEN** the remove operation is called for a class that was previously
  added
- **THEN** the engine removes that editor plugin

### Requirement: Registration guards

The add operation SHALL reject adding a class that is already registered as a
plugin, and the remove operation SHALL reject removing a class that was never
added. A rejected operation MUST report the misuse through the library's
recoverable panic convention and MUST NOT change the registration state.

#### Scenario: Duplicate add rejected

- **WHEN** the add operation is called twice with the same class name
- **THEN** the second call panics recoverably and the plugin is not added a
  second time

#### Scenario: Unknown removal rejected

- **WHEN** the remove operation is called with a class name that is not
  currently registered as a plugin
- **THEN** the call panics recoverably and registration state is unchanged

### Requirement: Rejection outside editor builds

Editor plugin operations SHALL be rejected with a recoverable panic when the
extension runs outside an editor build, where the engine-side registration
would silently do nothing. The check SHALL NOT depend on engine calls
succeeding silently.

#### Scenario: Runtime build refuses registration

- **WHEN** the add operation is called from a runtime (non-editor) build
- **THEN** the call panics recoverably, no engine-side registration occurs,
  and the runtime test gate remains green with the plugin code linked in

### Requirement: Automatic unload at editor shutdown

Every plugin class that was added and not removed SHALL be removed
automatically when editor-level deinitialization runs, after any
user-registered editor terminator callback has run. The registration state
SHALL be empty afterwards.

#### Scenario: Auto-unload leaves no leaks

- **WHEN** the editor shuts down with a plugin registered and never manually
  removed
- **THEN** the plugin is removed during editor-level deinitialization and the
  shutdown reports zero ObjectDB leaks

#### Scenario: Terminator sees live registration

- **WHEN** a user-registered editor terminator runs during shutdown
- **THEN** plugins added by that code are still registered at the time the
  terminator runs
