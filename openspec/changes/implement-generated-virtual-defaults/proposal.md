## Why

`implement-hierarchical-virtual-methods` made multi-level virtual coexistence possible; today it remains manual. Of the 1437 virtual methods declared across 106 classes in `godot_headers/extension_api.json`, every one must be hand-implemented and hand-bound (`ClassDBBindMethodVirtual`) by each user who needs it, with no discoverable Go-side trace of the full virtual surface. Three gaps follow:

- **No typed discovery**: users learn overridable virtuals from Godot docs, not from the generated `*Impl` layers they embed.
- **No safe engine-default delegation**: a leaf overriding an engine-GDVIRTUAL (e.g. `Control._get_maximum_size`) has no correct way to extend the engine default — calling the plain wrapper (`t.GetMaximumSize()`) re-enters the registered virtual chain and recurses infinitely; skipping the wrapper means reimplementing engine logic that drifts upstream.
- **Partial hierarchies are brittle**: intermediate wrapper levels (the `TestHierarchicalBase` pattern) exist only when users author them.

## What Changes

- Code generation emits default virtual implementations into `pkg/gdclassimpl` `*Impl` layers using the qualified `V_<ClassName>_<MethodName>` convention, covering the virtual surface declared in `extension_api.json`.
- Emitted defaults participate in embedding-chain resolution as shallowest-level fallbacks: user overrides win by depth, and delegation to a generated default via its promoted selector reaches engine behavior without re-entering the virtual dispatch cycle.
- Registration policy distinguishes virtual categories (see design.md): engine-GDVIRTUAL defaults must never *displace* engine fallbacks by falsely reporting presence where absence was meaningful.
- Update docs/overview.md to describe the generated virtual surface and delegation patterns.

## Capabilities

### New Capabilities

- `generated-virtual-defaults`: Codegen-emitted default virtual bodies across Impl layers, resolvable as chain-shallowest fallbacks and explicitly callable for engine-faithful delegation.

### Modified Capabilities

None. Dispatch mechanics remain pinned by `hierarchical-virtual-methods`; unimplemented-virtual fallback stays pinned by `gdvirtual-unimplemented-defaults`.

## Non-goals

- Auto-invocation chaining (base runs after derived automatically); delegation stays explicit.
- Dedicated creation-info callback implementations (revert/notification) — separate change.
- Callback-layer (`classdb_callback.go`) changes.
- Multiple unrelated embedded branches beyond Go promotion rules.
