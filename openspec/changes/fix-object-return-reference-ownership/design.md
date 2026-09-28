# Design

## Context

The engine's ptrcall return path for a refcounted return type constructs a `Ref<T>` into the caller-provided return slot: it adds one reference and transfers ownership of that reference to whoever reads the slot. In generated Go (`cmd/generate/gdclassimpl/classes.go.tmpl:206` → `NewRef{{T}}GDExtensionIternalConstructor(&ret)`), the wrapper is built with `builtin.NewRef`, which deliberately changes no refcount and sets no finalizer. The transferred +1 therefore has no owner and is never released. Evidence: 200 `CollisionShape2D.get_shape()` calls through Go produce `Leaked instance: CircleShape2D ... Reference count: 200` plus a leaked-RID error at engine exit — one orphaned reference per call.

Call-site classification (verified in `pkg/gdclassimpl/classes.refs.gen.go` and `cmd/generate/gdclassimpl/classes.refs.go.tmpl`):

| Site | Count | Reference source | Must own? |
|---|---|---|---|
| `NewRefXGDExtensionIternalConstructor` (generated) | 661 | Engine transfers a +1: the ptrcall return slot (`classes.go.tmpl:206`) or an engine-constructed instance handed to the extension (`classes.go.tmpl:75`, `NewXWithGodotOwnerObject`) | **Yes** |
| `NewRefXAsRef` (generated) | 661 | Wraps an existing `RefCounted` interface value whose reference is held by the caller | No — borrowing |
| `NewRef[T](nil)` in `test/pkg` (`example.go:597,621`; `object_arg_tests.go:135`) | 3 | Nil placeholder receiver/argument | Neither |
| `NewRefInit` / copy `Ref()` (`pkg/builtin/ref_generic.go`) | — | `InitRef()` / `Reference()`, finalizer installed | Already correct |

Sibling context: `fix-generated-call-pin-leak` (3fe924d) scoped per-call pinning so return wrappers became collectable. Collectable is not released: with no finalizer, GC reclaims the Go wrapper and leaves the Godot reference orphaned. The pin was never the cause; this change fixes the release half that the pin change exposed.

## Goals / Non-Goals

**Goals:**
- Each engine-transferred return reference is owned by the returned wrapper and released exactly once (finalizer or explicit `Unref`).
- Repeated object returns keep the refcount stable; the engine exit leak check passes for a `get_shape` loop.
- Explicit `Unref()` remains idempotent with the finalizer (no double release).
- Borrowing wrappers (`AsRef`) never release a reference they do not own.

**Non-Goals:**
- Changing `NewRefInit`, copy `Ref()`, or the borrowing contract of `NewRef`/`NewRefXAsRef`.
- Redesigning the Ref ownership model (dispose-only discipline, pools, arenas).
- Touching argument-decode or variant-return encoding (covered by sibling changes).
- Editing `*.gen.*` directly.

## Decisions

**Decision 1: Option (b) — a distinct owning constructor.** Add to `pkg/builtin/ref_generic.go` a constructor (e.g. `NewRefTransfer[T]`) that wraps without changing the refcount — the +1 already exists and was transferred — and installs `runtime.SetFinalizer(r, (*RefBase[T]).Unref)`. The generated `*GDExtensionIternalConstructor` factories switch to it via the template; `NewRef` keeps its borrowing contract for `AsRef` and any future non-owning use.

Rejected — option (a), a finalizer on `NewRef` itself: it changes the contract of all 1,323 call sites at once, including the 661 borrowing `AsRef` sites. A collected borrower would call `Unreference()` on a reference owned elsewhere — a premature free, strictly worse than the current leak (crash/corruption vs. memory drift), and it would silently mis-bind any future hand-written borrowing caller. The generated code already names the two roles differently (`IternalConstructor` vs `AsRef`), so separating them in the runtime costs one function and makes each role correct by construction.

**Decision 2: Finalizer target is the existing `(*RefBase[T]).Unref`.** `Unref` already calls `runtime.SetFinalizer(r, nil)`, unreferences once, and zeroes `m_ref`, so explicit release followed by GC is a no-op. This satisfies the promise already made in `RefBase`'s doc comment ("released when the RefBase becomes unreachable via a finalizer") without introducing new state.

**Decision 3: Template-only propagation.** Edit `cmd/generate/gdclassimpl/classes.refs.go.tmpl` (the `IternalConstructor` factory body only) and run `make generate`. The return lines in `classes.gen.go` are unchanged — they call the factory, whose body changes.

**Decision 4: Verify the `WithGodotOwnerObject` sub-case before flipping.** The factory is shared by the ptrcall return path and `NewXWithGodotOwnerObject` (`classes.go.tmpl:75`). Cross-reference godot-cpp's `_gde_internal_constructor` usage to confirm the engine transfers a held reference in that path too (task 1.2); if it does not, that call site gets init-style handling instead of the transfer constructor.

## Risks / Trade-offs

- **Earlier frees where a leak previously masked them.** Wrappers now release on GC; code that stashed the raw object pointer past the wrapper's lifetime will now see the object freed. That is the intended contract; callers needing a longer lifetime copy via `Ref()` (+1).
- **`WithGodotOwnerObject` semantics.** If the engine does not transfer a reference in that path, using the transfer constructor there would over-release. Mitigated by explicit verification (task 1.2) before the template change lands.
- **Finalizer timing is non-deterministic.** Release happens at GC, not at scope exit; the acceptance gates are the refcount-stability tests and the engine exit leak check, not mid-run determinism. Same trade-off the existing copy/finalizer path already accepts.
- **Finalizers vs engine shutdown.** Finalizers run in the Go process; if the engine tears down first, releases could race. The existing `Ref()` copy path already relies on this ordering and the sibling pin change surfaced no such issue; the exit test covers the new scale.

## Corrections found during implementation

**The classification table was wrong for half of the `IternalConstructor` sites, and Decision 3 was unsafe because of it.** The table recorded `NewRefXGDExtensionIternalConstructor` as uniformly engine-transferred. It is shared by two callers that pass it completely different things:

```go
// classes.go.tmpl:207 — ptrcall return. The engine filled &ret through
// PtrToArg<Ref<T>>::convert, which constructs a Ref and so calls
// reference(). The +1 is ours.
return NewRefShape2DGDExtensionIternalConstructor(&ret)

// classes.go.tmpl:75 — WithGodotOwnerObject. A Go-built struct wrapping an
// object someone else supplied. No engine write, no convert, no +1.
return NewRefXxxGDExtensionIternalConstructor(inst)
```

Decision 3 said the return lines would be unchanged and only the factory body would flip. Flipping the shared factory would have made `WithGodotOwnerObject` release a reference it never acquired — the premature-free risk this change was supposed to avoid, arriving through the factory rather than through rejected option (a). Resolved by adding `NewRefXGDExtensionReturnOwner` for the return line only and leaving `IternalConstructor` borrowing, which does mean the return line changes after all.

**A null return needs the finalizer suppressed.** `Unref` is not a field write; it issues a live `unreference` ptrcall against the wrapper's owner. A `get_shape()` on an empty collision shape hands back a wrapper over no engine object, so an unconditional finalizer would dereference a null owner during collection. `NewRefTransfer` checks `ObjectArgPtr(obj) != nil` before installing. This was not anticipated anywhere in the design.

**Finalizer timing made the first version of the tests flaky, and the flakiness was not in the binding.** `runtime.GC()` does not guarantee the finalizer goroutine has drained, so a single collection can read a count that has not settled: the same build produced `3 -> 3` on one run and `3 -> 45` on the next. Worse, the baseline handle itself became garbage. `held` is dead after its last use under Go's liveness analysis, so its own finalizer fired mid-test and the count dropped by one that had nothing to do with the wrapper under observation — `base=3, dropped=2`, which looks exactly like a double-release in the code under test. Fixed with `defer runtime.KeepAlive(held)` in each test and polling toward the target rather than sampling once.

Worth separating cleanly, because both looked like binding bugs and neither was: the drift above baseline was real scheduling, the drift below baseline was the test releasing its own observer.

**The proposal's call-site count was off by two.** 1,325 call sites, not 1,323; the extra grep hits are the `func NewRef[T ...]` definition itself.
