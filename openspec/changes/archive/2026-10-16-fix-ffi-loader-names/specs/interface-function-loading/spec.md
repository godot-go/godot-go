# Spec Delta

## Purpose

Ensures the extension resolves every GDExtension interface function under the
registration name the engine publishes for it in the interface header, so no
generated wrapper is left as a nil function by a symbol-name mismatch and future
header refreshes cannot silently reintroduce such mismatches.

## ADDED Requirements

### Requirement: Interface Functions Load Under Their Header-Published Registration Names

The FFI loader SHALL request each GDExtension interface function using the
registration name recorded in the `@name` metadata of that function's typedef in
the bundled interface header (`godot_headers/godot/gdextension_interface.h`),
regardless of whether that name matches the snake-cased Go typedef identifier.

#### Scenario: Custom-callable user-data symbol loads

- **WHEN** the extension loads and runs the interface loader
- **THEN** the `callable_custom_get_userdata` slot resolves to a non-nil function
  and no "Attempt to get non-existent interface function" error is emitted for
  `callable_custom_get_user_data`

#### Scenario: Versioned class-registration symbols load

- **WHEN** the extension loads against a Godot version that registers
  `classdb_register_extension_class5` and `classdb_register_extension_class6`
- **THEN** both slots resolve non-nil and no "Attempt to get non-existent interface
  function" error is emitted for `classdb_register_extension_class_5` or
  `classdb_register_extension_class_6`

#### Scenario: Variant internal-getter symbol loads

- **WHEN** the extension loads against a Godot version that registers
  `variant_get_ptr_internal_getter`
- **THEN** the corresponding interface slot resolves non-nil and no "Attempt to get
  non-existent interface function" error is emitted for
  `get_variant_get_internal_ptr_func`

#### Scenario: Full load against the pinned engine

- **WHEN** the test suite loads the extension against the pinned Godot version
- **THEN** every interface-function slot the loader attempts resolves non-nil and
  the run log contains no "Attempt to get non-existent interface function" errors

### Requirement: Generated Load Names Derive From Header Metadata With An Enforced Fallback

The FFI code generator SHALL derive each generated load name from the interface
header's `@name` metadata whenever that metadata is present, SHALL apply the
legacy identifier-based name transform only for typedefs without `@name`, and a
generator test SHALL fail when any collected interface-function typedef in the
bundled header lacks `@name` coverage.

#### Scenario: Metadata wins over identifier reconstruction

- **WHEN** the generator emits a load call for a typedef whose `@name` differs from
  the snake-cased typedef identifier, such as a typedef that keeps a legacy
  identifier while `@name` records the current engine symbol
- **THEN** the emitted load string is the `@name` value verbatim

#### Scenario: Fallback for typedefs without metadata

- **WHEN** the generator emits a load call for an interface typedef that carries no
  `@name` metadata
- **THEN** the emitted load string is produced by the legacy identifier transform

#### Scenario: Coverage guard trips on header drift

- **WHEN** a refreshed interface header introduces an interface-function typedef
  that has no `@name` metadata
- **THEN** the generator test fails and names the uncovered typedef
