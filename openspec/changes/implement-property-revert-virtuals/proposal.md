## Why

Issue #81: GDExtension classes registered from Go cannot implement the inspector revert contract. The class creation info wires Godot's `property_can_revert_func`/`property_get_revert_func` to Go trampolines (`GoCallback_ClassCreationInfoPropertyCanRevert`/`...GetRevert`, `pkg/core/classdb_callback.go:256-264`), but both are stubs that unconditionally return 0 — they never dispatch to a registered class's virtual methods. As a result:

- `V_PropertyCanRevert` already exists in the demo `Example` and is bound as `_property_can_revert` (test/pkg/example.go:290, 844), but the binding is dead code: Godot can call into the stub, and the stub always answers "cannot revert".
- `V_PropertyGetRevert` does not exist anywhere in the project.
- `Object.property_can_revert()` / `Object.property_get_revert()` called on any Go-registered instance therefore report that no property can be reverted, so the editor inspector's revert-to-default affordance never appears for Go classes.

## What Changes

- Implement `GoCallback_ClassCreationInfoPropertyCanRevert` to dispatch to the class's registered `_property_can_revert` virtual via `VirtualMethodMap`, passing the borrowed `StringName` argument, and return whether the Go method claims the property is revertable.
- Implement `GoCallback_ClassCreationInfoPropertyGetRevert` to dispatch to `_property_get_revert`, writing the returned Variant into `r_ret` only when the Go method returns true (mirroring godot-cpp's `(name) -> (value, handled)` semantics).
- Add `V_PropertyGetRevert` to the demo `Example` class and bind it as `_property_get_revert`; keep the existing `V_PropertyCanRevert` as-is.
- Add GDScript assertions exercising `example.property_can_revert(...)` / `example.property_get_revert(...)` end-to-end.

## Capabilities

### New Capabilities

- `property-revert-virtuals`: Dispatch of Godot's class creation info revert callbacks to Go-implemented `_property_can_revert`/`_property_get_revert` virtuals, including graceful behavior for classes that do not implement them.

### Modified Capabilities

None. No existing capability specifies property revert behavior; `gdvirtual-unimplemented-defaults` covers engine-side GDVIRTUAL defaults, not these creation-info callbacks.

## Impact

- `pkg/core/classdb_callback.go`: two exported callbacks gain real implementations; no signature changes (C trampolines in `classdb_callback.c/.h` and wiring in `classdb.go` already exist).
- `test/pkg/example.go`: one new virtual method + one bind line.
- `test/demo/main.gd`: new assertion block.
- No codegen changes (`cmd/generate` untouched); no API surface changes outside the newly functional virtuals.
