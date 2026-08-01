## 1. Fix GDVIRTUAL dispatch for unimplemented virtuals

- [x] 1.1 In `pkg/core/classdb_callback.go`, update `GoCallback_ClassCreationInfoGetVirtualCallWithData` to return `nil` when the Go class does not implement the requested virtual:
  - resolve the class name from `pUserdata`
  - look up `Internal.GDRegisteredGDClasses.Get(className)`
  - return `nil` when `ci.VirtualMethodMap[methodName]` is absent
  - return `pUserdata` otherwise
- [x] 1.2 Verify with `make build` + `make test` that `Example.get_maximum_size()` returns `(-1, -1)` and `set_size` takes effect

## 2. Remove the hash-collision generator workaround

- [x] 2.1 Remove `hasHashCollision` from `cmd/generate/gdclassimpl/templatefunctions.go`
- [x] 2.2 Remove the `{{ if hasHashCollision ... }}` branch from `cmd/generate/gdclassimpl/classes.go.tmpl`
- [x] 2.3 Remove the `"hasHashCollision": hasHashCollision` funcMap entry from `cmd/generate/gdclassimpl/generate.go`
- [x] 2.4 Run `make generate` and confirm `Control` methods (`SetSize`, `GetSize`, `SetPosition`, `GetPosition`, etc.) use the method-bind path

## 3. Restore test coverage

- [x] 3.1 `test/pkg/example.go`: re-enable `e.SetSize(size, true)`; remove hash-collision comments and the "expected to be wrong" debug text
- [x] 3.2 `test/demo/main.gd`: restore `assert_equal(example.get_size(), Vector2(100, 200))`

## 4. Remove incorrect documentation

- [x] 4.1 Delete `docs/hash-collisions.md`
- [x] 4.2 Revert the hash-collision section in `README.md`

## 5. Verify

- [x] 5.1 Run `make build`
- [x] 5.2 Run `make test` — all assertions pass, including the restored size assertion
