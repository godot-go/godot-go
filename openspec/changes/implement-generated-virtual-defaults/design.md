## Context

The virtual surface: 1437 virtuals across 106 of 1036 classes in `godot_headers/extension_api.json`. Post-`implement-hierarchical-virtual-methods`, resolution walks the embedding chain at registration and picks the most-derived qualified implementation; the callback layer (`get_virtual_call_data2`/`call_virtual_with_data`) is name-agnostic. Generated `*Impl` layers currently declare zero virtual bodies — every default is user-authored.

Not all virtuals are alike. Three categories with different "default" semantics:

| Category | Examples | Unregistered behavior | Engine default source |
|---|---|---|---|
| A. Engine GDVIRTUAL | `_get_maximum_size`, `_get_minimum_size` | engine C++ body runs (pinned by `gdvirtual-unimplemented-defaults`) | Godot core |
| B. Extension lifecycle | `_ready`, `_process`, `_input` | nothing runs (implicit no-op) | none |
| C. Creation-info routed | `_notification`, `_property_can_revert`, `_get`, `_set` | dedicated callback answers negatively/no-op (accurate once `implement-property-revert-virtuals` lands; today the callbacks are unconditional stubs that answer negatively for registered classes too) | none |

**The delegation trap (verified against dispatch architecture and Godot 4.7 source):** a leaf override of a Category-A virtual cannot safely reach the engine body today. Plain wrapper → MethodBind → engine C++ → `GDVIRTUAL_CALL(_x)` → extension call-data lookup → VirtualMethodMap → *the leaf override again*: infinite recursion. Note the engine side routes through `_get_extension()->call_virtual_with_data(...)` on **every** `GDVIRTUAL_CALL` (`gdvirtual.gen.h`; e.g. `Control::get_maximum_size()` at `control.cpp:1814`) with no upstream recursion guard — so re-entry resolves to *whatever is registered* (the leaf, or later a generated default), not to "nothing below". Reaching engine behavior therefore requires the binding itself to break the cycle; see Decision 4. This is the concrete functional hole generated defaults close.

## Goals / Non-Goals

**Goals:**
- Typed, discoverable virtual surface: every declared virtual has a generated, greppable Go counterpart on its Impl layer.
- Recursion-free engine-default delegation for Category A.
- User overrides win by chain depth; generated defaults are pure fallbacks.

**Non-Goals:**
- Auto-invocation chaining; replicating engine bodies in Go where delegation suffices; changing dispatch or callback layers; eager registration of high-frequency virtuals without a performance budget.

## Decisions

### Decision 1: Generate defaults only where they change behavior — Category A first
Category A gets bodies whose single statement delegates to the plain method wrapper (recursion-safe per the Context analysis). Category B/C defaults would be explicit no-ops that add presence-reporting cost (Godot calls into Go on every event) with zero behavioral gain — not generated initially. Coverage metric therefore targets Category A's subset of the 1437, documented explicitly, rather than raw totals.

### Decision 2: Emitted defaults register lazily — presence follows implementation
Registration-time resolution already reports absence when no level implements a virtual; generated bodies live at the shallowest level but must NOT auto-register for instances that never opt in... resolved as: `ClassDBRegisterClass` binds a generated default only when the class's own resolution walk finds no user implementation AND the class opts in via a codegen-emitted hook. Rationale: registering Category-A defaults globally would make `get_virtual_call_data2` report presence for every subclass, silently replacing engine fallbacks with Go round-trips (perf + drift risk). Alternatives rejected: eager global registration (presence-falsification), never registering (defaults exist but dead).

### Decision 3: Delegation contract is the promoted selector, documented over new API
`t.ControlImpl.V_Control_GetMaximumSize()` reaches the generated body; the generated body calls `t.GetMaximumSize()` once. No new public API; the pattern is docs/overview.md material.

### Decision 4: Delegation terminates via a CallVirtualWithData re-entrancy guard
A generated default (or a leaf override delegating through it) whose body calls the plain wrapper re-enters the engine, and Godot's `GDVIRTUAL_CALL` routes that re-entry back through `call_virtual_with_data` unconditionally — resolving to the *registered* implementation again. Termination therefore cannot come from Go-side call depth; the binding must break the cycle itself: `GoCallback_ClassCreationInfoCallVirtualWithData` tracks in-flight (instance id, method) pairs and treats a re-entrant lookup as **absence**, letting Godot's `GDVIRTUAL_CALL` yield false so the engine's own default body computes the value that unwinds back to the delegating Go frame. This delivers extend-the-engine-default semantics for both the override case and the opted-in no-override case, with one map check per virtual dispatch.

## Risks / Trade-offs

- **Presence-falsification** (registering defaults makes Godot skip engine fallbacks) → mitigated by Decision 2's conditional binding; sensitivity test asserts unopted-in classes keep engine-default behavior.
- **Codegen volume** → Category A only; measured before template work (spike task).
- **Drift if a generated body ever replicates instead of delegating** → lint-style check: emitted Category-A bodies must contain exactly one wrapper call.

**Constraint on future scope extension:** if generated defaults are ever extended beyond Category A into Categories B/C, they MUST NOT auto-register into `VirtualMethodMap` — the opt-in hook semantics apply there too. The `property-revert-virtuals` and `notification-virtual-dispatch` specs pin absence semantics ("missing `_notification` = silent no-op", "unimplemented revert virtuals report no revert support"); an auto-registered default would violate them by turning every absence into presence.

## Migration Plan

Additive. Generated files land under `pkg/gdclassimpl/*.gen.go`; existing user classes unaffected until they delegate.

**Implementation order:** land `implement-property-revert-virtuals` first. It is smaller and self-contained, and this change's spike (task 1.1) plus demo assertions benefit from its stabilized `test/demo/main.gd` assertion suite and accurate Category-C behavior column above.

Rollback: regenerate with templates disabled.

## Open Questions

- Exact Category-A census (which of the 1437 route through GDVIRTUAL vs creation-info) — spike task 1.1 settles this from `extension_api.json` before templates are written.
- Whether any Category-B virtual needs an opt-in default for ecosystem reasons (e.g., `_physics_process` bookkeeping) — deferred unless demoed need.
- Why must an opted-in class *without* overrides serve the generated Go default at all, given the engine answers identically natively at zero cost? The spec's opted-in scenario mandates the generated body runs; if pure delegation is all it ever does, consider relaxing that scenario (or documenting motivating future behavior) before templates are written.
