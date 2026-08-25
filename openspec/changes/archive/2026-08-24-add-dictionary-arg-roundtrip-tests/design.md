## Context

`fix-container-arg-refcount-leak` gave every container argument type varcall owned-copy decode plus post-call release via `destroyOwnedContainerArgs`, and added echo/consume demo tests for `Array` and the nine `Packed*Array` types — but not for `Dictionary`. The PR #140 review flagged this coverage gap: the Dictionary decode/release path and its shared-reference mutation semantics run unexercised.

Relevant ground truth (verified against current code):

- Varcall decode exists (`convertVariantToGoTypeReflectValue`: `arg.ToDictionary()` owned copy) and release machinery includes `Dictionary` (`destroyOwnedContainerArgs`, `pkg/core/method_bind_reflect.go`).
- There is NO `Dictionary` case in the ptrcall decode (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`), so a direct ptrcall with a `Dictionary` parameter is currently unsupported; only varcall dispatch (`.call(...)`) reaches the Go method.
- `Dictionary` return encoding uses `DictionaryEncoder.EncodeTypePtrArg` in ptrcall returns (no copy-constructor transfer) — out of scope here, tracked by the open question already recorded in that change's design.

## Goals / Non-Goals

**Goals:**
- End-to-end GDScript assertions proving Dictionary arguments decode correctly and retain no reference.
- Pin the shared-reference mutation behavior for `Dictionary` (mirrors `test_mutate_array` for `Array`).
- Document actual call-style support (varcall-only for arguments today).

**Non-Goals:**
- Adding a ptrcall decode case for `Dictionary` (separate follow-up if wanted).
- Moving Dictionary's ptrcall return encoding to copy-constructor.
- Changing any lifecycle semantics — this change adds tests and documentation only.

## Decisions

### Decision 1: Cover varcall only, and assert the ptrcall boundary honestly

Echo/consume/mutate assertions use `.call(...)` (varcall), the only supported dispatch for Dictionary arguments. Rather than silently skipping ptrcall, record what direct dispatch does today so a future ptrcall decode case lands with test intent already written. The exact observable (error vs panic vs silent failure) is determined during implementation and asserted accordingly.

**Alternative considered:** adding a ptrcall decode case now to enable symmetric coverage. Rejected — scope creep beyond the flagged gap; the container change deliberately deferred it ("Open Questions").

### Decision 2: Mirror the Array test shapes

Reuse the established patterns: echo method returning the received dictionary, consume method returning a scalar derived from contents (e.g. sum of int values or key count), mutate method adding one key. This keeps the suite uniform across container families and reuses the assertion style already proven in `main.gd`.

## Risks / Trade-offs

- **Shared-reference mutation assert could mask leaks** → a mutating echo would pass even if retention bugs existed elsewhere. Mitigation: keep echo, consume, and mutate as separate methods/assertions like the Array family does.
- **Ptrcall-boundary documentation may drift** if a later change adds the decode case. Mitigation: the spec scenario and design note name the exact mechanism, so the drift surfaces in review.

## Migration Plan

Test-only additions behind no flag; nothing to migrate. Rollback is a revert.

## Open Questions

- None blocking; ptrcall support status is resolved during implementation per Decision 1.
