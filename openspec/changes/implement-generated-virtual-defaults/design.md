## Context

The virtual surface: 1437 virtuals across 106 of 1036 classes in `godot_headers/extension_api.json`. Post-`implement-hierarchical-virtual-methods`, resolution walks the embedding chain at registration and picks the most-derived qualified implementation; the callback layer (`get_virtual_call_data2`/`call_virtual_with_data`) is name-agnostic. Generated `*Impl` layers currently declare zero virtual bodies — every default is user-authored.

**Engine constraint (verified against Godot 4.7 source, godot-cpp, and gdext):** overriding an engine GDVIRTUAL replaces its behavior wholesale. Godot resolves virtual presence once per object instance into a cached member pointer (`gdvirtual.gen.h`; absence caches as an INVALID sentinel) and never re-consults — so once a class registers a virtual, the engine's own default body is unreachable for those instances, and a wrapper call from inside the override re-enters the registered implementation forever (empirically pinned: depth-capped panic in `TestDelegationRepro`). Neither reference binding offers delegation:

| Binding | Override mechanism | Engine-default reach |
|---|---|---|
| godot-cpp | C++ inheritance; compile-time override detection via `if constexpr (!std::is_same_v<decltype(&B::_x), decltype(&T::_x)>)` | none |
| godot-rust | `I<Class>` capability traits; unimplemented trait methods are never registered | none ("No access to `super` methods", book) |

An earlier iteration of this change proposed generated delegating bodies plus a re-entrancy guard; the task-1.3 prototype falsified both variants (callback-layer guards yield zero values because presence is frozen; suppression windows cannot reach the per-instance cache). The pivot: declaration-only codegen with honest limits.

## Goals / Non-Goals

**Goals:**
- Typed, discoverable virtual surface: every Category-A virtual has a generated, greppable Go counterpart on its Impl layer.
- Compile-time override verification through interface conformance assertions.
- Delegation impossibility documented and behaviorally pinned so it is never rediscovered through debugging.

**Non-Goals:**
- Any runtime behavior attached to generated declarations (no bodies, no registration changes, no dispatch cost).
- Auto-invocation chaining or engine-default recovery (impossible; see Context).
- Category-B/C coverage beyond census classification.

## Decisions

### Decision 1: The generator derives the catalog from extension_api.json; the census is its fixture
The generator computes the Category-A classification itself — criterion: `is_virtual` AND non-void return AND a plain wrapper named without the leading underscore exists on the class hierarchy — straight from `godot_headers/extension_api.json`, keeping one source of truth and no regeneration tooling outside the build. The checked-in `cmd/generate/gdclassimpl/virtual_census.json` snapshot is demoted to a test fixture: codegen tests assert the emitted surface matches it exactly, so criterion drift fails loudly. Counts at census time: A=524, B=913, total 1437; Category C names never appear in the API surface.

### Decision 2: Declaration-only interfaces for cataloging; anonymous-interface assertions for verification
For each class with Category-A virtuals, the generator emits `type <Class>Virtuals interface { … }` declaring every virtual under qualified names with exact signatures. These are the catalog — greppable, godoc-visible, zero runtime footprint. They deliberately do NOT serve single-method verification: Go interface satisfaction is all-or-nothing, so asserting a fifty-method interface would force implementing every virtual. The verification idiom is a per-method anonymous-interface assertion, requiring no generated types:

```go
var _ interface{ V_Control_GetMaximumSize() Vector2 } = (*MyThing)(nil)
```

This compiles only when `MyThing` declares that exact method signature (embedding promotion included). Panic bodies were rejected: promoted methods would pollute resolution candidates and add binary weight for no behavioral gain.

### Decision 3: Delegation impossibility is a documented, pinned fact
docs/overview.md states the constraint with the evidence trail (engine caching semantics; godot-cpp/gdext parity). The suite keeps `TestDelegationRepro` bound but unexercised except by a dedicated test asserting the bounded-depth abort — converting accidental infinite recursion into a fast diagnostic.

## Risks / Trade-offs

- **Interfaces can drift from engine reality** → mitigated by generating from `extension_api.json` (same source as all other bindings) and by the census-completeness test.
- **Developers may expect delegation from the declarations** → mitigated by doc comments on generated interfaces stating replacement-wholesale semantics explicitly.
- **524-entry interface surface adds visual bulk to Impl layers** → interfaces live in their own generated file per package; nothing existing changes shape.

**Constraint on future scope extension:** if generated defaults are ever revisited (e.g., snapshot-based constants), they MUST NOT auto-register into `VirtualMethodMap`, and must respect the absence semantics pinned by `property-revert-virtuals` / `notification-virtual-dispatch`.

## Migration Plan

Additive. Generated files land under `pkg/gdclassimpl/*.gen.go`; existing user classes unaffected.

Rollback: regenerate with templates disabled.

## Open Questions

- Whether any Category-B virtual needs first-class support for ecosystem reasons (e.g., `_physics_process` bookkeeping) — deferred unless demoed need.
