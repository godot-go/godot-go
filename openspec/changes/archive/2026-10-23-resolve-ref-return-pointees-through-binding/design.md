# Design

## Context

The prior change (`resolve-object-returns-through-binding`, PR #190) introduced
`resolveGoStructObjectReturn[T]` in `pkg/gdclassimpl/object_return.go` and wired
it into the non-refcounted return branch of `classes.go.tmpl`. The refcounted
return branch was deliberately excluded (design D3: "the `Ref*` value owns the
+1 the return slot transfers"). This change lifts that exclusion for the
*pointee* inside the `Ref`, not the `Ref` interface type itself.

Current refcounted return template output (per site):
```go
return NewRefShape2DGDExtensionReturnOwner(&ret)
```

After the change:
```go
return NewRefShape2DGDExtensionReturnOwner(resolveGoStructObjectReturn[Shape2D](&ret))
```

## Goals / Non-Goals

**Goals:**
- `Ptr()` on a returned `Ref` yields the runtime-class wrapper when one resolves.
- The ownership invariant is untouched: the `Ref` still owns exactly one +1,
  released by `RefBase.Unref()` through `m_ref.Unreference()`.
- No new exported symbols, no new helper functions.

**Non-Goals:**
- Changing the `Ref[T]` type parameter (invariance wall).
- Providing an exported `AsRef[Sub]()` conversion.
- Modifying `decodeRefArgument` (the argument path already does binding-first).

## Decisions

### D1: Reuse `resolveGoStructObjectReturn` for the ref branch

The existing helper is type-parameterized on `T builtin.Object` and returns
`T`. When `T = Shape2D`, the helper receives `*Shape2DImpl`, attempts binding
lookup, and returns either the binding wrapper (e.g. `*CircleShape2DImpl` which
satisfies `Shape2D`) or the original `&ret` fallback.

The `NewRef{Declared}GDExtensionReturnOwner` factory takes the declared
interface and builds `RefTransfer[T]` over it. It does not inspect the concrete
type — only the owner pointer (`ObjectArgPtr`). So swapping which wrapper
satisfies the interface has zero effect on ownership bookkeeping.

**Alternative considered**: a separate `resolveRefPointee` helper. Rejected
because the resolution logic is identical and a second symbol adds no clarity.

### D2: Binding creation does not touch the refcount

`NewGDExtensionClassFrom{Class}Owner` (e.g. classes.gen.go:102246) constructs
`&CircleShape2DImpl{Owner: owner}` and pins it. No `Reference()`, no
`InitRef()`. The binding create callback calls this constructor only.

**Evidence**: repeated calls to `GetShape()` on the same `CollisionShape2D`
with the same `CircleShape2D` already pass `test_return_refcount_stability`
(200 iterations, refcount settles).

### D3: `RefBase.Unref()` is wrapper-agnostic

```go
func (r *RefBase[T]) Unref() {
    if r.m_ref != zero {
        if r.m_ref.Unreference() {
            CallFunc_GDExtensionInterfaceObjectDestroy(ObjectArgPtr(r.m_ref))
        }
        r.m_ref = zero
    }
}
```

`Unreference()` is a promoted method from `ObjectImpl` — a ptrcall against the
owner pointer. `ObjectArgPtr` extracts the owner pointer from any wrapper. The
result is the same whether `m_ref` holds `*Shape2DImpl` or `*CircleShape2DImpl`.

### D4: Template edit location

In `cmd/generate/gdclassimpl/classes.go.tmpl`, the refcounted return branch
(currently lines 218–224 region) wraps `&ret` directly. The edit inserts the
resolver call:

```
-  return NewRef{{ $fnReturnType }}GDExtensionReturnOwner(&ret)
+  return NewRef{{ $fnReturnType }}GDExtensionReturnOwner(resolveGoStructObjectReturn[{{ $fnReturnType }}](&ret))
```

The explicit type argument `[{{ $fnReturnType }}]` is required because Go type
inference would bind `T` to `*Shape2DImpl` from `&ret` and always fall through.

## Risks / Trade-offs

- **Perf**: one extra `object_get_instance_binding` call per refcounted return
  (574 sites). Same cost as the non-refcounted change (202 sites); binding
  lookup is a cheap map hit after first resolution. Accepted.
- **Retention**: the binding holds a `cgo.Handle` to the wrapper, anchoring it
  for the object's lifetime. This is already true for argument decode. The
  binding is deleted at object-free by the engine, so no permanent retention.
- **`Ptr()` identity guarantee**: callers who relied on `Ptr()` producing a
  fresh pointer per call (to distinguish calls) would break. No existing test
  or doc makes this claim; the guarantee is strictly stronger (canonical
  identity matching the argument path).
- **Null path unchanged**: `resolveGoStructObjectReturn` checks for nil owner
  before issuing any binding call. `test_null_return_schedules_no_release`
  remains valid.
