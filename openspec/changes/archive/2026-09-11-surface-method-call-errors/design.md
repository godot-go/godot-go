## Context

Today the varcall callback (`GoCallback_MethodBindMethodCall`) discards its `r_error` slot, and both the arity shortfall (`method_bind.go:306`) and type-conversion failure (`method_bind_reflect.go:42`) raise `log.Panic`. A panic inside an exported cgo function aborts the process, so a caller-induced mismatch is fatal. The bind metadata already carries everything needed to classify a mismatch: declared per-argument variant types (`gdeArgumentTypes`), trailing default arguments (`gdeDefaultArgumentPtrs`/`DefaultArguments`), and `IsVariadic`. See proposal.md for motivation; the behavior contract is in `specs/method-call-error-reporting/spec.md`.

## Goals / Non-Goals

**Goals:**
- Report caller-induced arity/type mismatches through `GDExtensionCallError` and skip the method body, keeping the process alive.
- Do it without introducing Variant/refcount leaks on the reject path (the `make test` leak gate must stay green).
- Leave accepted-call behavior byte-for-byte unchanged.

**Non-Goals:**
- ptrcall error reporting (no error slot in the typedef).
- Surfacing user Go-method runtime errors as call errors.
- `METHOD_NOT_CONST` handling.

## Decisions

### D1 — Validate at the callback boundary, not inside `GoMethodMetadata.Call`

`Call` returns a `Variant` and has no error channel, while the `r_error` slot is only in scope in the callback. Validating in the callback keeps the ABI-facing decision ("did this call satisfy the signature?") at the ABI boundary, matching godot-cpp's `call_with_variant_args` ordering (arity, then per-argument type). Alternative considered: thread `r_error` down into `Call` — rejected as it spreads an ABI type through internal reflection helpers that are reused elsewhere.

### D2 — Pre-conversion strict type check, not convert-then-rollback

Discovery during implementation: `convertVariantToGoTypeReflectValue` *coerces* (`arg.ToInt64()`, `arg.ToGoString()`, …) and returns an error only for a `nil` type — it never rejects a wrong-typed Variant, so wrapping it to "return an error instead of panic" catches nothing. The coercing converter also means there is nothing to roll back on a type rejection, because a *check* performs no conversion.

The faithful mechanism mirrors godot-cpp `VariantCasterAndValidate` (`binder_common.hpp:148`), which validates each argument with the engine predicate `variant_can_convert_strict(actualType, expectedType)` *before* casting. That predicate is already bound here (`CallFunc_GDExtensionInterfaceVariantCanConvertStrict`) and the supplied type is available via `Variant.GetType()`. So validation is a pure predicate pass over the already-marshalled argument `Variant`s — no `reflect` staging, no owned-container double-allocation, no rollback machinery.

The reject path releases **nothing**, and must not. Marshalling (`NewVariantCopyWithGDExtensionConstVariantPtr`, `variant.go:58`) is a raw bit-copy of the engine's borrowed argument, not `variant_new_copy`, so the resulting `Variant` views do not own a reference — calling `Destroy()` on them would free engine-owned memory (which is why the success path never destroys them). Because validation is a predicate that performs zero owned allocation, a rejection leaves nothing to clean up: the reject path is leak-free by construction. This supersedes the earlier "single-pass conversion with rollback" plan, which was built on a wrong assumption about where type errors surface and about the ownership of marshalled arguments.

### D3 — Arity is default- and variadic-aware

Reject `TOO_MANY_ARGUMENTS` when supplied > declared, unless `IsVariadic`. Reject `TOO_FEW_ARGUMENTS` when the call cannot be satisfied even after defaults; the required-count is derived from the *same* fill condition `Call` uses (`method_bind.go:300-312`), which is `max(supplied, defArgsCount) < declared`, so arity validation and default application cannot disagree. The `expected` field reports the count the caller must supply (`declared` for too-many; the required minimum for too-few). `defaultArguments` is `nil` for every existing binding, so `defArgsCount` is 0 and the required count reduces to `declared`; a method whose every argument is defaulted (required 0) exercises the defaults-satisfy-short-call path under the current leading-index fill model. Variadic methods skip both bounds (the declared count is the single variadic slice slot).

> **Correction (2026-09-20, `refine-varcall-call-error-reporting`):** The statement above that "`defaultArguments` is `nil` for every existing binding" is inaccurate. `DefArgs` is bound with two defaults (all-arguments-defaulted), and it is that binding which exercises the defaults-satisfy-short-call path. The surrounding reasoning remains valid for the all-defaulted case.

### D4 — `expected` for a type mismatch is the bound variant type

A parameter is accepted iff `variant_can_convert_strict(supplied.GetType(), gdeArgumentTypes[i])`. On the first failure, set `argument` to the failing zero-based index and `expected` to `gdeArgumentTypes[i]` (a `GDExtensionVariantType`), mirroring godot-cpp's `expected = argtype`. Because strict conversion allows Godot's numeric/string inter-conversions, a genuinely non-convertible pairing (e.g. `Array`/`Object` into a scalar or vector parameter) is what yields `INVALID_ARGUMENT` — the same set godot-cpp rejects. Class-specific `Ref`/`Object` sub-type validation (godot-cpp's `VariantObjectClassChecker`) is out of scope: the existing `obj.(Ref)` cast is left as-is and a wrong-class object argument remains a converter fault, not a reported call error.

### D5 — Nil return on reject; panics reserved for internal faults

On any rejection the callback writes a nil `Variant` into `r_return` (mirroring godot-cpp assigning a default `Variant` even when `call` errors) and returns. `log.Panic` is kept only for null method user data and null instance — conditions a caller cannot induce.

## Risks / Trade-offs

- **[Leak on reject path]** A rejection must not strand or wrongly free marshalled arguments. → Mitigation: argument views are non-owning bit-copies and validation performs zero owned allocation, so the reject path frees nothing and must not call `Destroy()` on the views (that would free engine-owned memory). `make test` leak gate plus a wrong-type-with-container test guard this.
- **[Accepted-call drift]** Adding a strict-conversion check could reject a call that previously coerced-and-succeeded. → Mitigation: `variant_can_convert_strict` is the engine's own predicate (godot-cpp parity); Godot-valid conversions (numeric/string inter-conversion, nil→object) still pass. The existing suite (1058 asserts) is the regression net.
- **[Residual converter panics]** Wrong-class `Object`/`Ref` arguments and unsupported bound Go types still hit `log.Panic` inside the converter. → Mitigation: documented as out of scope (D4); only caller-inducible count mismatches and non-convertible variant types are reported. A wrong-class object argument is rare and remains a converter fault.
- **[Default-slot mapping ambiguity]** The current `Call` applies defaults by low slot index, which does not match trailing-default semantics; the arity required-count is derived from the same `max(supplied, defArgsCount) < declared` fill condition `Call` uses, so validation and default application cannot disagree for any binding. No existing binding supplies defaults. If a latent default-mapping bug is exercised later, it is out of scope but must be logged.
- **[Ownership of `r_return`]** Writing nil into the engine's return slot must not later be double-freed → Mitigation: follow the existing return-encoding ownership convention already used on the success path.

## Open Questions

- Should the debug-only gating godot-cpp uses (`#ifdef DEBUG_ENABLED` around arity and strict-conversion checks) be mirrored, or should godot-go always validate? Default: always validate — the predicate is cheap and a mismatch is otherwise a silent wrong-value or abort.
