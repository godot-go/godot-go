## 1. Detection and encoding

- [ ] 1.1 In `isPtrcallBorrowEcho` (`pkg/core/method_bind.go`), add `StringName` and `NodePath` to the type switch so echoed borrows of these types are recognized
- [ ] 1.2 In `GDExtensionTypePtrFromReflectValue` (`pkg/core/variant_refect_value.go`), gate the `StringName`/`NodePath` `inst.Destroy()` calls on `destroySource`, matching the container cases
- [ ] 1.3 Add a `destroySource bool` parameter to `GDExtensionVariantPtrFromReflectValue` (mirroring its type-ptr counterpart), gate the `StringName`/`NodePath` destroys on it, and update both call sites in `GoMethodMetadata.Call`: the standard branch computes the flag via `isPtrcallBorrowEcho(ret[0], args)`, the variadic branch always passes true (it never holds borrows)
- [ ] 1.4 Remove the stale post-call StringName/NodePath arg-destroy loop from `GoMethodMetadata.Call` (design Decision 3 — it double-releases borrow-decoded args)
- [ ] 1.5 Verify `go build ./pkg/core/...` compiles

## 2. Tests

- [ ] 2.1 In `test/pkg/example.go`, add echo methods taking a StringName/NodePath argument and returning it unchanged, plus fresh-return methods returning a value with different contents than the argument (per design Decision 2); register via `ClassDBBindMethod`
- [ ] 2.2 In `test/demo/main.gd`, add ptrcall and varcall assertions for echo round-trips (returned value equals input) and fresh returns (expected different value), repeated enough iterations to surface refcount drift
- [ ] 2.3 Sensitivity check A: temporarily disable the new detection cases, rebuild, and confirm orphaned-reference/double-free symptoms appear on echo round-trips in both call styles; revert and confirm clean
- [ ] 2.4 Sensitivity check B: temporarily re-add a destroy of borrowed varcall args (equivalent to the deleted loop), rebuild, and confirm drift/symptoms on repeated varcall echo calls; revert and confirm clean

## 3. Validate

- [ ] 3.1 Run `go vet ./pkg/core/...` on changed files
- [ ] 3.2 Run `GODOT=/path/to/godot make build` and `make test`; suite green with no orphan StringName/NodePath warnings in output
