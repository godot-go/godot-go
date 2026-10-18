# Tasks

## 1. Engine-path enumeration (design D4)

- [x] 1.1 Enumerate every `get_property_list_func` / `free_property_list2` call site in `../godot` (`object.cpp`, `class_db.cpp` placeholder path, TOOLS/`clear_internal_extension` branches) and record the exposure/release pairing table in this file's Notes; verification: table present, every exposure has a same-scope or documented-later paired release; if any exposure lacks a release, stop and revisit design D1 fallback before continuing.

## 2. Class-level string cache and descriptors (design D2/D3)

- [x] 2.1 Introduce the Go-side property descriptor and the class-level engine-string cache on `ClassInfo`; change `NewClassInfo`'s `propertyList` input to descriptors and update all callers, including `cmd/generate` templates if generated registration builds the old type (regenerate with `make generate`, never hand-edit `.gen.*`); verification: `go build ./...` and `make test` property tests unchanged.
- [x] 2.2 Destroy the cache exactly once in `ClassInfo.Destroy` and delete the nil-check double-destroy dance (`pkg/core/types.go:67-77`); verification: a Go-level or demo check tears down a class registration after zero queries, after one query, and after many queries with no double-free.

## 3. Per-query C array (design D1)

- [x] 3.1 Rework `GoCallback_ClassCreationInfoGetPropertyList` to `calloc` a fresh `GDExtensionPropertyInfo` array per query from the cache and return it with no `pnr.Pin`; null-allocation returns count 0; verification: first-query property contents match pre-change output in the demo.
- [x] 3.2 Rework `GoCallback_ClassCreationInfoFreePropertyList2` to `C.free` the array only (no string destroys, no `ci.PropertyList = nil`); verification: existing single-query property tests still pass.
- [x] 3.3 Add GDScript driver `test_property_list_persistence` in `test/demo/main.gd` querying `get_property_list()` twice and asserting the list-declared custom property and identical metadata in both reads (this closes the coverage gap that hid the bug); verification: new asserts report in PASSES count and fail when 3.2 is reverted (one-off manual check, then restore).

## 4. Retention gate

- [x] 4.1 Add a Go churn fixture in `test/pkg` plus a GDScript driver driving thousands of query/free cycles, asserting release accounting via `pkg/core` alloc/free counters (the arrays are C memory — `runtime.ReadMemStats` cannot see a missing `C.free`) alongside a `runtime.ReadMemStats` band proving no Go retention (reuse the `instance_churn_tests.go` pattern); verification: gate green on the change and a deliberate `C.free` removal makes it fail.

## 5. Integration

- [x] 5.1 Full two-stage gate: `GODOT=/home/pcting/bin/godot make test` — PASSES ≥ 1386 baseline plus new asserts, FAILURES=0, no ObjectDB leak errors, editor stage green.
- [x] 5.2 `openspec validate fix-class-property-list-lifecycle --strict` passes; update `docs/overview.md` property-list lifetime description to the query-owned-array / class-owned-strings split.

## Notes

- Baseline: `main` at `4f823d4` (#166 merged) — 1386 PASSES, binding counters balanced at churn.
- Pairing table from 1.1 (godot @ checked-out master):

  | Exposure | Release | Notes |
  | --- | --- | --- |
  | `object.cpp:577` `current_extension->get_property_list(_extension_instance,&pcount)` | `free_property_list2` :582 (else deprecated `free_property_list` :586), same scope after deep-copy | TOOLS `is_placeholder` guard :571; the ONLY caller of class get_property_list in the codebase |
  | `free_property_list2(` invocations | — | Exist only at object.cpp:582; no stale-free site anywhere |
  | `class_db.cpp:143-149` placeholder get/free stubs | paired via same path | Engine stand-in (get returns null/0); never reaches extension arrays |
  | `script_language_extension.h:818` `native_info->get_property_list_func` | script-instance callback family | Separate typedef; godot-go registers no script instances — unaffected |
  | TOOLS `Object::clear_internal_extension` | no property-list calls | Confirmed no `free_property_list` invocation there |

  Conclusion: design D1 primary form safe — every exposure pairs an in-scope release.
- Gate evidence (two-stage `make test`): green run 1417 PASSES / 0 FAILURES (1386 baseline + 31 new asserts), editor stage PASS, no ObjectDB leaks; teardown check `zero/one/many-query` clean, churn `arrays_outstanding` balanced, `heap_delta` 0.
- Mutation checks (one-off, both reverted, restore `diff`-proven identical): list-wipe-after-first-query → persistence gate fails (8 failures, `first/second.has()` × 4 properties); release-callback no-op → churn gate fails (imbalance +4000 = exact batch) while the Go-heap band stayed +2 KiB of 250 KiB, confirming counters are the load-bearing release gate for C memory.

## Review round (post-archive addendum, PR #167)

Review verdict: *with fixes* — 2 Important, 6 Minor. Review comment: `https://github.com/godot-go/godot-go/pull/167#issuecomment-6246660864`.

Resolved in `fix: reseat churn verify and close property-list review gaps` (gate: 1425 PASSES / 0 FAILURES, editor stage green, no orphaned StringNames):

- [x] Important 1: `test_instance_churn_verify()` had been stranded inside the new `test_property_list_churn` fixture by the insertion anchor; moved back to the end of `test_instance_churn` so the #152 gate stays owned by its own fixture.
- [x] Important 2: `docs/overview.md` now states in changed-from framing that `[]PropertyDescriptor` replaces the `[]GDExtensionPropertyInfo` parameter and the removed `ClassInfo.PropertyList` field.
- [x] Minor 2: `(*GDExtensionPropertyInfo).Destroy()` documented creator-only — never callable on class-property-list elements, whose cells belong to `ClassInfo.Destroy`.
- [x] Minor 3: new `property_hint_string` descriptor with non-empty `HintString: "review-carrier"`, content-asserted in both persistence reads — the "Late query still reports the hint string" scenario now has a positive end-to-end check (+8 asserts).

Polish follow-ups (no behavior change), implemented after the deferral note in the same follow-up round (`refactor: polish class property cache internals`; gate re-run identical: 1425 PASSES / 0 FAILURES, no leaks):

- [x] `allocPropertyArray` returns `(*C.GDExtensionPropertyInfo, []C.GDExtensionPropertyInfo)` but both callers discard the slice — trimmed to a pointer return; the slice stayed an internal fill detail.
- [x] `classPropertySpec.destroy()` re-fetched both variant destructors on every call, including fully-nil idempotent teardowns — destructors now resolve once per process (`sync.Once`) and nil specs early-return with zero engine calls.
- [x] `cellSize` renamed to `stringCellSize` with a comment stating the pointer-cell model; sizing expressed via `unsafe.Pointer` instead of `uintptr(0)`.
- [x] `RunPropertyCacheTeardownCheck` now returns `error` — each failure phase carries its own reason (spec count, failed cycle index, balance mismatch); the demo fixture prints it.

Also corrected the stale `newClassPropertySpec` docstring, which claimed the caller reports an empty property list on allocation failure when `buildClassPropertySpecs` actually releases and panics.

## Workflow follow-up

- Commit pattern per the `pull-request` skill: `docs: propose ...` (done) → implementation → `docs:` → `chore: archive ... change`; archive syncs the `class-property-list` and `native-call-pinning` deltas into `openspec/specs/`.
- PR body = archived `proposal.md` + `---` + `design.md`; no issue exists yet — if one is filed for the vanishing-properties bug, append `Closes #NNN`.
