## Purpose

Makes the full engine-GDVIRTUAL surface discoverable and safely extensible in Go: generated `*Impl` layers ship default virtual implementations under the qualified naming convention, so user classes can override any such virtual and still reach engine-default behavior through explicit delegation — without infinite dispatch recursion and without displacing engine fallbacks for classes that opt out.

## ADDED Requirements

### Requirement: Generated Defaults Exist For Engine-GDVIRTUAL Virtuals
For every class in `extension_api.json` declaring an engine-side GDVIRTUAL method, code generation SHALL emit a corresponding `V_<ClassName>_<MethodName>` default implementation on that class's generated Impl layer, following the qualified convention from `hierarchical-virtual-methods`.

#### Scenario: Generated body exists and is typed
- **WHEN** a user embeds the Impl layer of a class with a declared GDVIRTUAL (e.g. `ControlImpl`)
- **THEN** a compilable, typed default implementation (`V_Control_GetMaximumSize`) is present and greppable without consulting Godot docs

### Requirement: Delegation Reaches Engine Behavior Without Recursion
A generated default implementation SHALL reach the engine's own behavior by calling the plain method wrapper exactly once, such that a most-derived user override delegating to the promoted generated selector terminates with the engine body rather than re-entering the registered virtual chain.

#### Scenario: Override extends engine default
- **WHEN** a leaf class overrides `V_<Leaf>_GetMaximumSize`, delegates via `t.ControlImpl.V_Control_GetMaximumSize()`, and Godot invokes `_get_maximum_size` on an instance
- **THEN** execution terminates in the engine C++ body and returns, with no repeated invocation of the leaf override

### Requirement: Defaults Do Not Displace Engine Fallbacks Unless Opted In
A class whose resolution walk finds no user implementation SHALL keep unregistered-virtual semantics by default; a generated default participates in registration only through the codegen-emitted opt-in hook.

#### Scenario: Non-opted class keeps engine fallback
- **WHEN** a class registers without opting into generated defaults and Godot queries its GDVIRTUAL call data
- **THEN** the binding reports absence and the engine executes its own default body

#### Scenario: Opted-in class serves the generated fallback
- **WHEN** a class opts in and Godot invokes the virtual with no deeper override present
- **THEN** the generated shallowest-level body runs and returns the engine-equivalent result
