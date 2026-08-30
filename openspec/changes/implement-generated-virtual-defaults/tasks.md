## 1. Investigation spike

- [x] 1.1 Census `extension_api.json`: classify all 1437 virtuals into categories A (engine GDVIRTUAL with plain wrapper), B (remaining lifecycle/side-effect hooks), C (creation-info routed); emit the Category-A list to a checked-in reference file; record counts in this change's design.md Open Questions resolution
- [x] 1.2 Write a failing recursion repro first: a demo class overriding a Category-A virtual that delegates through the plain wrapper; confirm the infinite-recursion trap exists today (guard with a depth counter, expect overflow) — this pins the bug the change fixes
- [x] 1.3 Prototype the delegation-termination mechanism before any template work — prototype falsified both variants (callback-layer guard yields zero values because Godot freezes presence per instance; suppression windows cannot reach the per-instance cache). Findings drove the pivot to declaration-only scope (design Context/Decision 3); `SuppressVirtualWhile` prototype reverted

## 2. Codegen

- [ ] 2.1 Extend `cmd/generate` gdclassimpl templates to derive the Category-A classification directly from `extension_api.json` (same criterion as the census fixture) and emit per-class virtual-surface interfaces: one interface per class, qualified names, exact signatures, declaration-only
- [ ] 2.2 Run `make generate`; verify `go build ./pkg/gdclassimpl/...` compiles and emitted output contains declarations only (lint: no function bodies generated for virtual-surface interfaces)
- [ ] 2.3 Verify generated surface completeness against the census fixture (`cmd/generate/gdclassimpl/virtual_census.json`): every Category-A entry has a declared counterpart with matching signature, and vice versa (scripted assertion in codegen tests)

## 3. Resolution wiring

- [ ] 3.1 Verify unbound catalog leaves behavior untouched: full suite green with regenerated interfaces present and no additional virtuals registered
- [ ] 3.2 Pin the recursion trap as a make-target expectation test — a normal suite test cannot observe a cgo-callback panic: separate godot invocation exercising the delegating repro, expecting non-zero exit and grepping for the bounded-depth recursion diagnostic

## 4. Demo coverage & docs

- [ ] 4.1 Add a compile-time conformance example: demo struct asserting a generated virtual-surface interface for its override, proving the signature-verification idiom builds
- [ ] 4.2 Update docs/overview.md virtual-methods section: generated catalog, conformance idiom, and the delegation-impossibility constraint with evidence trail

## 5. Validate

- [ ] 5.1 Run `go vet ./pkg/... ./test/pkg/...`
- [ ] 5.2 Run `GODOT=/path/to/godot make build` and `make test`; full suite green including new assertions
