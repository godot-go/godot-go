## 1. Detection and encoding

- [ ] 1.1 In `isPtrcallBorrowEcho` (`pkg/core/method_bind.go`), add `StringName` and `NodePath` to the type switch so echoed borrows of these types are recognized
- [ ] 1.2 In `GDExtensionTypePtrFromReflectValue` (`pkg/core/variant_refect_value.go`), gate the `StringName`/`NodePath` `inst.Destroy()` calls on `destroySource`, matching the container cases
- [ ] 1.3 Verify `go build ./pkg/core/...` compiles

## 2. Tests

- [ ] 2.1 In `test/pkg/example.go`, add echo methods taking a StringName/NodePath argument and returning it unchanged, plus fresh-return methods returning a value with different contents than the argument (per design Decision 2); register via `ClassDBBindMethod`
- [ ] 2.2 In `test/demo/main.gd`, add ptrcall and varcall assertions for echo round-trips (returned value equals input) and fresh returns (expected different value), repeated enough iterations to surface refcount drift
- [ ] 2.3 Sensitivity check: temporarily disable the new detection cases, rebuild, and confirm orphaned-reference/double-free symptoms appear on echo round-trips; revert and confirm clean

## 3. Validate

- [ ] 3.1 Run `go vet ./pkg/core/...` on changed files
- [ ] 3.2 Run `GODOT=/path/to/godot make build` and `make test`; suite green with no orphan StringName/NodePath warnings in output
