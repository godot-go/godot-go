## Why

The current virtual method convention (docs/overview.md#virtual-methods) uses flat Go names: `V_Ready` registered as `_ready`. A name identifies a virtual only within the single Go type that declares it, but Go struct embedding — the project's substitute for class inheritance — promotes methods across the whole embedding chain. That makes the flat scheme break down for real hierarchies:

- **One slot per virtual.** If a user type defines `V_PhysicsProcess` while an embedded `Impl` type also promotes one, Go shadowing hides the base version entirely — the base implementation becomes unreachable, so "extend the default behavior" is impossible.
- **Diamond ambiguity panics.** With two embedded branches both promoting the same `V_` name, the selector is ambiguous: `reflect.MethodByName("V_...")` fails and `classDBBindMethod` panics with "unable to find function".
- **Generated defaults are unsafe.** Because of the above, generated `*Impl` layers can never ship default virtual bodies without risking collisions in every downstream class.

godot-cpp does not have this problem because C++ classes each declare their own override in a real inheritance chain. The proposal explores a qualified convention — `V_<ClassName>_<MethodName>` (e.g. `V_PlayerCharacter_PhysicsProcess`) — so every level of a hierarchy can carry its own implementation simultaneously, dispatch resolves to the most-derived one, and deeper levels stay callable explicitly (super-style delegation).

## What Changes

- Introduce the qualified virtual naming pattern `V_<ClassName>_<PascalCaseMethodName>` as the only supported form; GDScript's leading underscore is dropped (`_physics_process` → `PhysicsProcess`).
- Extend virtual registration (`ClassDBBindMethodVirtual`) to resolve implementations by walking the receiver type's embedding chain and binding the most-derived implementation of each requested GDScript virtual.
- **Clean break:** flat `V_<MethodName>` registration panics at registration time with a migration-oriented diagnostic naming the expected qualified form. All in-repo flat bindings migrate in the same change; there is no compatibility window.
- Update docs/overview.md's Virtual Methods section to specify the required convention.
- Add demo/test coverage: a multi-level embedding chain where base and derived levels both implement the same virtual, asserting most-derived dispatch plus explicit delegation to the base-level implementation.

## Capabilities

### New Capabilities

- `hierarchical-virtual-methods`: Qualified `V_ClassName_MethodName` naming and most-derived-wins resolution of Godot virtuals across Go struct-embedding chains, including explicit delegation to shallower implementations and engine-default fallback when no level implements a virtual.

### Modified Capabilities

None. `gdvirtual-unimplemented-defaults` already pins the nil-call-data fallback for unimplemented virtuals; this change adds resolution semantics above it rather than altering them.

## Impact

- `pkg/core/classdb.go`: virtual registration path gains embedding-chain resolution (reflection walk over embedded Impl fields); no changes to the callback layer (`get_virtual_call_data2`/`call_virtual_with_data` still key off `VirtualMethodMap`).
- Generated code (`pkg/gdclassimpl`): unaffected initially; the convention unblocks generated default implementations later but generating them is out of scope here.
- `test/pkg/example.go`, `test/demo/main.gd`: new multi-level test classes and assertions.
- User-facing API: breaking change — qualified names are required; flat `V_X` bindings fail loudly at startup instead of being silently deprecated.
