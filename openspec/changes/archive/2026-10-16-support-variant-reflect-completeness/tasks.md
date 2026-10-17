# Tasks

## 1. Acquire ownership primitive (builtin + codegen)

- [x] 1.1 Add `NewRefAcquire[T]` to `pkg/builtin/ref_generic.go` per design D4: `Reference()` on non-null object (`ObjectArgPtr` check), finalizer `Unref`, zero wrapper without finalizer for null; doc comment contrasting the three ratified constructors and the godot-cpp `PtrToArg<Ref<T>>::convert` parity claim. Verify: `go build ./...`.
- [x] 1.2 Create `pkg/builtin/ref_generic_test.go` with a fake `RefCounted` counting `Reference`/`Unreference`: acquire bumps exactly once, wrapper drop (explicit `Unref` path) releases once, null object takes neither bump nor finalizer. Verify: `go test ./pkg/builtin/`.
- [x] 1.3 Edit `cmd/generate/gdclassimpl/classes.refs.go.tmpl` to emit `New{{X}}AsRefAcquired(reference RefCounted) Ref` via `NewRefAcquire`, and `cmd/generate/gdclassinit/classes.init.go.tmpl` to register the acquired flavor into `GDClassRefConstructors` (keep `AsRef` emitted and unchanged). Verify: `make generate && go build ./...`, spot-check `classes.refs.gen.go` / `classes.init.gen.go` for `Resource`.

## 2. Bind-time signature acceptance (design D1)

- [x] 2.1 In `pkg/core/variant_reflect_type.go`: interface kinds implementing the `Ref` protocol register as `OBJECT`, special-casing bare `builtin.Ref` (name length makes `tn[3:]` useless); struct and pointer kinds implementing `Ref` return a registration error naming the offending Go type. Extend `variant_reflect_type_test.go` with accept/reject tables. Verify: `go test ./pkg/core/`.
- [x] 2.2 In `pkg/core/object_class_name.go` (~line 49): same bare-`Ref` special case for the class-name probe; extend `object_class_name_test.go`. Verify: `go test ./pkg/core/`.
- [x] 2.3 Add bind-level Go tests (pattern: `TestBindRejectsExcessDefaults` in `method_bind_call_error_test.go`): binding a method with a `RefResourceImpl` parameter, `*RefResourceImpl`, and `*builtin.RefBase[Resource]` fails with the type named and registers nothing; `RefResource` + bare `builtin.Ref` signatures register and resolve to `OBJECT`. Verify: `go test ./pkg/core/`.

## 3. Varcall Ref decode rework (design D2-D4)

- [x] 3.1 Replace the varcall Ref branch (`method_bind_reflect.go:228-246`): `nil` → `reflect.Zero(t)` first; engine-guard helper (new, cached `ClassdbGetClassTag` + `ObjectCastTo`, design D3) rejects non-inheriting/non-refcounted objects with a returned error naming argument, declared class, actual class; bare `builtin.Ref` decodes via `NewRefAcquire[RefCounted]`; `GDClassRefConstructors` miss returns a typed error (no `log.Fatal`); delete the unchecked `obj.(RefCounted)` assert. Verify: `go build ./... && go vet ./pkg/core/`.
- [x] 3.2 Delete the unreachable varcall Pointer/Struct `Ref`-assert branches (~430-473) now that bind-time rejection precedes them; confirm nothing else references the removed helpers. Verify: `go build ./...` and `go test ./pkg/core/` green.
- [x] 3.3 Go-level decode tests (pattern: `TestVarcallArgDecodeErrorCarriesIndexAndParamType`): declared-class keying with a subclass object sliced to the declared type, bare-`Ref` acceptance of any refcounted, nil → zero value, mismatch → error text carries argument index and both class names. Verify: `go test ./pkg/core/`.

## 4. Ptrcall Ref decode rework (design D1-D3)

- [x] 4.1 Delete the dead `gdextension.Ref`-prefixed block (`method_bind_reflect.go` ~1090-1134) and the `log.Panic` fallthroughs at ~1144/~1152; replace with a reachable Ref decode for interface kinds: object pointer from the argument cell, null → zero, D3 guard as a loud typed failure (message names argument index and classes; no naked reflect panic), acquired registry constructor or generic `NewRefAcquire[RefCounted]` for bare `Ref`. Verify: `go build ./...`.
- [x] 4.2 Go-level ptrcall decode tests mirroring 3.3 over synthetic argument cells (class-tag guard is engine-backed: cover its logic with the engine running only in group 6; decode-path unit tests cover nil/zero and dispatch selection). Verify: `go test ./pkg/core/`.

## 5. Ptrcall object decode mirror (design D5)

- [x] 5.1 Extract varcall's object resolution sequence (`method_bind_reflect.go:247-296`: binding getter → `Implements` → `GDNativeConstructors` by runtime class → `Implements` guard) into a helper taking a raw object pointer; rewire varcall through it. Verify: `go test ./pkg/core/` unchanged-green.
- [x] 5.2 Rewire the ptrcall Interface object branch (~618-692) through the same helper; remove the `log.Panic` at ~659 behind the D3 loud-typed-failure contract. Verify: `go build ./... && go test ./pkg/core/`.

## 6. Fixture matrix and completeness audit (design D6)

- [x] 6.1 Extend `test/pkg` fixtures: bind methods exercising `RefResource` parameter echo, bare `builtin.Ref` parameter, `Ref` return echo, nil-tolerant Ref parameter, and a stash-and-read-later Ref parameter with an explicit drop-and-`runtime.GC()` release hook (leak-snapshot discipline per design Risks). Verify: `go build ./...` for the c-shared extension build used by `make build`.
- [x] 6.2 Add `test/demo/main.gd` fixture chains: Ref crossing on varcall and ptrcall with same-object identity (`get_instance_id`) on echo; user-defined `ExampleRef` instance passed to a `RefCounted`-typed parameter through BOTH paths resolves to the same live Go state (no crash); mismatch (`Node` → `RefResource`, unrelated refcounted → `RefResource`) rejected via `expect_rejected_call` with subsequent assertions still executing; stash fixture reads on a later frame and releases before shutdown. Verify: `GODOT=/home/pcting/bin/godot make test` — full suite green banner, zero failures, shutdown marker, no leaked-object reports, only the known allowed interface WARNs.
- [x] 6.3 Synchronous audit fixtures covering all 36 variant types across argument/return × varcall/ptrcall in `test/pkg` + `test/demo/main.gd` (no `await`s); delete the stale TODO at `pkg/core/variant_refect_value.go:80`. Verify: `GODOT=/home/pcting/bin/godot make test` green with the enlarged assertion count.
- [x] 6.4 Reconcile spec-delta scenarios against fixtures: every scenario in `ref-method-boundary` and the modified `ptrcall-object-argument-decode` pinned by at least one fixture or Go test; fill gaps. Verify: scenario-to-test cross-reference list in the PR-less commit message body.

## 7. Close-out

- [x] 7.1 `openspec validate support-variant-reflect-completeness --strict` passes; update `docs/roadmap.md` marking A3 landed and correcting its stale line-number citations.
- [x] 7.2 `gofmt`/`goimports` clean, `revive` per `.revive.toml` clean on touched files, full `GODOT=/home/pcting/bin/godot make test` one final time, commit as `feat: support variant reflect completeness` style per repo conventions.
