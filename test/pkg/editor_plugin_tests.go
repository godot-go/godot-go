package pkg

// Fixtures for the editor plugin harness (openspec change
// support-editor-plugins). ExampleEditorPlugin is registered with ClassDB
// when the demo extension's editor-level init runs and added through
// AddEditorPlugin; the engine instantiates the class itself when the editor
// adds it to the tree, and the bound _enter_tree virtual dispatches into Go
// to print the gate marker. At editor shutdown the core framework's
// automatic unload removes the plugin, and the terminator unregisters the
// class so no registration names outlive the extension.
//
// Engine-side instantiation builds instances with reflect.New on the
// registered struct -- the constructor function passed to ClassDBRegisterClass
// serves the decode path, not engine creation -- so the virtual below is the
// observable instantiation signal (see the change's verification notes 1.2).

import (
	"fmt"
	"strings"
	"sync/atomic"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	"github.com/godot-go/godot-go/pkg/gdclassimpl"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// ExampleEditorPlugin is the fixture plugin class descending from the
// engine's EditorPlugin.
type ExampleEditorPlugin struct {
	gdclassimpl.EditorPluginImpl
}

func (p *ExampleEditorPlugin) GetClassName() string {
	return "ExampleEditorPlugin"
}

func (p *ExampleEditorPlugin) GetParentClassName() string {
	return "EditorPlugin"
}

// V_ExampleEditorPlugin_EnterTree fires when the editor adds the plugin to
// the tree, proving the engine instantiated the Go class. WARN level keeps
// the marker visible under the gate run's LOG_LEVEL=WARN, matching the
// main-loop shutdown marker convention.
func (p *ExampleEditorPlugin) V_ExampleEditorPlugin_EnterTree() {
	log.Warn("GODOTGO_EDITOR_PLUGIN_ENTERED")
	probeEditorCallableBridge(p)
}

// probeEditorCallableBridge composes the C1 Go-Callable bridge with the
// editor plugin surface (openspec change support-editor-plugins, task 4.4):
// the plugin registers a tool-menu item whose callback is a Go func wrapped
// by NewGoCallable, then invokes that same Callable through the engine
// channel. The headless editor builds no GUI, so nothing would ever click
// the menu item itself; the direct call proves the callable crossed the
// boundary into the editor's menu registration and still answers from Go.
func probeEditorCallableBridge(p *ExampleEditorPlugin) {
	var invocations int64
	callable := NewGoCallable(func(args []Variant) (Variant, error) {
		atomic.AddInt64(&invocations, 1)
		return NewVariantInt64(7), nil
	})
	defer callable.Destroy()

	title := NewStringWithUtf8Chars("GoMenuItem")
	p.AddToolMenuItem(title, callable)
	title.Destroy()

	args := NewArray()
	defer args.Destroy()
	ret := callable.Callv(args)
	ret.Destroy()

	if invocations == 1 {
		log.Warn("GODOTGO_EDITOR_C1_CALLABLE_PASS")
	} else {
		log.Warn("GODOTGO_EDITOR_C1_CALLABLE_FAIL", zap.Int64("invocations", invocations))
	}
}

func NewExampleEditorPluginFromOwnerObject(owner *GodotObject) GDClass {
	obj := &ExampleEditorPlugin{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

// RegisterEditorPluginFixture registers the fixture class at the editor init
// level and adds it as an editor plugin. The engine instantiates it when the
// EditorNode drains the queued add.
//
// TOOLS-enabled engine builds dispatch the EDITOR init level even when
// running the project as a game (main.cpp gates initialize_extensions(EDITOR)
// on TOOLS_ENABLED, not on the editor session), and the plugin API guard
// correctly refuses there -- so the fixture engages only inside an actual
// editor session, the pattern extensions must follow.
func RegisterEditorPluginFixture() {
	if !gdclassimpl.GetEngineSingleton().IsEditorHint() {
		return
	}
	editorFixtureActive = true
	ClassDBRegisterClass(NewExampleEditorPluginFromOwnerObject, nil, nil, func(t *ExampleEditorPlugin) {
		ClassDBBindMethodVirtual(t, "V_ExampleEditorPlugin_EnterTree", "_enter_tree", nil, nil)
	})
	AddEditorPlugin(&ExampleEditorPlugin{})
}

// editorFixtureActive pins whether the editor-level registration happened, so
// the terminator stays symmetric when the EDITOR level initializes during a
// TOOLS-build game run without the fixture engaging.
var editorFixtureActive bool

// UnregisterEditorPluginFixture releases the fixture class registration at
// editor deinitialization, mirroring UnregisterExampleTypes for the scene
// level. It runs as the user terminator; the framework's automatic plugin
// unload removes the plugin itself immediately afterwards.
func UnregisterEditorPluginFixture() {
	if !editorFixtureActive {
		return
	}
	editorFixtureActive = false
	ClassDBUnregisterClass[*ExampleEditorPlugin]()
}

// EditorPluginRejectedOutsideEditor drives the plugin API from runtime code
// and reports the verdict (1 == pass): add and remove must refuse with a
// recoverable panic at the editor-build guard, and the refused add must not
// have touched the registry (the second add must hit the same guard, never
// the duplicate-registration message). These probes log PANIC lines as part
// of doing their job; the announcement below marks them as expected.
func (e *Example) EditorPluginRejectedOutsideEditor() int64 {
	fmt.Printf("  editor_plugin_guard: about to call the plugin API outside an editor build; three recoverable PANIC logs are expected from this probe\n")

	msg := capturePanic(func() { AddEditorPlugin(&ExampleEditorPlugin{}) })
	if msg == "" {
		fmt.Printf("  editor_plugin_guard: FAIL: add outside an editor build did not panic\n")
		return -1
	}
	if !strings.Contains(msg, "editor build") {
		fmt.Printf("  editor_plugin_guard: FAIL: add panicked for the wrong reason: %s\n", firstLine(msg))
		return -2
	}
	msg2 := capturePanic(func() { AddEditorPlugin(&ExampleEditorPlugin{}) })
	if msg2 == "" || !strings.Contains(msg2, "editor build") {
		fmt.Printf("  editor_plugin_guard: FAIL: refused add mutated the registry: %s\n", firstLine(msg2))
		return -3
	}
	msg3 := capturePanic(func() { RemoveEditorPlugin(&ExampleEditorPlugin{}) })
	if msg3 == "" || !strings.Contains(msg3, "editor build") {
		fmt.Printf("  editor_plugin_guard: FAIL: remove outside an editor build not refused at the guard: %s\n", firstLine(msg3))
		return -4
	}
	fmt.Printf("  editor_plugin_guard: PASS - add and remove refused at the editor-build guard, registry untouched\n")
	return 1
}
