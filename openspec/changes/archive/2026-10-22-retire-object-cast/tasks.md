# Tasks

## 1. Remove the entry point

- [x] 1.1 Delete `ObjectCastTo` from `pkg/builtin/wrapped.go` and remove its `packagePinnerAllowlist` entry plus the comment lines describing it in `pkg/gdclassimpl/pin_scoping_test.go`; verify `go build ./...`, `go vet ./pkg/builtin/`, and `go test ./pkg/gdclassimpl/` all pass.
- [x] 1.2 Drop `builtin.ObjectCastTo` from the `object-wrapper-cast` row in `docs/overview.md`; verify `grep -rn "ObjectCastTo" docs/` returns nothing and the row still names the surviving wrapper entry points.

## 2. Rewire the demo callers

- [x] 2.1 Rewrite `Example.TestObjectCast` (`test/pkg/example.go:477`) to witness identity, rejection, and inheritance through the surviving API: resolve the receiver's owner with `ObjectFromInstanceBinding` and compare `GetGodotObjectOwner()`, require a type assertion to `Image` to fail instead of yielding a wrapper, and keep the `ObjectInheritsClass` checks for `Control`, `Node`, and `Image` with their existing failure codes; verify `go build ./...` and that the method contains no `ObjectCastTo` call.
- [x] 2.2 Replace the cast in `Example.V_Example_Input` (`test/pkg/example.go:653`) with a type assertion on the decoded ref, and update the comment above the `test_object_cast` assertion in `test/demo/main.gd` so it describes the assertion-based witness; verify `go build ./...` and `grep -rn "ObjectCastTo" pkg/ test/ docs/` returns nothing.

## 3. Engine-side and artifact verification

- [x] 3.1 Run the demo suite with `GODOT=<path> make test` and confirm both `_input: H` (`test/demo/main.gd:554-555`) and `test_object_cast` (`:1655`) pass, with no `event not InputEventKey` warning in the run output; verify from the run log.
- [x] 3.2 Run `go test ./pkg/core/ ./pkg/gdclassimpl/ ./pkg/log/` and `openspec validate retire-object-cast --type change --strict`; verify both report success. (`go test ./...` is not usable as a gate here: `pkg/builtin` needs the loaded engine and two `test/concepts` packages fail vet — both reproduce unchanged at HEAD.)

## Workflow follow-up

- Archive the change once the demo suite is green, so the delta removes the cast requirement; verify the archived `openspec/specs/object-wrapper-cast/spec.md` lists five requirements and `openspec validate object-wrapper-cast --type spec --strict` passes.
