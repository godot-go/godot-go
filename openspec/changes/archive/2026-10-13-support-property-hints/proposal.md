# Proposal

## Why

Go-registered classes declare properties through `ClassDBAddProperty`, which
hardcodes `PROPERTY_HINT_NONE`, an empty hint string, and
`PROPERTY_USAGE_DEFAULT` (`pkg/core/classdb.go:133-144`). Without hints, the
Godot inspector and GDScript reflection treat every Go-declared property as a
bare scalar: no resource picker, no enum dropdown, no range slider, and
object-typed properties never advertise their class. The class-name
derivation that resource-type hints require is already ported but unplaced
(`TODO.md`). This closes roadmap item B1 and folds in D2.

## What Changes

- Add a Go-native property spec (variant type, name, setter/getter, hint,
  hint string, usage flags, optional class qualifier) with a new
  registration entry point `ClassDBAddPropertyInfo` in `pkg/core`.
- Re-implement the existing `ClassDBAddProperty` on top of it as the
  hint-free convenience (hint NONE, empty hint string, usage DEFAULT);
  existing callers keep compiling and behaving identically.
- Port the godot-cpp derivation: when the hint is a resource-type hint, the
  registered property's `class_name` comes from the hint string; otherwise
  from the owning class or the explicit qualifier.
- Honor explicit usage flags instead of always `PROPERTY_USAGE_DEFAULT`.
- Resolve the ownership TODO in `pkg/ffi/property_info.go:25`: the binding
  owns the `StringName`/`String` temporaries for the registration call
  (pin before the engine call, destroy after — the engine copies the data,
  as the signal-registration path already relies on), removing the
  caller-must-preallocate-pointer burden.
- Register test-fixture properties exercising resource-type, enum, and
  range hints plus non-default usage flags; assert through the GDScript
  test harness via the `ClassDB` property list.

## Capabilities

### New Capabilities

- `property-hint-registration`: the metadata contract for registering Go
  class properties — hint, hint string, usage flag delivery to ClassDB,
  `class_name` derivation from resource-type hints, and pointer lifecycle
  ownership during registration.

### Modified Capabilities

None. `property-revert-virtuals` governs revert queries and is unaffected;
`ClassDBAddProperty`'s observable default behavior is unchanged.

## Impact

- `pkg/core/classdb.go` — `ClassDBAddProperty` refactor + new
  `ClassDBAddPropertyInfo` entry point
- `pkg/core/object.go` — `NewGDExtensionPropertyInfoFromNames` reused or
  superseded by the spec constructor
- `pkg/ffi/property_info.go` — TODO removed; ownership documented
- `test/pkg/example.go`, `test/pkg/exampleref.go` — new hinted fixtures
- `test/demo/main.gd` — property-list assertions
- No generated-code or FFI-surface changes

## Non-goals

- Hint support for `ClassDBAddSignal` parameters (signal params stay
  `PROPERTY_HINT_NONE`)
- Editor plugin or inspector UI work beyond what hints unlock automatically
- Runtime `_validate_property`/property-revert behavior changes
- Method argument/return `PropertyInfo` metadata (`method-argument-type-metadata`
  stays as-is)
