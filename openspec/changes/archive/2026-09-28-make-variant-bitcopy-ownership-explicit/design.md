# Design

## Context

See proposal.md - Why.

The helper has five call sites, and the codebase already understands what it does:
`method_bind_callback.go:57` says "These are non-owning bit-copy views of the
engine's borrowed arguments. Nothing is allocated here, so the reject path frees
nothing and must not call Destroy on them." That comment is exactly the contract
that needs to move into the name. The other four sites carry no such note.

The two helpers in question, side by side in the same file:

```go
func NewVariantCopy(dst, src Variant) {                       // variant_new_copy: owned
func NewVariantCopyWithGDExtensionConstVariantPtr(ptr ...) Variant {  // byte loop: borrowed
```

## Goals / Non-Goals

**Goals:**

- Make the ownership of a returned Variant readable at the call site.
- Record the POD-only safety boundary where a reader will meet it.
- Give a future editor something that fails when the two helpers are swapped.

**Non-Goals:**

- Changing what the bit-copy does. Its allocation-free behaviour is load-bearing
  for the reject-path design.
- Redesigning the owned-copy sibling. `NewVariantCopy(dst, src)` has an awkward
  out-parameter signature, but it is correct and out of scope here.
- Auditing every Variant in the codebase for ownership confusion. This change
  fixes the naming that caused one observed crash; it is not a general sweep.

## Decisions

### Name the view by what it is, not by what it copies

**Chosen: `VariantViewFromConstPtr`.** Dropping the `New` prefix is the point --
nothing new is created, so a name that reads as a constructor would be lying.
"View" carries the non-ownership, and it matches vocabulary Go programmers
already have for a value that shares backing storage.

**Rejected: `NewVariantBitCopyView`.** More explicit about the mechanism, but it
keeps `New`, and `New` is the token that signals "you own this" throughout this
package. A name whose first word contradicts the package's own convention is not
safer than the current one.

**Rejected: keep the name and add a doc comment.** Five call sites, one comment.
The comment sits on the definition, and the mistake happens at the call site.
This is the status quo that produced the crash.

### Document the boundary at the definition, not in a wiki

The doc comment must state that the view is safe to read for any Variant and safe
to destroy for none of them, and that value Variants (`bool`, `int`, `float`,
`Vector2`, and the other inline types) are the only case where a view and a copy
are interchangeable -- because there is nothing behind them to alias.

### Pin the difference with a source-scan check rather than a runtime one

There is no cheap way to observe a Variant's refcount from Go: `Array` and
`Dictionary` do not expose the underlying reference count, and deliberately
destroying a borrowed view to prove it crashes is not a test -- it is a crash.

Instead, a test that scans the package sources and asserts no value produced by
the view helper reaches a `Destroy()` call. Crude, but it fails on exactly the
edit that matters: someone routing a borrowed view into an owning path.

**Rejected: runtime refcount assertion.** Would require exposing engine internals
that the binding deliberately keeps hidden, for a check a grep does at zero cost.

## Risks / Trade-offs

- **[BREAKING rename for downstream extensions] →** Only five internal call sites,
  but the helper is exported from `pkg/builtin`. If downstream breakage matters
  more than a clean tree, keep the old name as a deprecated alias that forwards
  and says so. Decide before implementing, not after someone files an issue.
- **[A source-scan test is brittle] →** It keys on identifiers, not formatting,
  so ordinary reformatting will not break it. It can produce false negatives if
  the value is stashed in a variable and destroyed two hops away; accept that --
  it catches the direct mistake, which is the one that was actually made.
- **[Someone "fixes" the helper by making it own] →** That would add an engine
  call to every varcall argument marshalled and quietly break the reject-path
  invariant that nothing is allocated. The doc comment must say why the byte loop
  is deliberate, not lazy.
- **[Rename touches the hot path files] →** Mechanical, compiler-checked, and the
  existing suite covers the marshalling paths. Low risk, but do not fold it into
  an unrelated commit.
