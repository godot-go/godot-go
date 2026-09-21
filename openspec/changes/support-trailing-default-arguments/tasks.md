# Tasks: support-trailing-default-arguments

## 1. Trailing-convention fill (D1)

- [x] 1.1 In `GoMethodMetadata.Call` (`pkg/core/method_bind.go`), replace the leading-index fill loop with the trailing model: compute `defaultStart := len(md.gdeArgumentTypes) - defArgsCount`; fill slot `i` from `gdArgs[i]` when `i < supplied`, else from `md.DefaultArguments[i-defaultStart]` when `i >= defaultStart`, else leave unfilled. Update the loop comment to state the trailing mapping (`DefaultArguments[k]` → parameter `declared - defaults + k`) and cite godot-cpp `call_with_variant_args_dv`
- [x] 1.2 Confirm the all-defaulted case is unchanged: with `defaults == declared`, `defaultStart == 0` and the fill reduces to today's behavior (spot-check against the `DefArgs` trio)

## 2. Arity gate (D2)

- [x] 2.1 In `classifyVarcallArity` (`pkg/core/method_bind_call_error.go`), change the too-few condition from `max(supplied, defaults) < declared` to `declared - supplied > defaults`; keep the `expected` value as the declared count (owned by the prior refine change). Update the doc comment to describe the trailing satisfiability mirror
- [x] 2.2 Rewrite `TestClassifyVarcallArityMatchesCallFill` so its inline fill model matches D1's trailing condition (`i < supplied || i >= declared-defaults`), keeping the mechanical tie between gate and fill
- [x] 2.3 Update `TestClassifyVarcallArity` boundary cases: `declared=3 defaults=1 supplied=2` now OK (was too-few); `declared=3 defaults=1 supplied=1` still too-few; add `declared=3 defaults=2 supplied=1` OK

## 3. Direct-Go validation (D3)

- [x] 3.1 In `Call`'s fill loop, replace the silent unfilled-slot case with `log.Panic` naming `md.GdMethodName` and the unfilled slot index; update the surrounding comment to note the varcall path pre-validates so this fires only for direct Go callers
- [x] 3.2 Add a unit test that calls `Call` directly with too few args (via a fixture binding) and asserts the panic fires with the method name and slot index

## 4. Bind-time default-count guard (D4)

- [x] 4.1 In the method-metadata constructor (`pkg/core/method_bind.go`), panic when `len(defaultArguments) > argumentCount`, naming the method; place it near the existing `argumentNames > argumentCount` guard
- [x] 4.2 Add a unit test asserting registration with excess defaults panics

## 5. Engine end-to-end (PartialDefArgs)

- [x] 5.1 In `test/pkg/example.go`, add `PartialDefArgs(p_a, p_b int32) int32` (returns `p_a + p_b`) and bind it with a single trailing default `[]Variant{NewVariantInt64(200)}` for `b`
- [x] 5.2 In `test/demo/main.gd`, assert `partial_def_args(5)` succeeds returning `205` (a=5, b=200 default), `partial_def_args(5, 7)` returns `12`, and `partial_def_args()` is rejected as too-few (body does not run)

## 6. Documentation

- [x] 6.1 In `docs/overview.md`, update the default-arguments section to state the trailing convention: bound defaults map to the last N parameters; a caller may omit up to N trailing arguments; excess defaults are rejected at bind time

## 7. Validate

- [x] 7.1 `go vet ./pkg/... ./test/pkg/...` clean and `go test ./pkg/core/` green with the updated gate/fill/guard/validation tests
- [x] 7.2 `GODOT=/path/to/godot make build` and `make test` green (the `PartialDefArgs` trio passes; the pre-existing `to_string` line-27 failure and leak gate behave as before)
- [x] 7.3 `openspec validate support-trailing-default-arguments --strict` passes

## 8. Variadic fill skip (review follow-up, D5)

- [x] 8.1 In `Call` (`pkg/core/method_bind.go`), guard the default-fill loop with `if !md.IsVariadic` so variadic bindings skip fill entirely; update the comment to note the variadic branch passes `gdArgs` directly to `CallSlice` and never reads `callArgs` (fill extracted into `fillCallArgs` so the skip is unit-testable without live engine FFI)
- [x] 8.2 Add a unit test asserting `fillCallArgs` returns nil (no panic) for a variadic binding with zero args even when it carries defaults. The end-to-end empty-slice dispatch and with-arguments no-default-injection cases are covered by the engine demo (`varargs_func` with 0, 1 and 4 args), since `Call`'s return path requires live engine FFI pointers unavailable in bare unit tests
- [x] 8.3 In `test/demo/main.gd`, add a zero-argument `example.varargs_func()` assertion (returns `0`) to pin the engine varcall path
- [x] 8.4 `go vet ./pkg/... ./test/pkg/...` clean and `go test ./pkg/core/` green
- [x] 8.5 In `NewGDExtensionClassMethodInfoFromMethodBind` (`pkg/core/method_bind.go`), register pure-varargs bindings with `argument_count = 0` (no named `argument_info` entries) plus the vararg flag, so GDScript parse-time accepts zero-argument varargs calls (the engine counts each registered argument as required; vararg only relaxes too-many — cf. `GDScript.new`); the varcall callback receives raw args regardless
- [x] 8.6 `GODOT=/path/to/godot make build` and `make test` green with the zero-arg `varargs_func()` demo assertion
