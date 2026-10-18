# Tasks

## 1. Engine-path enumeration (design D4)

- [ ] 1.1 Enumerate every `get_property_list_func` / `free_property_list2` call site in `../godot` (`object.cpp`, `class_db.cpp` placeholder path, TOOLS/`clear_internal_extension` branches) and record the exposure/release pairing table in this file's Notes; verification: table present, every exposure has a same-scope or documented-later paired release; if any exposure lacks a release, stop and revisit design D1 fallback before continuing.

## 2. Class-level string cache and descriptors (design D2/D3)

- [ ] 2.1 Introduce the Go-side property descriptor and the class-level engine-string cache on `ClassInfo`; change `NewClassInfo`'s `propertyList` input to descriptors and update all callers, including `cmd/generate` templates if generated registration builds the old type (regenerate with `make generate`, never hand-edit `.gen.*`); verification: `go build ./...` and `make test` property tests unchanged.
- [ ] 2.2 Destroy the cache exactly once in `ClassInfo.Destroy` and delete the nil-check double-destroy dance (`pkg/core/types.go:67-77`); verification: a Go-level or demo check tears down a class registration after zero queries, after one query, and after many queries with no double-free.

## 3. Per-query C array (design D1)

- [ ] 3.1 Rework `GoCallback_ClassCreationInfoGetPropertyList` to `calloc` a fresh `GDExtensionPropertyInfo` array per query from the cache and return it with no `pnr.Pin`; null-allocation returns count 0; verification: first-query property contents match pre-change output in the demo.
- [ ] 3.2 Rework `GoCallback_ClassCreationInfoFreePropertyList2` to `C.free` the array only (no string destroys, no `ci.PropertyList = nil`); verification: existing single-query property tests still pass.
- [ ] 3.3 Add GDScript driver `test_property_list_persistence` in `test/demo/main.gd` querying `get_property_list()` twice and asserting the list-declared custom property and identical metadata in both reads (this closes the coverage gap that hid the bug); verification: new asserts report in PASSES count and fail when 3.2 is reverted (one-off manual check, then restore).

## 4. Retention gate

- [ ] 4.1 Add a Go churn fixture in `test/pkg` plus a GDScript driver driving thousands of query/free cycles with band assertions sized per batch (reuse the `instance_churn_tests.go` pattern, including `runtime.ReadMemStats` band, not absolutes); verification: gate green on the change and a deliberate `C.free` removal makes it fail.

## 5. Integration

- [ ] 5.1 Full two-stage gate: `GODOT=/home/pcting/bin/godot make test` — PASSES ≥ 1386 baseline plus new asserts, FAILURES=0, no ObjectDB leak errors, editor stage green.
- [ ] 5.2 `openspec validate fix-class-property-list-lifecycle --strict` passes; update `docs/overview.md` property-list lifetime description to the query-owned-array / class-owned-strings split.

## Notes

- Baseline: `main` at `4f823d4` (#166 merged) — 1386 PASSES, binding counters balanced at churn.
- Pairing table from 1.1 goes here.

## Workflow follow-up

- Commit pattern per the `opening-pull-requests` skill: `docs: propose ...` (done) → implementation → `docs:` → `chore: archive ... change`; archive syncs the `class-property-list` and `native-call-pinning` deltas into `openspec/specs/`.
- PR body = archived `proposal.md` + `---` + `design.md`; no issue exists yet — if one is filed for the vanishing-properties bug, append `Closes #NNN`.
