# Proposal

## Why

`NewVariantCopyWithGDExtensionConstVariantPtr` copies a Variant's bytes with a
plain loop. It never calls `variant_new_copy`, so the result is a bitwise alias of
the source's storage and holds no reference to whatever that Variant points at.
Its name, though, files it beside `NewVariantCopy`, which does call
`variant_new_copy` and returns a genuinely owned value.

Two nearly identical names with opposite ownership semantics. A caller that does
the obvious thing and destroys the result frees memory the source is still using.
This is not hypothetical: writing the varcall decode tests for
`fix-varcall-arg-error-reporting` produced exactly that crash. Building an
`Array` argument with the bit-copy helper and destroying it on the way out took
the harness down with a SIGSEGV; switching to `NewVariantArray`, which uses the
real Variant constructor, fixed it.

## What Changes

- Rename the bit-copy helper so non-ownership is in the name rather than in a
  comment that only one of its five call sites currently carries. The rename is
  the change: a caller choosing between two helpers should not have to read the
  body to learn which one it is safe to destroy.
- Document the safety boundary explicitly. The bit-copy is correct and cheap for
  value Variants (`bool`, `int`, `float`, `Vector2`, and friends) and unsafe to
  destroy for heap-backed Variants (`String`, `Array`, `Dictionary`, object).
- Update all five call sites.
- Add a regression test that pins the ownership difference between the two
  helpers, so a future edit that swaps one for the other fails loudly instead of
  corrupting memory at runtime.

Explicitly not changing: the bit-copy's behaviour. Making it own properly would
put an engine call on the varcall argument marshalling path, and the current
reject-path design depends on marshalling allocating nothing -- see the comment at
`method_bind_callback.go:57`, "the reject path frees nothing". Changing what the
helper does would be the more expensive and more disruptive fix; changing what it
is called costs a rename.

## Capabilities

### New Capabilities
- `variant-copy-ownership`: every helper that produces a Variant from another
  Variant states whether the result owns a reference, and the name makes that
  readable without opening the body.

### Modified Capabilities
- None. No existing requirement changes; the marshalling paths keep behaving as
  they do today.

## Impact

- `pkg/builtin/variant.go` -- the helper's name and documentation.
- Five call sites: `pkg/core/method_bind_callback.go:63`,
  `pkg/core/method_bind_reflect.go:291`, `:857`, `:1066`, and
  `pkg/core/classdb_callback.go:458`.
- No behavioural change intended. The point is to make a contract that is
  currently carried by comments impossible to miss.
- Rename is BREAKING for any downstream Go code calling the old name. The helper
  is `pkg/builtin`-exported, so external extensions may use it; the old name can
  be kept as a deprecated alias for one cycle if that matters more than a clean
  tree.
