# Design

## Context

Generated class methods are emitted by `cmd/generate/gdclassimpl/classes.go.tmpl`. For vararg methods the template declares `var ret <DeclaredType>` and passes `&ret` as the return buffer to `CallFunc_GDExtensionInterfaceObjectMethodBindCall`. Non-vararg methods use the declared type directly, which is correct for ptrcall. Vararg methods use varcall, and the engine always writes a `Variant` into the return buffer. See proposal.md - Why for the motivation.

Audit of `godot_headers/extension_api.json` (current API):

| Source | Vararg methods | Declared return | Buffer today | Status |
|---|---|---|---|---|
| Class methods | 27 | 17 void, 7 `Variant`, 3 `Error` (`emit_signal`, `rpc`, `rpc_id`) | `Error` for the 3 | **Broken** |
| Built-in class methods (`Callable`, `Signal`) | 6 | `Variant` (`call`), `Callable` (`bind`), void (4) | Typed `CallBuiltinMethodPtrRet[T]` | Buffer sized to `T` (ptrcall, not varcall), see below |
| Utility functions | 12 | `Variant` (`max`, `min`), `String` (`str`), void (9) | Typed `ret` via ptrcall | Buffer sized to `T` |

Built-in and utility vararg calls go through ptrcall with a typed buffer, so the engine writes a value of the declared size. The audit confirms they are not affected by the varcall issue. The `Variant`-returning class methods already use a `Variant` buffer. Void methods pass a null return pointer.

## Goals / Non-Goals

**Goals:**
- Every vararg class-method call uses a `Variant` return buffer and converts to the declared type.
- The generator enforces that rule, so a new vararg method cannot silently regress.
- Record the audit result so the built-in and utility paths are documented as checked.

**Non-Goals:**
- Changing ptrcall paths for built-in and utility functions. They are already sized correctly.
- Changing vararg argument marshalling.
- Changing ownership of `Variant` results. That is handled separately (see proposal.md - Non-goals).

## Decisions

**Decision 1: Always use a `Variant` buffer for vararg class calls.** The engine writes a `Variant` for every vararg call, so the buffer must always be `Variant`. Alternative considered: enlarge the buffer per declared type. Rejected because it encodes an engine detail per type and would break silently if the engine changed.

**Decision 2: Convert from `Variant` only for non-`Variant` declared returns.** `Error` is converted with `Error(v.ToInt64())`. `Variant` returns are returned as-is. Void methods skip conversion. The `Variant` is destroyed after conversion for the `Error` case, since the engine transferred it and nothing else holds it. Alternative considered: a generic `FromVariant[T]` helper. Deferred: only one non-`Variant` vararg return type exists today, so a per-type branch in the template is simpler. The generator guard (Decision 3) makes a second type an explicit change.

**Decision 3: Generator guard for unsupported vararg return types.** The generator fails when a vararg method's return type has no `Variant` conversion. Supported types today are `Variant`, `Error`, and void. Alternative considered: silently fall back to `Variant` and convert with a default. Rejected because it would hide a type that needs a real conversion.

**Decision 4: Regression test in `test/demo`.** Reproduce the issue's shape: a loop of signal emits interleaved with a 1 KB string return, run with garbage collection active, checking every returned string. The issue reports the failure as roughly 1 in 15 headless runs, so the test must loop enough iterations to make the failure visible, and the fixed build must pass repeatedly.

## Risks / Trade-offs

- [Risk] The bug is probabilistic and may not reproduce in a single run → Mitigation: loop several hundred iterations, run the test several times on the fixed build, and check the `GOGC=off` control case.
- [Risk] `ToInt64` on an `Error`-returning `Variant` may not match the engine's encoding → Mitigation: verify against godot-cpp (`../godot-cpp`) and cover the value in the regression test.
- [Risk] The audit could miss a method that only appears in a newer engine API → Mitigation: the audit is a generator-time check, so a regenerated API is re-checked by the same guard.

## Migration Plan

Regenerate `pkg/gdclassimpl/classes.gen.go` with `make generate` after the template change. No public Go signatures change, so callers need no migration. Rollback is reverting the template and regenerated file together.

## Open Questions

None.
