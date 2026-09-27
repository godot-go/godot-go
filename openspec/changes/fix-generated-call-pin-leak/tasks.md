# Tasks

## 1. Primary method template (`cmd/generate/gdclassimpl/classes.go.tmpl`)

- [ ] 1.1 In the ptrcall method-body template, add `var pinner runtime.Pinner` + `defer pinner.Unpin()` as the first statements of each generated method body and route the `retPtr`, `cOwner`, `cArgs`, and per-argument-slot pins (lines ~140-191) to `pinner`; verify by rendering `Node.AddChild` and `CollisionShape2D.GetShape` and reading the emitted Go to confirm every `Pin` targets the local pinner and the `defer` is present.
- [ ] 1.2 In the varcall branch of the same template, route the variant temporaries (`vN`), the varargs-loop slot pins, and the `err` slot to the scoped pinner; verify by rendering a variadic method and confirming defer LIFO ordering puts `pinner.Unpin()` last, after every `vN.Destroy()`.

## 2. Remaining generated call-path templates

- [ ] 2.1 `cmd/generate/builtin/builtinclasses.go.tmpl`: scoped pinner in constructors (`ptr`, `args[k]`), `Destroy` (`bx`), and method bodies (`bx`, `args[j]`); verify by rendering a `String` constructor and a `Vector2` method.
- [ ] 2.2 `cmd/generate/builtin/variant.go.tmpl`: scoped pinner in `NewVariantX`, `GDExtensionVariantPtrFromX`, and `ToX` bodies (`ptr`, `encodedPtr`); verify by rendering one reference and one non-reference variant encoding.
- [ ] 2.3 `cmd/generate/gdutilfunc/utilityfunctions.go.tmpl`: scoped pinner for `retPtr`, `args[...]`, and `typePtrArgs` (lines ~59-87); verify by rendering one fixed-arity and one vararg utility function.
- [ ] 2.4 `cmd/generate/ffi/templatefunctions.go` (`cgoPinReturnType`): drop the emitted `pnr.Pin(ret)` global return-value pin per design D4; audit every `CallFunc_*` pointer-returning wrapper to confirm no return value flows into a later cgo call as a Go pointer without a body-level scoped pin (all current ones are engine-memory pointers).

## 3. Hand-written per-call helpers

- [ ] 3.1 `pkg/ffi/builtin_ptrcall.go`: convert each of the six `CallBuiltin*` helpers to a body-scoped `runtime.Pinner` with `defer pinner.Unpin()`; verify each helper's pins (`argsPtr`, `a`, `ptr`, `l`, `r`) target the local pinner.
- [ ] 3.2 `pkg/core/method_bind_callback.go`: convert the per-virtual-call pins (`instPtr`, `inst`, `argPtrSlice[i]`, `rReturn`) to a body-scoped pinner with `defer pinner.Unpin()`, keeping the pins (not deleting them) per design D5; verify the reject paths (`rejectVarcallArity`, `rejectVarcallInvalidArgument`) still return cleanly with the deferred unpin in place.

## 4. Preserve registration-time retention

- [ ] 4.1 Audit that registration-time pins remain on the package pinner and are untouched: `pkg/core/method_bind.go:511-560` (`GoMethodMetadata` return/arg `PropertyInfo` `StringName`/`String` retention), `pkg/builtin/lib.go:51-53` (instance-binding callbacks), `pkg/ffi/lib.go:54` (`FFI.GodotVersion`), `pkg/core/types.go:121` (`ClassInfo`), `pkg/gdclassinit/*`; add a comment at each package `pnr` declaration stating it is for program-lifetime retention only.
- [ ] 4.2 Add a guard test that scans the generated call-path sources (`pkg/gdclassimpl/classes.gen.go`, `pkg/gdutilfunc/utilityfunctions.gen.go`, generated builtin call bodies) and fails if `pnr.Pin` appears inside any generated method, constructor, utility-function, or variant-conversion body, and passes only when each such body declares `defer pinner.Unpin()`.

## 5. Regenerate and build

- [ ] 5.1 Run `make generate` and then `go build ./...`; both must exit 0.
- [ ] 5.2 Review the regenerated diff: changed lines are exclusively the scoped-pinner pattern (local `runtime.Pinner` declaration, `defer pinner.Unpin()`, `pinner.Pin` replacing `pnr.Pin`); confirm no signatures, argument-encoding lines, or other logic moved, and that the ~63k pin sites now target scoped pinners.

## 6. Tests and verification

- [ ] 6.1 Go unit test for collectability: call a generated object-returning method, drop the returned value, force GC, and assert a `weak.Pointer` to the return wrapper reports empty (pre-fix this stays non-empty because the global pin retains it).
- [ ] 6.2 Engine regression in `test/pkg/` + `test/demo/main.gd`: loop `CollisionShape2D.GetShape()` (mirroring the sibling change's repro) and assert the run exits with no `Leaked instance: CircleShape2D` and no leaked `GodotShape2D` RID allocations.
- [ ] 6.3 Metadata-after-unpin check: after the call loop, read back registered method info from Godot (method list / property info names) and assert the `GoMethodMetadata` `StringName`/`String` values are still correct, proving registration pins were not disturbed.
- [ ] 6.4 Run `GODOT=/home/pcting/bin/godot make test` and confirm the suite is green, ignoring the known startup warnings (`get_variant_get_internal_ptr_func`, `callable_custom_get_user_data`, `classdb_register_extension_class_5/6`).
