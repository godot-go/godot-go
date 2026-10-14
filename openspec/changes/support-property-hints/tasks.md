# Tasks

## 1. PropertySpec and registration path

- [ ] 1.1 Define `PropertySpec` in `pkg/core` per design D1 (field docs state
  zero-value semantics: `Usage` zero means `PROPERTY_USAGE_DEFAULT`,
  `ClassName` empty means derive) and verify `go build ./...` and `gofmt`
  pass
- [ ] 1.2 Extract pure helpers `deriveClassName(spec, owningClassName)`
  (design D2: `RESOURCE_TYPE` -> hint string, else qualifier, else owning) and
  `validatePropertySpec` (keeps getter/setter arity rules, adds
  `RESOURCE_TYPE`-with-empty-hint-string panic), with Go unit tests for both
  including all derivation branches and the new panic; verify
  `go test ./pkg/core -run 'PropertySpec|ClassName'` passes
- [ ] 1.3 Implement `ClassDBAddPropertyInfo` using the helpers: core-owned
  `StringName`/`String` temporaries pinned with `pnr` and destroyed via
  `defer` after the engine call (design D3), empty hint string built as an
  empty Godot `String`; verify `go build ./...` passes and the extension
  still loads via `make test`
- [ ] 1.4 Re-implement `ClassDBAddProperty` as a hint-free wrapper delegating
  to `ClassDBAddPropertyInfo` with NONE/empty/DEFAULT; verify existing
  fixtures (`test/pkg/example.go:1113`, `test/pkg/exampleref.go:45`) compile
  unchanged and `make test` stays green

## 2. FFI ownership cleanup

- [ ] 2.1 Replace the `// TODO: create StringName locally here` in
  `pkg/ffi/property_info.go:25` with the ownership contract comment from
  design D5 (pointer-based constructor, caller-owned temporaries; `pkg/core`
  owns creation for property registration); verify the TODO is gone via grep
  and `go build ./...` passes

## 3. Fixtures and harness assertions

- [ ] 3.1 Register new `Example` fixture properties through
  `ClassDBAddPropertyInfo`: float with `RANGE` hint `0:100:1`, int with `ENUM`
  hint `idle,walk,run`, `Resource`-typed property with `RESOURCE_TYPE` hint
  `Texture2D`, an `OBJECT`-typed property with explicit qualifier, and one
  storage-only (`PROPERTY_USAGE_STORAGE`) property, each with working
  getter/setter bound methods; verify the extension builds and loads in
  `make test`
- [ ] 3.2 Add GDScript assertions in `test/demo/main.gd` against
  `ClassDB.class_get_property_list("Example")`: hint, hint_string, usage, and
  class_name values for every new fixture; `class_name` is `Texture2D` for the
  resource-type property and the qualifier for the explicit-qualifier
  property; the legacy hint-free fixture still reports `NONE`/empty/DEFAULT;
  verify `make test` reports the new assertions passing
- [ ] 3.3 Add a GDScript assertion that property metadata reads back intact
  after registration (hint string text uncorrupted on a late property-list
  read), pinning the spec requirement on temporary lifetime; verify via
  `make test`

## 4. Documentation and close-out

- [ ] 4.1 Extend the property section of `docs/overview.md` with a
  `ClassDBAddPropertyInfo` example showing hint/hint string/usage/qualifier
  usage; verify the documented signatures match the implemented ones
- [ ] 4.2 Delete the unplaced-hint-logic entry from `TODO.md` and mark B1
  (and folded-in D2) as done in `docs/roadmap.md`; verify both files reflect
  completion
- [ ] 4.3 Run the full suite `make test` and verify green, ignoring only the
  three known interface-function warnings listed in `AGENTS.md`
