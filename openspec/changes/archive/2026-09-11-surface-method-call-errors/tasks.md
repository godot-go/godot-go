## 1. Strict argument-conversion check

- [x] 1.1 Add a strict-conversion check on `GoMethodMetadata`: for a marshalled argument `Variant` and slot `i`, report convertibility via `variant_can_convert_strict(supplied.GetType(), gdeArgumentTypes[i])` (helper on the metadata, engine predicate parity with godot-cpp)
- [x] 1.2 Document (in the callback) that marshalled argument `Variant`s are non-owning bit-copies and the reject path frees nothing — calling `Destroy()` would free engine-owned memory
- [x] 1.3 Add a pure-Go unit test for the arity classifier in group 2 (engine-independent); strict-conversion behavior is engine-bound and validated in group 5 via the harness that `make test` runs

## 2. Arity classification (default- and variadic-aware)

- [x] 2.1 Add a required/declared arity helper on `GoMethodMetadata` derived from `len(gdeArgumentTypes)`, `len(gdeDefaultArgumentPtrs)`, and `IsVariadic`, matching `Call`'s fill condition `max(supplied, defArgsCount) < declared`
- [x] 2.2 Implement a classify-arity method returning `{OK, TOO_MANY, TOO_FEW}` with the `expected` count, treating too-many as fatal only when not variadic (pure integer logic, no engine calls)
- [x] 2.3 Unit-test arity classification (pure `go test ./pkg/core`): exact match, too many, too many (variadic OK), too few beyond defaults, short call satisfied by all-defaults method

## 3. Type-mismatch classification via strict conversion

- [x] 3.1 Implement classify-arguments over the marshalled argument views: first non-convertible slot (per 1.1) yields `INVALID_ARGUMENT` with the failing index and `gdeArgumentTypes[index]` as expected — no reflect staging, no conversion performed, nothing to release (views are non-owning)
- [x] 3.2 Exercise non-convertible-argument reporting through the harness in group 5 (engine predicate cannot run in a bare unit test)

## 4. Wire the varcall callback

- [x] 4.1 In `GoCallback_MethodBindMethodCall`, classify arity before allocating argument `Variant` views; on arity reject write `r_error` (code+expected), set `r_return` to a nil `Variant`, and return
- [x] 4.2 After building argument views, run classify-arguments; on type reject write `r_error` with `INVALID_ARGUMENT`+argument+expected, set `r_return` to nil, and return without invoking (free nothing — views are non-owning)
- [x] 4.3 On accepted calls reuse the existing `bind.Call` invoke path unchanged; set `r_error.error` to `GDEXTENSION_CALL_OK`
- [x] 4.4 Remove the too-few-args `log.Panic` in `GoMethodMetadata.Call` (now covered by arity reporting); keep `log.Panic` only for null method user data and null instance

## 5. Test coverage in the demo harness

- [x] 5.1 Add a bound Go method with a fixed signature (side-effect flag it sets when it runs) plus an all-arguments-defaulted method, both callable from GDScript
- [x] 5.2 Add `main.gd` cases: too-many and too-few-beyond-defaults calls assert the process survives and the side-effect flag stayed unset
- [x] 5.3 Add a wrong-type (non-convertible, e.g. Array into an int param) case and an Array-argument reject case, asserting survival and that the `make test` leak gate stays green
- [x] 5.4 Add a defaults-satisfy-short-call case and a variadic-too-many case asserting success (side-effect flag set)

## 6. Validate

- [x] 6.1 `go vet ./pkg/... ./test/pkg/...` clean and `go test ./pkg/core/` green for the new unit tests
- [x] 6.2 `GODOT=/path/to/godot make build` and `make test` green, including new cases and leak gate
- [x] 6.3 `openspec validate surface-method-call-errors --strict` passes
- [x] 6.4 Update `docs/overview.md` with a short note on call-error semantics for varcall binds
