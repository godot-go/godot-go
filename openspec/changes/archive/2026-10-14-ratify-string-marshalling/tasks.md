# Tasks

## 1. Owned-copy helpers in pkg/builtin

- [x] 1.1 Add `Copy()` methods on `String`, `StringName`, and `NodePath` in
      `pkg/builtin/char_string.go` (or companion file) built on the existing
      `StringNameCopyConstructor`/`NodePathCopyConstructor` pattern, wiring a
      `StringCopyConstructor` over `globalStringMethodBindings.constructor_1`
      if absent; verify with `go build ./...` and `go vet ./pkg/builtin/`.
- [x] 1.2 Write doc comments on each `Copy()` stating the ownership contract
      (independent refcount via copy constructor, caller must `Destroy()`,
      valid on borrowed receivers, safe past the call) and the borrow rule
      they replace; verify revive lint clean on the touched files.

## 2. Demo fixtures pinning the ratified table

- [x] 2.1 Add bound fixture methods in `test/pkg/example.go` that echo
      `StringName`, `NodePath`, and an int variant arriving at a native
      `string` parameter as a failure-count `int64` (batch of crossings,
      following the `== 0` convention), called from `test/demo/main.gd` in
      both direct and generic (`call`) styles; verify with
      `GODOT=/home/pcting/bin/godot make test` showing zero failures.
- [x] 2.2 Add a retention fixture: receive a borrowed `StringName`, call
      `Copy()`, store it on the receiver, re-read it in a later frame and
      `Destroy()` it, reporting failures as `int64 == 0` from GDScript;
      verify the full `make test` run stays green through the shutdown
      marker with no orphaned-reference warnings.
- [x] 2.3 Add content-preservation fixtures for `Copy()` on `String`,
      `StringName`, and `NodePath` (copy text equals source, copy destroyed
      after use); verify with `make test` leak and orphan gates green.

## 3. Ratification bookkeeping

- [x] 3.1 Replace the `TODO` at `pkg/core/method_bind_reflect.go:199` with a
      comment summarizing the ratified table and pointing at the
      `string-boundary-marshalling` spec; verify `go build ./...` and that
      the TODO text no longer appears in that file.
- [x] 3.2 Mark roadmap item A2 **DONE** in `docs/roadmap.md` (naming change
      `ratify-string-marshalling`) and check its Phase 1 `Done` checkbox;
      verify by re-reading the A2 section and Phase 1 table.

## 4. Final gate

- [x] 4.1 Run `GODOT=/home/pcting/bin/godot make test` and confirm the full
      gate set: assertion banner with increased counts, zero failures,
      `GODOTGO_MAIN_LOOP_SHUTDOWN` marker, no leaked engine objects or
      orphaned string warnings (ignoring only the three known
      `AGENTS.md` interface-function warnings); run
      `openspec validate ratify-string-marshalling` and confirm it passes.
