# Tasks

## 1. Settle the naming contract

- [ ] 1.1 Confirm `VariantViewFromConstPtr` as the new name, or record the alternative chosen; decide explicitly whether the old name survives as a deprecated forwarding alias or is removed outright. This is the one decision with downstream consequences, so make it here rather than mid-implementation. Verify by recording the decision in this file before touching code.
- [ ] 1.2 Grep for any use of the old name outside the five known call sites -- generated code, docs, examples -- so the blast radius is known before the rename rather than discovered by the compiler. Verify the count matches expectations or the extra sites are listed.

## 2. Rename and document

- [ ] 2.1 Rename the helper in `pkg/builtin/variant.go` and rewrite its doc comment to state, at the definition: the result is a non-owning view; it must never be destroyed; and value Variants (`bool`, `int`, `float`, `Vector2` and friends) are the only case where a view and a copy behave identically, because there is nothing behind them to alias. Verify by reading the comment without opening the body and being able to answer "may I destroy this?".
- [ ] 2.2 State in the same comment why the byte loop is deliberate rather than lazy: the varcall reject path depends on marshalling allocating nothing, so an owning copy here would break an invariant elsewhere. Verify the comment answers "why not just call `variant_new_copy`?" before someone tries.
- [ ] 2.3 Update all five call sites: `pkg/core/method_bind_callback.go:63`, `pkg/core/method_bind_reflect.go:291`, `:857`, `:1066`, `pkg/core/classdb_callback.go:458`. Verify `go build ./...` is clean.
- [ ] 2.4 Where a call site already carries a hand-written non-owning comment (`method_bind_callback.go:57`), keep it if it adds path-specific detail or trim it to a pointer to the helper's doc if it is now redundant. Verify no call site relies on a comment the rename made stale.

## 3. Pin the difference

- [ ] 3.1 Add a source-scan test asserting no value produced by the view helper reaches a `Destroy()` call anywhere in the package. Verify it passes on the renamed tree, and verify it fails when a `Destroy()` is temporarily added to a view in a scratch copy -- a check that cannot fail is not a check.
- [ ] 3.2 Add the complementary assertion that the owned sibling (`NewVariantCopy`, or a `NewVariant*` constructor) is what appears wherever a Variant is destroyed after use. Verify by inspecting each destroy site and confirming it reaches an owned value.
- [ ] 3.3 Record in the test's comment why the check is a source scan rather than a runtime refcount assertion: `Array` and `Dictionary` do not expose the underlying reference count, and deliberately destroying a borrowed view to demonstrate the crash is a crash, not a test.

## 4. Verify

- [ ] 4.1 Run `make test` and verify the full suite is green with no new aborts or leaks -- the marshalling paths are covered by the existing suite, so a wrong rename shows up here.
- [ ] 4.2 Grep the tree for the old name and verify it appears nowhere except a deprecated alias, if one was kept. Record the grep output in the change notes.
- [ ] 4.3 Re-run the varcall decode tests added in `fix-varcall-arg-error-reporting` and verify they still pass, since that is the suite that surfaced this hazard and it exercises both helpers.
