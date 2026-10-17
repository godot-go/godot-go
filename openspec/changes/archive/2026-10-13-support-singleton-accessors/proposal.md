# Proposal

## Why

Engine singletons (`Engine`, `OS`, `ProjectSettings`, `RenderingServer`, …)
can only be reached today via `builtin.GetSingleton(name)` plus a manual
`New<Class>WithGodotOwnerObject` conversion; the one typed helper that exists,
`GetInputSingleton()`, is handwritten. `extension_api.json` already declares
all 41 singletons and the parser already ingests them, so typed accessors are
mechanical one-line generation per entry (roadmap item D3).

## What Changes

- New template `cmd/generate/gdclassimpl/singletons.go.tmpl` producing
  `pkg/gdclassimpl/singletons.gen.go`: one accessor per `singletons` entry —
  `func Get<Name>Singleton() <Type>`.
- Cached semantics: the first call resolves and pins one wrapper; every later
  call returns that same wrapper (godot-cpp `get_singleton()` parity, steady
  state is one branch + load).
- A per-singleton `sync.Once` makes concurrent callers safe; the engine
  resolution runs exactly once per singleton per process.
- A missing singleton (engine returns NULL — e.g. `EditorInterface` in a
  runtime build) yields a nil interface, and the absence is never cached.
- **BREAKING** (name-preserving): the handwritten `GetInputSingleton()` in
  `pkg/gdclassimpl/lib.go` is deleted; the generated accessor replaces it with
  the same name and signature, so existing call sites compile unchanged.

## Capabilities

### New Capabilities

- `typed-singleton-accessors`: every engine singleton declared in the
  extension API is reachable from Go as a typed, identity-stable, nil-safe
  accessor generated from the API metadata.

### Modified Capabilities

None. Removing the handwritten helper is an implementation detail subsumed by
the new capability; no existing spec-level behavior changes.

## Impact

- `cmd/generate/gdclassimpl/`: new embedded template, `GenerateSingletons`
  wired into `Generate()`, small coverage in `generate_test.go`.
- `pkg/gdclassimpl/`: new `singletons.gen.go`; handwritten helper removed from
  `lib.go`.
- `test/pkg/example.go`, `test/demo/main.gd`: new demo assertions (identity,
  runtime singletons non-nil, editor singleton nil, existing `Input` usage).
- `docs/overview.md` gains a "Singleton Accessors" section; `docs/roadmap.md`
  marks D3 DONE.
- No parser or FFI changes: `ExtensionApi.Singletons` already parses.

## Non-goals

- No caching machinery for non-singleton wrappers; the package `Pinner`
  pile-up in `New*WithGodotOwnerObject` stays a separate observation.
- No typed access for autoloads or `/root` — engine singletons only.
- No general goroutine-safety for engine calls; `sync.Once` guards slot
  initialization, not engine thread affinity.
- No filtering of editor/platform-specific singletons: all 41 are generated,
  missing ones simply return nil.
- No cache-reset or deregistration API.
