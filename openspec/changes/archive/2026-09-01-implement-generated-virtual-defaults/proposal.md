## Why

`implement-hierarchical-virtual-methods` made multi-level virtual coexistence possible, but the 1437 virtual methods declared across 106 classes in `godot_headers/extension_api.json` remain invisible to Go developers: there is no typed, greppable counterpart for any of them, no compile-time signature verification when overriding, and no record of which virtuals exist. Three gaps follow:

- **No typed discovery**: users learn overridable virtuals from Godot docs, not from the generated `*Impl` layers they embed.
- **No override verification**: a typo'd or mis-typed qualified method fails only at registration time (or silently never registers), with no compile-time check against Godot's declared signatures.
- **Delegation trap is undocumented**: overriding an engine GDVIRTUAL replaces its behavior wholesale — engine-default bodies are unreachable from GDExtension once a virtual is registered (Godot caches presence per instance at first `GDVIRTUAL_CALL`). Both reference bindings confirm this: godot-cpp and godot-rust offer no super-call mechanism either. Without documentation and pinned tests, every contributor risks rediscovering this through infinite recursion.

## What Changes

- Code generation emits **per-class virtual-surface interfaces** (`ControlVirtuals`, …) declaring every Category-A virtual under the qualified `V_<ClassName>_<MethodName>` convention with exact Godot signatures — declaration-only, zero runtime behavior (the Go analog of godot-cpp's header declarations).
- A checked-in census (`cmd/generate/gdclassimpl/virtual_census.json`) records the full classification of all 1437 virtuals with documented criteria; codegen tests assert emitted-interface completeness against it.
- docs/overview.md documents that GDExtension virtual overrides replace engine behavior wholesale (delegation to engine defaults is impossible), and the suite pins the recursion trap with a guarded regression test.
- Compile-time adoption idiom documented: asserting a user struct against a generated interface verifies override signatures at build time.

## Capabilities

### New Capabilities

- `virtual-surface-catalog`: Typed, discoverable catalog of Godot's virtual surface — generated per-class interfaces with exact signatures, a checked-in census, and honest documentation of delegation limits.

### Modified Capabilities

None. Dispatch mechanics remain pinned by `hierarchical-virtual-methods`; unimplemented-virtual fallback stays pinned by `gdvirtual-unimplemented-defaults`.

## Non-goals

- Engine-default delegation from overrides — **impossible**: Godot resolves virtual presence once per instance and never re-consults; both reference bindings lack it (verified against godot-cpp source, gdext docs/issues, and Godot's `gdvirtual.gen.h`).
- Generated default bodies, panic bodies, or any runtime behavior attached to generated declarations.
- Auto-registration or opt-in hooks that change which virtuals dispatch.
- Category-B/C coverage beyond the census classification.
