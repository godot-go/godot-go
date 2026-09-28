# Tasks

**Decisions taken at 1.1:** new name is `VariantViewFromConstPtr`; hard rename, no deprecated alias.

## 1. Settle the naming contract

- [x] 1.1 Confirm `VariantViewFromConstPtr` as the new name, or record the alternative chosen; decide explicitly whether the old name survives as a deprecated forwarding alias or is removed outright. This is the one decision with downstream consequences, so make it here rather than mid-implementation. Verify by recording the decision in this file before touching code.
  - `VariantViewFromConstPtr`. The dropped `New` prefix is the mechanism: nothing new is created, so a name that reads as a constructor would be lying. Hard rename with no forwarding alias -- keeping the old name alive would preserve the exact confusable token the change exists to remove.
- [x] 1.2 Grep for any use of the old name outside the five known call sites -- generated code, docs, examples -- so the blast radius is known before the rename rather than discovered by the compiler. Verify the count matches expectations or the extra sites are listed.
  - Blast radius was exactly as predicted: 1 definition + 5 production call sites. **No generated-code references** -- the templates in `cmd/generate` never emit the helper, so the rename needs no regeneration. No mention in the main specs under `openspec/specs/`. The only other hits are archived change records and this change's own proposal/design, which are historical and deliberately left untouched.

## 2. Rename and document

- [x] 2.1 Rename the helper in `pkg/builtin/variant.go` and rewrite its doc comment to state, at the definition: the result is a non-owning view; it must never be destroyed; and value Variants (`bool`, `int`, `float`, `Vector2` and friends) are the only case where a view and a copy behave identically, because there is nothing behind them to alias. Verify by reading the comment without opening the body and being able to answer "may I destroy this?".
  - The comment also calls out why the exception is dangerous rather than merely informative: a view of an `int` survives everything a view of an `Array` does not, so tests that happen to use value Variants prove nothing.
- [x] 2.2 State in the same comment why the byte loop is deliberate rather than lazy: the varcall reject path depends on marshalling allocating nothing, so an owning copy here would break an invariant elsewhere. Verify the comment answers "why not just call `variant_new_copy`?" before someone tries.
- [x] 2.3 Update all five call sites: `pkg/core/method_bind_callback.go:63`, `pkg/core/method_bind_reflect.go:291`, `:857`, `:1066`, `pkg/core/classdb_callback.go:458`. Verify `go build ./...` is clean.
  - All five updated; `go build ./...` clean.
- [x] 2.4 Where a call site already carries a hand-written non-owning comment (`method_bind_callback.go:57`), keep it if it adds path-specific detail or trim it to a pointer to the helper's doc if it is now redundant. Verify no call site relies on a comment the rename made stale.
  - Kept. It carries path-specific detail the helper's doc cannot: that *this* path's reject branch frees nothing because marshalling allocated nothing. That is a property of the caller, not of the helper.

## 3. Pin the difference

- [x] 3.1 Add a source-scan test asserting no value produced by the view helper reaches a `Destroy()` call anywhere in the package. Verify it passes on the renamed tree, and verify it fails when a `Destroy()` is temporarily added to a view in a scratch copy -- a check that cannot fail is not a check.
  - `pkg/core/variant_view_safety_test.go`. Passing on the clean tree. Falsification was run for real: injecting `args[i].Destroy()` immediately after the marshalling assignment produced
    `method_bind_callback.go: args[i] is a non-owning Variant view and must not be destroyed`,
    and the injection was reverted.
  - Two rounds of false positives shaped the matching. A file-wide name match flagged every unrelated `v.Destroy()` in `method_bind_reflect.go`, because `v` is reused freely. Scoping to the enclosing brace block was still not enough: Go `case` labels do not open a block, so the window ran past the clause that created the view and picked up sibling branches. The check now bounds the window at the closing brace *or* the next `case`/`default`, whichever comes first. Truncating early only narrows what is inspected, so a misjudged boundary costs coverage rather than producing a false accusation.
- [x] 3.2 Add the complementary assertion that the owned sibling (`NewVariantCopy`, or a `NewVariant*` constructor) is what appears wherever a Variant is destroyed after use. Verify by inspecting each destroy site and confirming it reaches an owned value.
  - Done by inspection rather than automation -- the task asks for inspection, and a regex cannot infer which destroys reach owned values. Every destroy reachable from a view-producing function operates on an owned value:
    - `method_bind_callback.go:85` -- `cn := inst.GetClass()`, owned `StringName`.
    - `method_bind_reflect.go:1075,1087` -- `gdsn := StringName{}` and `gds := gdsn.AsString()`, both locally constructed and owned; both also sit in the `default:` clause, outside the view's `case Variant:` window.
    - `classdb_callback.go`, `method_bind_reflect.go:291`, `:857` -- no destroy in the view's scope.
- [x] 3.3 Record in the test's comment why the check is a source scan rather than a runtime refcount assertion: `Array` and `Dictionary` do not expose the underlying reference count, and deliberately destroying a borrowed view to demonstrate the crash is a crash, not a test.
  - Recorded, along with the accepted limit: a view returned from one function, rebound elsewhere, and destroyed there sits outside the window. The direct mistake -- the one actually made -- is caught.

## 4. Verify

- [x] 4.1 Run `make test` and verify the full suite is green with no new aborts or leaks -- the marshalling paths are covered by the existing suite, so a wrong rename shows up here.
  - `PASS: extension loaded, 1103 assertions ran, 0 failures, no leaked engine objects`.
- [x] 4.2 Grep the tree for the old name and verify it appears nowhere except a deprecated alias, if one was kept. Record the grep output in the change notes.
  - **Zero production references.** The only Go hit is `variant_view_safety_test.go`, where the old name appears inside the guard that stops it coming back. Remaining hits are `openspec/changes/archive/**` and this change's own proposal/design -- historical records, deliberately not rewritten.
- [x] 4.3 Re-run the varcall decode tests added in `fix-varcall-arg-error-reporting` and verify they still pass, since that is the suite that surfaced this hazard and it exercises both helpers.
  - Pass. Assertion count is 1103 before and after the rename -- unchanged, so nothing was lost or silently skipped in the move.
