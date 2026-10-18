# Proposal

## Why

The class property-list callbacks are leaky and functionally broken after the first engine query: `GoCallback_ClassCreationInfoGetPropertyList` pins the class-level `ci.PropertyList` array on the program-lifetime pinner `pnr` and never unpins, while `GoCallback_ClassCreationInfoFreePropertyList2` destroys every element's strings and nils the slice — and nothing ever rebuilds it. The engine copies the returned array and calls `FreePropertyList2` back-to-back in one scope (`Object::get_property_list`, object.cpp:575-584), so after the first query, property-list-declared custom properties silently vanish from every subsequent `get_property_list()` (inspector refreshes, serialization), and the destroyed array stays pinned for the process lifetime (accumulating again on each re-registration, e.g. editor hot-reload).

## What Changes

- Class property lists are rebuilt per query from persistent Go-side property definitions and exposed as a C-allocated `GDExtensionPropertyInfo` array; `FreePropertyList2` destroys the elements and frees the array. Zero Go memory crosses the boundary, so no `pnr` pin is involved.
- `ClassInfo` no longer stores the published `[]GDExtensionPropertyInfo` array; it keeps the source-of-truth definitions from which each query builds.
- The double-destroy dance between `ClassInfo.Destroy()` and `FreePropertyList2` (`pkg/core/types.go:67-77`) is removed with the shared-array ownership it guarded.
- New demo gates: a GDScript driver that queries `get_property_list()` twice and asserts `ci`-sourced custom properties survive both reads, plus a churn/retention assert proving repeated queries do not grow pinned retention.

## Capabilities

### New Capabilities
- `class-property-list`: the get/free property-list callback contract for registered classes — properties remain queryable across repeated engine queries, and the exposed array is owned and released by the paired free callback.

### Modified Capabilities
- `native-call-pinning`: memory exposed to the engine across a paired get/release callback sequence is pinned on a pinner whose lifetime matches that pair and released by the release callback — never accumulated on the program-lifetime pinner.

## Impact

- `pkg/core/classdb_callback.go` (`GetPropertyList`, `FreePropertyList2`), `pkg/core/types.go` (`ClassInfo`, `Destroy`), property-definition construction feeding `NewClassInfo` (possible `cmd/generate` template and generated-registration signature changes).
- `test/demo/main.gd` and `test/pkg` gain the double-query and retention gates.
- No public GDScript-visible API change; fixes behavior GDScript already expects.

## Non-goals

- The broader audit of remaining `pnr.Pin` sites (registration-time pins stay; per-site re-scoping of anything else is a separate follow-up).
- `property-hint-registration` (ClassDB `add_property`) path — engine-side storage, unaffected.
- Per-instance property lists; class-level scope only.
