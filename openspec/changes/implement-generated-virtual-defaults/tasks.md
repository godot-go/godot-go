## 1. Investigation spike

- [ ] 1.1 Census `extension_api.json`: classify all 1437 virtuals into categories A (engine GDVIRTUAL), B (extension lifecycle), C (creation-info routed); emit the Category-A list (class + method + signature) to a checked-in reference file; record counts in this change's design.md Open Questions resolution
- [ ] 1.2 Write a failing recursion repro first: a demo class overriding a Category-A virtual that delegates through the plain wrapper; confirm the infinite-recursion trap exists today (guard with a depth counter, expect overflow) — this pins the bug the change fixes

## 2. Codegen

- [ ] 2.1 Extend `cmd/generate` templates to emit Category-A default bodies on Impl layers: single plain-wrapper delegation statement per body under qualified `V_<ClassName>_<MethodName>` names
- [ ] 2.2 Emit the opt-in registration hook per class (codegen marker the resolver consults) per design Decision 2
- [ ] 2.3 Run `make generate`; verify `go build ./pkg/gdclassimpl/...` compiles and emitted bodies pass the one-wrapper-call lint check from design Risks
- [ ] 2.4 Verify generated surface completeness against task 1.1's census file: every Category-A entry has an emitted counterpart (scripted assertion in codegen tests)

## 3. Resolution wiring

- [ ] 3.1 Wire registration-time resolution to consult the opt-in hook: no user implementation + opted-in → bind generated shallowest default; otherwise keep absence reporting unchanged
- [ ] 3.2 Verify `go build ./pkg/core/...` and existing suite green (no behavior change for unopted classes)

## 4. Demo coverage

- [ ] 4.1 Convert task 1.2's repro into a passing regression test: leaf override delegates via promoted selector to generated default and returns engine-equivalent value without recursion
- [ ] 4.2 Add GDScript assertions: opted-in class with override extends engine default; non-opted class still answers via engine fallback (presence-falsification sensitivity check)
- [ ] 4.3 Update docs/overview.md virtual-methods section: generated defaults, delegation pattern, opt-in semantics

## 5. Validate

- [ ] 5.1 Run `go vet ./pkg/... ./test/pkg/...`
- [ ] 5.2 Run `GODOT=/path/to/godot make build` and `make test`; full suite green including new assertions
