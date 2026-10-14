# Design

## Context

`extension_api.json` carries a top-level `singletons` array of `{name, type}`
entries (41 at the current API version; every entry's `type` has a generated
class and a `New<Type>WithGodotOwnerObject()` constructor). The parser already
exposes it as `ExtensionApi.Singletons` (`cmd/extensionapiparser/model.go`),
but no consumer exists. The sole typed helper, `GetInputSingleton()` in
`pkg/gdclassimpl/lib.go`, is handwritten and stateless. See proposal.md - Why.

## Goals / Non-Goals

Design-level boundaries beyond the proposal:

- The generation must be pure metadata templating: a `derive` step testable
  without running codegen against the full API.
- Wrapper construction must reuse the existing `New<Type>WithGodotOwnerObject`
  constructors unchanged — no special-casing of their pinning semantics here.
- The steady-state accessor body must stay within Go's inline budget; cold
  construction lives in a separate function.

## Decisions

### D1: Dedicated template and output file

New `cmd/generate/gdclassimpl/singletons.go.tmpl` rendered by
`GenerateSingletons()` into `pkg/gdclassimpl/singletons.gen.go`, wired into
`Generate()` beside `GenerateClasses/GenerateClassRefs/GenerateVirtuals`.
Alternative: per-class blocks inside `classes.gen.go` — rejected: a ~200-line
reviewable file beats diff-noise in a 167k-line generated file, and the
singletons list is a flat cross-class concern.

### D2: Generate all entries, no filtering

Every `singletons` entry gets an accessor, including `EditorInterface`,
`JavaClassWrapper`, `JavaScriptBridge`, `GDScriptLanguageProtocol`. The JSON
list is the parity contract; filtering would encode editorial judgments the
reference implementation does not make. Absent singletons are handled at call
time by D5, not at generation time.

### D3: Per-singleton mutex + atomic slot (not `sync.Once`)

Generated code pairs, per entry, a `sync.Mutex` and an
`atomic.Pointer[holder]` where `holder` wraps the interface value. The
accessor fast path is one atomic load of the slot plus a branch — a
single-word pointer load, so no torn two-word interface can ever be observed,
and it stays within the inline budget. The cold path locks, double-checks the
slot, and resolves the engine lookup exactly once under the lock.
Alternatives: plain `if cached == nil` — a data race on a two-word interface;
`sync.Once` — one-shot by construction, so it cannot express D5's
"absence must be retried" without abusing internals (rejected in favor of
mutex + atomic during design, superseding the earlier explore-mode lean);
central `sync.Map` table — global mutable state and a slower hot path for no
codegen saving.

### D4: Cold builder separated from the accessor

The accessor body is slot-load + return (hit) or cold-path call (miss); the
engine lookup (`GetSingleton(name)` + `New<Type>WithGodotOwnerObject`) lives
in a separate `build<Type>Singleton()` helper so the accessor stays
inlinable. `name` comes from the entry's `Name`, the constructor and return
type from `Type` (equal for all current entries, but the fields are used
distinctly so future divergence is safe).

### D5: NULL resolves to a nil interface, never cached

If `GetSingleton(name)` returns NULL, the cold path stores nothing and the
accessor returns a nil interface — the faithful translation of godot-cpp
returning `nullptr`, checkable with `if GetEditorInterfaceSingleton() == nil`
instead of a delayed crash inside a method call. A later call re-enters the
cold path and returns non-nil once the engine provides the singleton
(`Engine.register_singleton` makes this reachable in practice). The
double-checked mutex guarantees at most one engine lookup in flight, and the
atomic slot makes successful resolution a once-per-process event.

### D6: Delete the handwritten helper; name-preserving

`GetInputSingleton()` is deleted from `lib.go`; the generated accessor has the
same name and same `Input` interface return, so `test/pkg/example.go`'s only
call site compiles untouched. Alternative: keep handwritten and exclude Input
from generation — rejected: special cases are exactly what D3 (roadmap) exists
to remove.

### D7: Generator unit test mirrors the `DeriveCategoryA` pattern

`cmd/generate/gdclassimpl` gains a small test deriving the accessor set from
a fixture API (N entries in, N accessor descriptors out, name/type mapping
used distinctly), without invoking codegen against the full API — following
the existing `generate_test.go` precedent.

## Risks / Trade-offs

- [Double-checked locking with "don't cache absence" is subtle] → A wrong
  implementation either races (two-word interface), freezes absence, or
  resolves twice; pinned by three demo checks — missing singleton returns
  nil, later-registered singleton resolves non-nil, concurrent callers get
  identical wrappers.
- [Caching pins 41 wrappers into the package `Pinner` forever] → Bounded,
  intended (singletons live for the process), and strictly better than the
  per-call pinning of stateless helpers; noted as fixed cost.
- [Editor-only accessors can mislead users into expecting non-nil] → Document
  the nil contract in `docs/overview.md`; demo asserts the nil case.
- [API regeneration on older extension_api.json without `singletons`] → Field
  parses to an empty slice; generator emits a valid file with zero accessors,
  behavior unchanged.

## Migration Plan

Generated file is additive; the only removal is a name-preserving helper. No
user-facing call changes. Rollback = revert the commit (regenerated file
disappears, handwritten helper returns).
