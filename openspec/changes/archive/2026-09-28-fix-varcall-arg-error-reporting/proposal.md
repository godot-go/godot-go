# Proposal

## Why

A varcall whose object argument cannot be decoded into the bound Go parameter type aborts the whole process. `reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs` calls `log.Panic` on the first conversion error and nothing recovers — there is no `recover()` in the production varcall path — so the panic runs off the cgo boundary and godot dies with exit 134.

The varcall callback (`method_bind_callback.go`) already rejects arity and variant-kind mismatches through the `r_error` slot without invoking the bound method, and `method-call-error-reporting` requires any caller-induced mismatch be reported rather than fatal. The gap: **argument decoding** happens later, deep inside `GoMethodMetadata.Call`, with no error channel — so its only failure mode could ever be `log.Panic`.

Reachable today, verified against the engine's strict-convert table (`core/variant/variant.cpp:560`):

| Caller passes | Bound parameter | Strict convert | Go decode |
|---|---|---|---|
| a `Label` | `Sprite2D` | passes (OBJECT→OBJECT) | fails — `unsupported object class` (`method_bind_reflect.go:249`) → `log.Panic` |
| an object with an unresolvable instance binding | any object parameter | passes | fails — `ObjectFromVariant` error (`:229`) → `log.Panic` |

Scalar coercions are **not** triggers: `variant_can_convert_strict` rejects `String`→`int` (`//STRING` is commented out of `INT`'s valid sources), so those never reach the decoder. The reachable set is object-typed arguments, where OBJECT→OBJECT always clears the pre-pass and only the Go side can tell whether the class satisfies the declaration.

## What Changes

- Make the varcall decode stage fail without panicking: every decode failure carries the argument index and a reason.
- Thread it through `GoMethodMetadata.Call` to the varcall callback, which writes `INVALID_ARGUMENT` with the offending index into the engine's `GDExtensionCallError` slot, leaves a nil return, and returns without invoking the method — reusing `rejectVarcallInvalidArgument`.
- Release owned container arguments already decoded before the failure, so a mid-marshal reject does not leak.
- Keep authoring errors fatal: a missing `Ref` constructor or unsupported Go kind is not caller-inducible, per the spec's rule.

## Capabilities

### New Capabilities

- `varcall-argument-decode`: how a varcall argument that passes variant-level validation but fails to decode into the bound Go parameter type is rejected — typed failure carrying argument index and reason, engine-visible `INVALID_ARGUMENT`, nil return, no method invocation, process alive, no leaked owned arguments.

### Modified Capabilities

None. `method-call-error-reporting`'s requirements stay as written and become more completely satisfied; observable behavior is unchanged, so no delta.

## Impact

- `pkg/core/method_bind_reflect.go` — caller-induced decode cases return errors; partial-marshal unwind.
- `pkg/core/method_bind.go` — `CallWithError` added; `Call`'s exported signature preserved.
- `pkg/core/method_bind_callback.go` — decode-stage reject wired to `rejectVarcallInvalidArgument`.
- `test/pkg/`, `test/demo/main.gd` — rejection tests through a real varcall.
- No `*.gen.*` or template changes.

## Non-goals

- **No `recover()` around `bind.Call`** — method-body panics stay fatal.
- **Ptrcall decode failures** — the ptrcall ABI has no call-error slot, so a bounded ptrcall failure is a Go-level outcome the engine cannot see; deferred to its own change.
- No retry or best-effort value-fallback conversion: a failure is a failure.
- No hand-editing of `*.gen.*` files.
- No change to already-landed out-of-band rejection machinery (arity, variant-kind).
