## Context

Virtuals today: user declares `V_Ready` on their type and registers it with `ClassDBBindMethodVirtual(t, "V_Ready", "_ready", nil, nil)`. Registration (`classDBBindMethod`, pkg/core/classdb.go:281) resolves the Go method with `reflect.TypeFor[T]().MethodByName(goMethodName)` — a flat name lookup over T's full promoted method set — and stores it in `ci.VirtualMethodMap["_ready"]`. Dispatch is already decoupled from names: `get_virtual_call_data2` reports presence, `call_virtual_with_data` looks up by GDScript name and ptrcalls the stored bind (pkg/core/classdb_callback.go:66-150). So all resolution change lives at registration time; the callback layer needs nothing.

The embedding chain: generated Impl types form a single-parent chain mirroring Godot inheritance (`CharacterBody2DImpl` embeds `PhysicsBody2DImpl`, which embeds ... `ObjectImpl`; 1036 such structs in classes.gen.go). User classes embed exactly one leaf Impl. Go promotes methods up the chain; ambiguous diamond promotions are excluded from the method set entirely (MethodByName misses → panic), and shadowing makes the shallower method unreachable.

docs/overview.md#virtual-methods documents the flat convention and the get_virtual_call_data2 path; it must be updated as part of this change.

## Goals / Non-Goals

**Goals:**
- Every embedding level can implement any virtual simultaneously under `V_<ClassName>_<MethodName>`.
- Most-derived-wins dispatch for instances; explicit super-style delegation from derived to base implementations.
- Zero callback-layer changes; engine-default fallback preserved.
- Clean break: flat names fail loudly at registration, so no dual-convention ambiguity ever exists.

**Non-Goals:**
- Generating default virtual bodies in `pkg/gdclassimpl` (unblocked later; separate change).
- Multiple-implementation *chaining* (auto-invoking base after derived); delegation stays explicit.
- Supporting multiple *unrelated* embedded branches (true multiple inheritance) beyond what Go promotion allows.

## Decisions

### Decision 1: Name mapping drops GDScript's leading underscore and PascalCases the rest
`_physics_process` at level `PlayerCharacter` becomes `V_PlayerCharacter_PhysicsProcess`. This matches the existing flat style (`V_Ready` ↔ `_ready`) and avoids ugly double underscores. The registration API keeps taking the explicit GDScript name (`"_physics_process"`), so the Go name is convention only — never parsed back.

### Decision 2: Resolution walks the embedding chain at registration time; no flat fallback
`ClassDBBindMethodVirtual` gains a qualified-mode resolver: starting from `reflect.TypeFor[T]`, inspect T itself for `V_<T.Name()>_<Method>`; if absent, walk each embedded field (depth-first, following the single-Impl chain) looking for `V_<FieldType.Name()>_<Method>`; first hit wins (nearest = most-derived). When no level provides a qualified implementation, registration panics — there is deliberately no fallback to the exact flat name. Rationale: one map keyed by GDScript name, no new state, resolution once per class at startup, and a missing implementation is a loud authoring error rather than a silent style fork. Alternative — resolve at dispatch time via reflection on the instance — rejected: per-call cost and duplicates what registration already pins.

### Decision 3: Clean break — flat names panic instead of deprecating softly
If the resolver finds no qualified implementation for the requested virtual anywhere in the chain, or if a caller passes an unqualified `V_<Method>` Go name directly, registration panics with a diagnostic naming the GDScript virtual and the expected `V_<ClassName>_<Method>` shape. A soft transition (flat still working, "preferred" docs wording) was considered and rejected: the repo is explicitly experimental, in-repo usage is one demo class, and two live conventions would recreate exactly the ambiguity this change removes. Panicking at startup is deterministic, immediate, and self-explanatory.

### Decision 4: Super-delegation is plain Go, not machinery
Because each level's implementation has a distinct name, a derived implementation calls the base explicitly via the promoted selector (`e.CharacterBody2DImpl.V_CharacterBody2D_PhysicsProcess(delta)`). No runtime chaining, no token passing. This is the capability flat names made impossible; documenting the pattern is part of the deliverable.

### Decision 5: Demo coverage uses a two-level chain in test/pkg
Add a `TestHierarchicalNode`-style demo class extending `Node3DImpl` (or another simple leaf) where the Impl-level file defines `V_TestBase_Process`-style virtuals... concretely: register a small hierarchy — base-level qualified implementation on an intermediate wrapper struct plus a derived override — and assert from GDScript that the derived behavior runs while a second instance exercising only the base level proves shallow implementations still serve. Keep the demo minimal; the point is resolution semantics, not breadth.

## Risks / Trade-offs

- **Longer names** → more verbose declarations. Accepted: explicitness about *which* level implements a contract is the feature.
- **Reflection walk cost at registration** → negligible (once per bound virtual, chains are ~10 deep).
- **Hard break for downstream users** → any external code using flat `V_X` bindings stops working at startup, with a panic that states the fix. Accepted deliberately: the project warns its API is experimental and subject to change; the diagnostic makes migration mechanical.
- **Future generated defaults** → when Impl layers gain bodies, most-derived-wins keeps them as pure fallbacks; users override without shadowing anything. Verified by design, not by this change.

## Migration Plan

1. Ship resolver + panic-on-flat (this change); migrate every in-repo flat binding (the demo `Example` class) to qualified names in the same change so the suite starts clean.
2. Update docs/overview.md to present qualified names as the only supported convention.

Rollback: revert; flat registrations work exactly as before.

## Open Questions

- None blocking; naming-casing choice (Decision 1) is revisitable before apply if preferred style differs.
