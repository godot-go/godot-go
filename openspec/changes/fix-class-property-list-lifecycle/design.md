# Design

## Context

See proposal.md - Why. Current state and constraints:

- The engine's consumer (`Object::get_property_list`, object.cpp:575-584) calls `get_property_list`, deep-copies each `GDExtensionPropertyInfo` element into its own `List<PropertyInfo>`, and calls the paired `FreePropertyList2` in the same scope before returning. The exposed pointer's true lifetime is "get until the paired free2", and free2 **reads** the array — so the array must stay valid across two separate Go callback invocations, which makes a function-local pinner in `GetPropertyList` unsound.
- Today `ci.PropertyList` is built once at class registration (`NewClassInfo`), pinned once per exposure on the program-lifetime `pnr`, destroyed and nil-ed by the first `FreePropertyList2`, and never rebuilt. A `ClassInfo.Destroy` nil-check dance guards the resulting double-destroy window (`pkg/core/types.go:67-77`).
- `GDExtensionPropertyInfo` elements are plain C structs whose string fields point at engine-allocated `StringName`/`String` objects; the engine copies them by value and never frees them itself — the extension owns the strings behind any array it exposes.
- `native-call-pinning` spec already requires pinner lifetimes to match engine retention; the existing `pin_scoping_test.go` guardrail shows how this repo enforces pin discipline statically.

## Goals / Non-Goals

**Goals:**
- Repeated `get_property_list()` queries return identical, valid custom properties.
- No Go memory and no `pnr` pin on the exposure path; every allocation released by the paired free callback.
- Remove the `ClassInfo.Destroy`/`FreePropertyList2` double-destroy coordination.

**Non-Goals:**
- Caching arrays across queries or copy-on-change optimization (queries are inspector/serialization frequency, not per-frame).
- Touching the `add_property`/ClassDB hint path or other `pnr` sites (proposal Non-goals).

## Decisions

### D1: Per-query rebuild into a C-allocated array; free2 frees the array
`GetPropertyList` allocates the struct array with `C.calloc` from the class's cached descriptors and returns it; `FreePropertyList2` calls `C.free` and destroys nothing else. Alternative considered: keep the Go slice and attach a per-list `runtime.Pinner` unpinned in free2 (rejected: it cements the nil-after-first-fetch breakage, and unpin coordination has to thread `ClassInfo.Destroy`, free2, and hot-reload paths). Also rejected: per-query Go array plus a package-level pin registry keyed by pointer (more machinery than a `calloc`).

This matches godot-cpp's shape: expose freshly materialized C memory per query, release on the paired free.

### D2: Class-level immutable string cache; per-query struct copies only
`ClassInfo` stores, per list property, the engine-allocated `StringName`/`String` handles for name/class_name/hint_string plus the type/hint/usage scalars, created once at registration. Each query copies these pointers by value into the fresh C array; the strings outlive every get/free2 pair and are destroyed exactly once in `ClassInfo.Destroy`. Alternative considered: rebuild strings per query and destroy them in free2 (godot-cpp's exact shape — rejected as needless per-query engine-string churn once the scalars-plus-cached-pointers variant is provably sufficient: the engine deep-copies strings into its own `PropertyInfo`). This is what deletes the double-destroy dance: ownership is split cleanly — array owned by the query, strings owned by the class.

### D3: `ClassInfo` holds descriptors, not published arrays
`NewClassInfo`'s `propertyList []GDExtensionPropertyInfo` parameter becomes a Go-side descriptor slice (name, class, hint string as Go strings or handles, type, hint, usage). Registration materializes the D2 cache from it. If generated registration code constructs the current parameter type, update the `cmd/generate` templates and `make generate` — never hand-edit `.gen.*` (AGENTS.md).

### D4: Gates
- GDScript driver: query `example.get_property_list()` twice; both reads must contain the `Example` list-declared property (`property_from_list`) with identical name/type/hint/usage — closes the coverage gap that let the regression stand.
- Retention gate reusing the fix-instance-handle-lifecycle churn pattern: N query/free cycles keep live-heap-in-band and make no `pnr` growth observable; assert allocation balance rather than absolutes.
- Engine-path enumeration before implementation: list all `get_property_list_func` call sites in `../godot` (including the `class_db.cpp:143` placeholder path) and confirm every one pairs a free2 on the same object/thread; the placeholder branch is engine-guarded by `is_placeholder`.

## Risks / Trade-offs

- [Engine path pairs free differently than enumerated] → D4 enumeration first; if any path skips free2, the leak is bounded to one calloc per query — escalate to per-class outstanding-array registry (one slot, freed on next query) as fallback.
- [calloc-per-query churn] → accepted; same order as godot-cpp's per-query `List<PropertyInfo>` construction, driven by inspector/serialization, not hot loops.
- [String cache lifetime vs hot-reload `clear_internal_extension`] → cache destroyed only in `ClassInfo.Destroy`; reload re-registers a fresh `ClassInfo`. Verify against the editor-stage gate.
- [Null calloc] → return count 0 and null pointer; engine treats it as no custom properties.

## Migration Plan

Single self-contained change on `pkg/core` internals; no user-visible API or saved-data migration. Rollback = revert the change.

## Open Questions

- Whether the double-query gate reuses `Example` or adds a minimal `ExamplePropertyList` fixture — either satisfies D4; decide when touching the test package.
- Exact descriptor struct field set (Go strings vs pre-made handles at registration input) — internal, resolvable during implementation without touching specs.
