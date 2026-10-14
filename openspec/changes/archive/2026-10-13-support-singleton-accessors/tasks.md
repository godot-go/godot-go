# Tasks

## 1. Codegen pipeline

- [x] 1.1 Add a `DeriveSingletonAccessors(eapi)` helper in
      `cmd/generate/gdclassimpl` returning one descriptor per `Singletons`
      entry (accessor function name from `Name`, engine lookup string from
      `Name`, Go type + constructor from `Type`), with a unit test in
      `generate_test.go` following the `DeriveCategoryA` pattern (fixture
      with N entries yields N descriptors; name/type used distinctly);
      verify `go test ./cmd/generate/gdclassimpl/` passes
- [x] 1.2 Add `singletons.go.tmpl` + `//go:embed` + `GenerateSingletons()`
      writing `pkg/gdclassimpl/singletons.gen.go`, wired into `Generate()`;
      the generated per-entry body implements the D3/D5 cache:
      `atomic.Pointer[holder]` fast path, mutex-guarded cold path,
      `NULL` results return a nil interface without storing; run
      `make generate` and verify the file contains 41 accessors
      (`grep -c '^func Get.*Singleton()'`), `gofmt -l` is clean, and
      `go build ./...` succeeds

## 2. Handwritten-helper replacement

- [x] 2.1 Delete the handwritten `GetInputSingleton()` from
      `pkg/gdclassimpl/lib.go` (the generated accessor replaces it, same name
      and signature); verify `go build ./...` succeeds — `test/pkg/example.go`
      compiles untouched — and `grep -rn "func GetInputSingleton"` matches
      only the generated file

## 3. Demo fixture and driver assertions

- [x] 3.1 In `test/pkg/example.go`, add a Go-side `TestSingletonHelpers()`
      bound as `test_singleton_helpers`, returning a failure count asserted
      as 0 like the other `Test*` fixtures: identity across two
      `GetEngineSingleton()` calls,
      non-nil `Engine`/`OS`/`ProjectSettings`/`Input`, repeated nil for
      `EditorInterface` in the runtime build (second call still nil —
      absence retried, not frozen), and 8 concurrent
      `GetTranslationServerSingleton()` callers (untouched singleton, so they
      all race the cold path) all receiving identical wrappers; the
      never-store-on-NULL later-visibility rule (D5) is verified by review of
      the generated cold path; verify `go build ./...` succeeds
- [x] 3.2 In `test/demo/main.gd`, call `example.test_singleton_helpers()`
      and `assert_equal(result, 0)`; verify
      `GODOT=/home/pcting/bin/godot make test` reports the new assertions as
      passes and the run stays green

## 4. Docs and final gate

- [x] 4.1 Add a "Singleton Accessors" section to `docs/overview.md`
      documenting generation coverage, wrapper identity, and the nil contract
      for missing singletons; verify by reading the section against the
      spec's three requirements
- [x] 4.2 Mark D3 DONE in `docs/roadmap.md` (gap entry, Phase 0 order table
      row, master gantt bar), matching the B1/C3 DONE styling; verify the
      three places render consistently
- [x] 4.3 Run the full gate: `GODOT=/home/pcting/bin/godot make test` green
      (banner, positive counts, zero failures, shutdown marker, no leaks) and
      `openspec validate "support-singleton-accessors" --strict` passes
