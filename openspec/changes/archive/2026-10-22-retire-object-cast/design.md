# Design

## Context

See `proposal.md` — Why. The current shape that matters:

- `ObjectCastTo` (`pkg/builtin/wrapped.go:68`) is a composition of
  `classdb_get_class_tag` + `object_cast_to` and `ObjectFromInstanceBinding`. It
  logs on entry and exit, panics on an unknown class name, and returns nil for
  every failure shape.
- The decode paths do not use it. `decodeRefArgument`
  (`pkg/core/method_bind_reflect.go:159`) checks inheritance with
  `ObjectInheritsClass` and can return `"object of class %q is not a %s"`;
  `resolveObjectArgument` (`:215`) resolves the binding and reports the class.
- Generated wrappers embed their ancestor chain
  (`InputEventKeyImpl` → `InputEventWithModifiersImpl` → … → `ObjectImpl`), so a
  wrapper of a subclass already satisfies its parents' interfaces.
- `decodeRefArgument` prefers the binding wrapper but falls back to a wrapper
  built over the *declared* class when no binding can be resolved (`:185-189`).
  That fallback is the one case where a wrapper is sliced below its runtime class.
- Two in-repo callers: the demo witness `Example.TestObjectCast`
  (`test/pkg/example.go:477`) and `Example.V_Example_Input` (`:653`).
- `instance-binding-lookup` already pins wrapper identity
  ("Wrapper identity preserved across calls"), and `object-wrapper-cast` pins the
  inheritance check separately.

## Goals / Non-Goals

**Goals:**

- Remove the cast entry point without dropping any demo coverage that currently
  distinguishes a working cast from a wrong wrapper.
- Leave the spec describing only behaviour that an implementation exists for.

**Non-Goals:**

- No replacement cast API of any shape (generic, error-returning, pointer-taking).
- No change to the slicing behaviour of `decodeRefArgument`, to binding
  resolution, or to the inheritance check.
- No new demo capability or test harness change.

## Decisions

### D1 — Delete the function instead of reshaping it

Alternatives considered:

- **Error-returning form** (`(Object, error)`): matches what decode paths want,
  but they already have it — `decodeRefArgument` names both classes. A second
  error-producing cast in `pkg/builtin` would be a parallel implementation with no
  caller, which is the defect being removed.
- **Generic `CastTo[T]`** deriving the class name from `T`'s identifier (the
  `refArgumentClass` trick, `method_bind_reflect.go:141`): attractive typing, but
  it invents public surface for a need only the demo expresses, and it breaks
  for any user-defined class that has no generated interface. Deferred, not
  rejected for good — recorded as an open question.

Deletion is the only option that lets the requirement be removed rather than
rewritten, which is the point of the change.

### D2 — Assert on the wrapper in hand; do not resolve again first

`decodeRefArgument` has already resolved the binding before the demo method runs,
so `V_Example_Input` becomes `event.(InputEventKey)`. Re-resolving from the engine
pointer inside the handler would perform a second lookup and, worse, would paper
over a wrapper/class disagreement instead of surfacing it as a failed assertion.

This holds for a wrapper that already carries its runtime class, as the decoded ref
does. A wrapper returned by a generated method is typed to the method's declared
class, so narrowing one needs the binding re-resolution described in the migration
text; asserting on it directly reports false for every subclass.

### D3 — The witness asserts the same three facts through the surviving API

`TestObjectCast` keeps its name and its GDScript assertion, and covers:

- **identity** — resolve the engine owner with `ObjectFromInstanceBinding` and
  compare `GetGodotObjectOwner()` with the receiver's owner;
- **rejection** — assert the same wrapper to `Image` (a class `Example` does not
  implement) and require the assertion to fail rather than yield a wrong wrapper;
- **inheritance** — `ObjectInheritsClass` true for `Control` and `Node`, false for
  `Image`, which the method already checks.

### D4 — Removing the requirement loses no unique guarantee

The removed requirement's observable promises map onto surviving specs: identity
to `instance-binding-lookup` ("Wrapper identity preserved across calls"), and the
inheritance half to the requirement that stays in `object-wrapper-cast`. The
nil-safety half is already pinned by the wrapper requirements. Nothing in the
substrate relied on the cast's nil-on-failure convention: the decode paths return
errors, and the demo handled nil with a warning.

### D5 — Keep the one GDScript witness

`test/demo/main.gd:1655` keeps `assert_equal(example.test_object_cast(), 0)`; only
the comment above it changes, because the method still answers a pass/fail code.
Deleting the assertion would remove the only executable evidence for identity and
rejection, which the proposal keeps.

## Risks / Trade-offs

- **A sliced wrapper fails the new assertion where the old cast would have
  upgraded it** — if `decodeRefArgument` had to build a declared-class wrapper
  (no binding available), `event.(InputEventKey)` yields false and the demo's
  `_input` assertion (`main.gd:554-555`) fails. → Mitigation: that assertion is
  exactly the witness; the task list runs the demo suite, and if it fails the
  answer is to fix the slicing in the decode path, not to restore the cast.
- **Public API removal breaks downstream callers** — the project is experimental
  and unversioned, and the proposal marks the removal BREAKING. → Mitigation: the
  spec delta's Migration text names both the old and the new calls.
- **Stale allowlist entry removed with the function** — the entry documents a
  callbacks-pointer pin the implementation no longer performs. → Mitigation:
  `go test ./pkg/gdclassimpl/` after the edit; the test only consults the
  allowlist for functions that pin on the package pinner.
- **The main spec's Purpose is edited before the delta is archived** — the
  capability briefly describes no cast while the requirement text still exists. →
  Mitigation: archive in the same change; do not leave the change open across a
  release.
- **Docs drift** — `docs/overview.md:18` lists `builtin.ObjectCastTo` in the
  capability map. → Mitigation: it is an explicit task, not a follow-up.

## Migration Plan

1. Delete the function and the stale allowlist entry; `go build ./...` and
   `go vet ./pkg/builtin/`.
2. Rewrite the witness and the input handler; the Go-side change compiles without
   the engine, so catch mistakes before launching anything.
3. Run the demo suite (`GODOT=... make test`), which executes both GDScript
   assertions.
4. Archive the change so the delta removes the requirement and keeps the edited
   Purpose.

Rollback is a revert: the function is self-contained, and no other code path was
rewired to depend on its removal.

## Open Questions

- Whether a typed cast API (`CastTo[T]`) is worth adding once a real consumer
  exists. Deferrable: it changes no requirement in this change, and the migration
  text already tells callers what to use instead.
