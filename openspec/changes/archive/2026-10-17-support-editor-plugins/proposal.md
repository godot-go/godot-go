# Proposal

## Why

The EDITOR init level is registerable and the GDExtension ABI exposes
`editor_add_plugin`/`editor_remove_plugin` — already loaded into `pkg/ffi`
with generated `CallFunc_*` wrappers — but no Go API reaches them, so a Go
extension cannot ask the editor to instantiate an `EditorPlugin` subclass or
unload it at shutdown (roadmap C4). The wrappers are dead code today.

## What Changes

- Add `AddEditorPlugin[T GDClass](inst T)` and `RemoveEditorPlugin[T GDClass](inst T)`
  to `pkg/core`, mirroring godot-cpp `EditorPlugins::add_by_type<T>()` /
  `remove_by_type<T>()`: the instance carries the registered class name, the
  call crosses to `editor_add_plugin`/`editor_remove_plugin`.
- Add a mutex-guarded plugin-class registry with godot-cpp's exact guards:
  duplicate add and unknown remove are rejected; at editor-level
  deinitialization every added plugin is removed automatically, ordered after
  the user-registered editor terminator (mirrors godot-cpp
  `GDExtensionBinding::deinitialize_level`).
- Reject misuse outside editor builds (engine-side calls are silent no-ops
  without `TOOLS_ENABLED`): the guard checks the `EditorInterface` singleton,
  which is nil outside the editor, and panics recoverably per repo misuse
  convention.
- Add an editor-mode verification harness: a headless editor run of the test
  demo (`godot --headless -e --quit-after N`) gated like the existing runtime
  run — driver summary lines, clean shutdown, zero ObjectDB leaks.
- Add a fixture plugin (Go struct embedding the generated `EditorPluginImpl`,
  registered at editor init level). Engine instantiation fires its Go
  constructor, which is the observable success signal. Optional composition
  probe: `AddToolMenuItem` with a C1 Go `Callable`.
- Runtime behavior is unchanged: the same library loads in both modes, the
  runtime gate never touches the plugin path, and no `.gdextension` config
  key changes.

## Capabilities

### New Capabilities
- `editor-plugin-registration`: registering Go classes as editor plugins,
  removal, misuse rejection, and automatic unload at editor shutdown.

### Modified Capabilities

(none)

## Non-goals

- Virtual-method overrides on plugin classes (B3 generated virtual
  interfaces); lifecycle is observed via constructor instantiation and
  explicit calls only.
- Editor UI surfaces (docks, panels) beyond what the registration probe needs.
- `editor = true` packaging, editor help-XML loading, and
  `editor_register_get_classes_used_callback`.
- Interactive (non-headless) editor testing.

## Impact

- `pkg/core`: new editor-plugin registry file; `godot.go` deinit hook.
- `test/pkg` fixture + `test/demo` harness, `Makefile` and
  `test/check_test_output.sh` editor-run stage.
- New spec `editor-plugin-registration`; `docs/overview.md` and
  `docs/roadmap.md` closeout.
