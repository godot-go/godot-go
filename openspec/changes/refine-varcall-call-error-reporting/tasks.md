# Tasks: refine-varcall-call-error-reporting

## 1. Align too-few `expected` with godot-cpp

- [x] 1.1 In `pkg/core/method_bind_call_error.go`, change `classifyVarcallArity` to return the declared argument count (`len(md.gdeArgumentTypes)`) for `varcallArityTooFew` instead of `required = declared - defArgs`; drop the now-unused required computation and update the doc comment to cite godot-cpp `call_with_variant_args_dv` (`expected = sizeof...(P)` for both arity errors)
- [x] 1.2 Update `TestClassifyVarcallArity` in `pkg/core/method_bind_call_error_test.go`: too-few cases now expect the declared count (e.g. declared=2 defaults=0 supplied=1 → expected 2; declared=3 defaults=1 supplied=1 → expected 3)
- [x] 1.3 Verify `TestClassifyVarcallArityMatchesCallFill` still passes unchanged (it asserts only the OK/too-few boundary, not the `expected` value)

## 2. Lockstep and caveat comments

- [x] 2.1 In `classifyVarcallArity`, add a comment stating the condition must equal `GoMethodMetadata.Call`'s fill satisfiability, that partial trailing-default bindings (`0 < defaults < declared`) are rejected under the leading-index model, and that moving the fill model to the trailing convention requires changing the check to `declared - supplied > defaults` and updating `TestClassifyVarcallArityMatchesCallFill` in the same commit
- [x] 2.2 In `GoMethodMetadata.Call`'s argument-fill loop (`pkg/core/method_bind.go`), extend the unfilled-slot comment to note the exported-surface caveat: direct Go callers bypass varcall validation and receive zero-value `Variant`s (≈ NIL) in unfilled slots instead of the removed panic

## 3. Debug diagnostic on reject paths

- [x] 3.1 Pass the `*GoMethodMetadata` (or bound method name) into `rejectVarcallArity` and `rejectVarcallInvalidArgument` in `pkg/core/method_bind_callback.go` and emit `log.Debug` with the bound method, error kind, and the `argument`/`expected` values written (arity reject: method + kind + expected; type reject: method + kind + argument index + expected variant type)
- [x] 3.2 Confirm the accepted path is unchanged (no new log on success beyond the existing debug log)

## 4. Documentation corrections

- [x] 4.1 In `docs/overview.md`, replace the ambiguous "numeric/string inter-conversion ... still pass" phrasing with the precise strict-conversion set: numeric inter-conversion (int/float/bool), STRING↔STRING_NAME/NODE_PATH, and `nil`→object pass; STRING→number, number→STRING, and `nil`→scalar are rejected as `INVALID_ARGUMENT`
- [x] 4.2 In the archived `openspec/changes/archive/2026-09-11-surface-method-call-errors/design.md`, correct the statement "defaultArguments is nil for every existing binding" with a dated correction note: `DefArgs` is bound with two defaults (all-arguments-defaulted) and exercises the defaults-satisfy-short-call path; the surrounding reasoning remains valid for that case

## 5. Validate

- [x] 5.1 `go vet ./pkg/... ./test/pkg/...` clean and `go test ./pkg/core/` green with the updated arity expectations
- [x] 5.2 `GODOT=/path/to/godot make build` and `make test` green (engine harness does not assert `expected`; `test_call_error_reporting` and the leak gate stay green)
- [x] 5.3 `openspec validate refine-varcall-call-error-reporting --strict` passes
