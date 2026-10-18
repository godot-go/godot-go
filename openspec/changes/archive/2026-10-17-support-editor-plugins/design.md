# Design

## Context

See proposal.md for motivation. Engine facts verified against `../godot`:

- `editor_add_plugin`/`editor_remove_plugin` are registered as interface
  functions on **every** build; their bodies are no-ops without
  `TOOLS_ENABLED` (`gdextension_interface.cpp:1682-1695`). So an FFI nil check
  guards nothing.
- Before `EditorNode` exists, add queues the class name and the EditorNode
  constructor drains it (`gdextension.cpp:1097`, `editor_node.cpp:9569`);
  afterwards it routes deferred/immediate to `EditorNode`'s hooks. Removal is
  symmetric. Both paths are safe at any editor stage.
- Editor startup: `register_editor_types()` (which calls
  `EditorInterface::create()`) runs **before** `initialize_extensions(EDITOR)`
  and unregisters **after** its deinit (`main.cpp:797-800, 860-862`). The
  `EditorInterface` singleton therefore exists at every legal call site of
  the plugin API inside an editor.
- The EDITOR init/deinit levels dispatch through
  `GDExtensionBindingInitializeLevel` / `GDExtensionBindingDeinitializeLevel`
  (`pkg/core/godot.go`); `EditorPluginImpl` and native-parent class
  registration already exist.
- `pkg/ffi` already loads `editor_add_plugin`/`editor_remove_plugin` with
  generated `CallFunc_*` wrappers — currently unreferenced.

## Goals / Non-Goals

**Goals:**
- godot-cpp `EditorPlugins` parity at the Go API boundary.
- Editor-mode verification that keeps the runtime gate byte-identical.

**Non-Goals:**
- Validating that the class was registered with ClassDB before add (engine
  reports it; godot-cpp does not check either).
- Editor-lifecycle fidelity beyond registration/removal (no virtual hooks —
  B3), no interactive editor runs.

## Decisions

### D1 — API and registry shape

`pkg/core/editor_plugins.go` (new file):

```go
func AddEditorPlugin[T GDClass](inst T)
func RemoveEditorPlugin[T GDClass](inst T)
```

Instance-carried class name, following the `ClassDBBindStaticMethod[T](inst T, …)`
precedent; alternative `AddEditorPluginByName(string)` rejected for API
consistency. Registry is a mutex-guarded ordered `[]string` mirroring
godot-cpp's `Vector<StringName> plugin_classes` (insertion order preserved for
auto-drain). Per call the class name is converted to a `StringName`, passed to
the FFI wrapper, and `Destroy()`ed immediately — the engine copies; no exit-time
orphan StringNames (lesson from `ensureMethodNameAvailable`).

### D2 — Guard order and misuse reporting

Mirror godot-cpp exactly: duplicate-add and unknown-remove checks run against
the registry **before** any FFI call; rejection is `log.Panic` (the repo's
recoverable misuse convention); registry mutates only on accepted calls.

### D3 — Editor-build guard is the `EditorInterface` singleton

`GetSingleton("EditorInterface") == nil` → recoverable panic in both add and
remove (implemented in `pkg/core`, which cannot import `gdclassimpl` for the
typed helper but shares its underlying `builtin.GetSingleton`). Alternatives
rejected: FFI nil check (symbol exists on all builds, D-context); ClassDB
capability probe (no `class_exists` interface function in the vendored
header). Inside an editor session the singleton exists at every legal call
site (verified above) and the lookup never caches absence. Caveat found in
task 4.1: TOOLS-enabled builds dispatch the EDITOR init level even while
running the project as a game (`main.cpp` gates `initialize_extensions(EDITOR)`
on `TOOLS_ENABLED`, not on the editor session), and there the singleton is
absent — the guard correctly panics; extensions must gate plugin registration
on the editor session (`Engine.is_editor_hint`), as the demo fixture does.

### D4 — Auto-drain in `GDExtensionBindingDeinitializeLevel`

At EDITOR deinit, after the user terminator callback runs, the registry drains
itself in insertion order (each entry through the engine remove call) and is
cleared. This mirrors godot-cpp `GDExtensionBinding::deinitialize_level`:
user terminate callback → `EditorPlugins::deinitialize`. Runtime builds never
run the EDITOR level and the guard keeps the registry empty there, so no
runtime drain is needed.

### D5 — Verification = second gated headless-editor run, spike first

New harness stage runs the same library in the editor:
`godot --headless -e --quit-after N --path test/demo`, with its own gate
fixture-constructors ran, clean shutdown, zero ObjectDB leaks. Fixture:
`ExampleEditorPlugin` (struct embedding `EditorPluginImpl`, class name equal
to struct name per the `ClassDBRegisterClass` rule), registered with the engine
and added from a demo editor-level init callback; engine instantiation fires
the Go constructor, which prints the PASS signal. Optional probe:
`AddToolMenuItem` with a C1 Go `Callable`. Task 1 is a throwaway spike proving
headless editor instantiation + teardown is leak-green before any gate work.
Fallback if `--quit-after` is not honored with `-e`: `--quit` or an explicit
quit call from the fixture constructor.

## Risks / Trade-offs

- **EditorNode teardown vs EDITOR deinit ordering**: the engine may free
  plugins when `EditorNode` dies before our drain runs. Both engine paths are
  then safe (hook pointers nulled; queue erase), but only the spike proves the
  combination stays leak-green under the strict gate.
- **Engine-owned plugin lifetime**: plugin instances are freed by the engine,
  so the Go instance-tracking must release deterministically in `free_func`
  (the C1 pattern); a miss shows up as an ObjectDB or pin-table leak at gate
  time.
- **Headless editor startup cost**: `-e` triggers resource import on the demo
  project; the editor gate run is slower than the runtime run (accepted for CI).
- **Unregistered class passed to add**: engine instantiate failure surfaces as
  engine-side errors, not a Go rejection — deliberately matching godot-cpp.
- Engine-free unit tests must target pure seams (decision functions over the
  registry), following the B2a precedent; FFI crossings are covered only by
  the editor harness.
