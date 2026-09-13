## Why

The varcall method-bind callback receives a `GDExtensionCallError *r_error` slot and discards it (`method_bind_callback.go:29`). On a bad call it does not report to the engine — it calls `log.Panic` for too-few arguments (`method_bind.go:306`) and for per-argument type-conversion failures (`method_bind_reflect.go:42`), while too-many-arguments is silently ignored. A panic raised inside an exported cgo callback cannot unwind through the C stack, so it aborts the entire process. The net effect is that a single mistyped call from GDScript or the editor — e.g. passing a `String` where an `int` is bound — hard-kills the editor instead of surfacing a diagnostic call error.

## What Changes

- The varcall callback validates the call **before** invoking the bound Go method, matching the godot-cpp reference path (`binder_common.hpp:call_with_variant_args`): arity first, then per-argument type conversion.
- On a mismatch the callback writes `{error, argument, expected}` into `r_error`, leaves `r_return` initialized to nil, and returns **without** invoking the method — the engine then reports a normal call error and stays alive.
- `GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS`, `TOO_FEW_ARGUMENTS`, and `INVALID_ARGUMENT` are produced with the same `argument`/`expected` semantics as godot-cpp. The "too few" check accounts for godot-go's trailing default arguments (required = total − defaults), not the blanket arity godot-cpp uses.
- Argument conversion becomes error-returning rather than panicking, so a conversion failure maps to `INVALID_ARGUMENT`.
- `log.Panic` is reserved for genuinely impossible internal faults (null method user data, null instance), not caller-triggerable mismatches.
- **BREAKING**: calls that previously aborted the process now return a call error to the engine; callers relying on the old crash-on-mismatch behavior would observe different behavior (no such test exists today).

## Capabilities

### New Capabilities
- `method-call-error-reporting`: The varcall bind reports argument-count and argument-type mismatches to the engine through `GDExtensionCallError` instead of panicking, and never invokes the method body on a rejected call.

### Modified Capabilities
<!-- None: no existing capability's requirements change; this adds error-reporting behavior to the varcall path. -->

## Impact

- `pkg/core/method_bind_callback.go` — varcall callback: validate arity + types, write `r_error`, early-return on rejection.
- `pkg/core/method_bind_reflect.go` / `method_bind.go` — conversion and arg assembly return errors instead of `log.Panic`; default-arg aware arity check.
- `test/demo` + `test/pkg` — add cases that call with wrong arity/type and assert the engine reports the error while the process survives.
- Reference: godot-cpp `binder_common.hpp`; gdext equivalent call-error handling for parity.
- Not touched: the ptrcall path (`GDExtensionClassMethodPtrcall` has no `r_error` parameter).

## Non-goals

- No error reporting on the ptrcall path — the GDExtension ptrcall typedef carries no error slot.
- No new user-facing API for returning errors from bound Go methods (that remains a separate design question).
- No change to registration-time validation (`ClassDBBindMethod`) or to `GDEXTENSION_CALL_ERROR_METHOD_NOT_CONST` handling.
- No change to how successful return values are encoded.
