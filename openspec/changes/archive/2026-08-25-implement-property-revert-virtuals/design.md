## Context

The revert callbacks are stubs (`pkg/core/classdb_callback.go:256-264`); their C trampolines, function-pointer types, and creation-info wiring already exist (classdb_callback.c/.h, classdb.go:533-534, ffi/class_create_info.go), so this is purely a Go-side dispatch implementation. The demo `Example` already binds `V_Example_PropertyCanRevert(p_name StringName) bool` as `_property_can_revert` (test/pkg/example.go:290, 844) — its body mirrors godot-cpp's test class (compares against `"property_from_list"`, true only while the stored value differs from `Vector3(42,42,42)`). `_property_get_revert` has no Go counterpart yet.

Working dispatch patterns to mirror live in the same file:

- `_get`/`_set` callbacks resolve the instance handle, look up `ci.VirtualMethodMap["_get"]`, build reflect args with the receiver first, call `GoMethodMetadata.Func.Call(args)`, and write results into Godot's out-params.
- The queried name arrives as a Godot-owned `GDExtensionConstStringNamePtr`; the `_set`/`_get` callbacks wrap it read-only and never destroy it.

godot-cpp reference semantics (test/src/example.cpp:166-183): `_property_can_revert(name) -> bool`; `_property_get_revert(name)` writes the default into an out-param and returns whether it was handled.

## Goals / Non-Goals

**Goals:**
- Both callbacks dispatch through `VirtualMethodMap` exactly like `_get`/`_set`.
- End-to-end demo coverage via engine-side `Object.property_can_revert()`/`property_get_revert()`.
- Classes without these virtuals keep working (negative answers, no panic).

**Non-Goals:**
- Script-instance-level property functions (`GDExtensionScriptInstanceProperty*`) — separate plumbing.
- Changing `_get_property_list`, `_validate_property`, or any other creation-info callback.
- Exposing new public API; virtuals are registered with the existing `ClassDBBindMethodVirtual`.

## Decisions

### Decision 1: Dispatch via VirtualMethodMap lookups keyed by GDScript names
Look up `_property_can_revert` / `_property_get_revert` in `ci.VirtualMethodMap`; when absent, return 0 immediately (no log spam per query). This matches `_get`/`_set` and keeps unimplemented-virtual behavior uniform. Alternative — a dedicated typed registration path — rejected: no benefit over the existing map.

### Decision 2: Keep StringName-typed virtual signatures; pass a borrowed wrapper
`V_Example_PropertyCanRevert` already takes `StringName`, so both new-path calls pass `reflect.ValueOf(*(*StringName)(pName))` wrapped from the Godot-owned pointer without copy-construction or destroy (borrow, per the refcount rules that guard StringName handling). `reflect.Call` requires exact parameter types, so the callback must match whatever the bound method declares; binding a `string`-typed variant would also work but would fork the two revert methods' styles for no gain.

### Decision 3: `V_Example_PropertyGetRevert(p_name StringName) (Variant, bool)` mirrors V_Get's shape
Two return values express "handled + value" naturally in Go and match how `V_Get` is dispatched today (check `reflectedRet[1].Bool()` before writing). Write `*(*Variant)(unsafe.Pointer(r_ret)) = v` only when handled, then return 1; otherwise leave `r_ret` untouched and return 0. Alternative — single `(Variant)` return with nil meaning unhandled — rejected: allocates a Variant on every miss and diverges from the established V_Get pattern.

### Decision 4: Example implements the godot-cpp test contract
`V_Example_PropertyGetRevert` returns `(Vector3(42,42,42), true)` for `"property_from_list"` and `(nil-Variant, false)` otherwise, mirroring godot-cpp's example.cpp so behavior is directly comparable across bindings. `V_Example_PropertyCanRevert` stays unchanged in body; its migration from the flat `V_PropertyCanRevert` name already landed with `implement-hierarchical-virtual-methods`.

### Decision 5: Notification virtual takes `(what int32, reversed bool)`
The creation-info callback receives both `p_what` and `p_reversed`; dropping `reversed` would discard information Godot already delivered (it distinguishes reversed scene-tree traversal order). GDScript's `_notification(what)` hides it, but Go bindings can afford fidelity here since the reflection dispatch matches whatever the bound method declares — and this change fixes that declaration once, so no per-user ambiguity exists. Missing `_notification` registration is a silent no-op (unlike revert queries, notifications are high-frequency; logging each miss would spam).

## Risks / Trade-offs

- **Borrowed StringName misuse** → a future edit could destroy the Godot-owned name. Mitigation: follow the existing comment convention at the wrap site ("Godot-owned — NEVER Destroy").
- **Deprecated engine API in assertions** → `Object.property_can_revert()`/`property_get_revert()` are deprecated-but-functional in Godot 4.x; if removed upstream, assertions migrate to inspector-equivalent queries. Mitigation: noted here; failure mode is loud.
- **Variant ownership on the write-back** → mirrors the proven `_get` write-back; if leaks surface, the fix applies to both paths symmetrically.

**Boundary with `implement-generated-virtual-defaults`:** this change's absence semantics (missing `_notification` = silent no-op, unimplemented revert virtuals report no revert support) are pinned by its specs. If that change later extends generated defaults beyond Category A into Categories B/C, such defaults must never auto-register into `VirtualMethodMap` — only via its opt-in hook — or they would turn every absence this change relies on into presence.

## Implementation Pitfalls (regression guards)

These were hit during implementation; the suite pins them, but the reasoning lives here:

- **Go cgo composite-object scan**: builtin calls that pass a pointer *into* a composite Go object (e.g. `e.propertyFromList.Equal_Vector3(...)` taking `&field`) make Go's cgo checker scan the entire containing allocation. Once instances hold unpinned Go references (the `NotificationCodes` slice), such calls abort with "argument of cgo function has Go pointer to unpinned Go pointer". Rule for virtual method bodies: copy scalar/builtin fields to locals before passing them through generated wrappers. Pinned by the revert-assertions running after notification recording.
- **Godot 4.7 renamed `Object.notify` → `Object.notification`**: calling the removed name on an extension instance triggers a pre-existing engine-side leak (2 ObjectDB instances + 1 CanvasItem RID per call) while assertions still pass — leak warnings do not fail the suite today. Demo coverage uses `notification(what, reversed)`. If leak-sensitive CI is wanted, `make test` would need to fail on `ObjectDB instances were leaked` output (not currently enforced).

## Migration Plan

Additive; no migration. Rollback is a revert.

## Open Questions

- None blocking; callback-to-engine routing is confirmed by the existing creation-info wiring.
