package core

// Editor plugin registration mirrors godot-cpp's EditorPlugins
// (src/classes/editor_plugin_registration.cpp): a registry of plugin class
// names, add/remove guarded against duplicates and unknown entries, and an
// automatic removal of everything added when the editor level deinitializes.
// The engine instantiates the class from the name and owns the instance;
// outside editor builds the engine-side calls are silent no-ops, so the
// registry refuses to pretend otherwise.

import (
	"runtime"
	"sync"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// AddEditorPlugin adds T's registered class as an active editor plugin, the
// Go equivalent of godot-cpp EditorPlugins::add_by_type<T>(). inst is used
// only to resolve the registered class name; the engine instantiates the
// class itself when the plugin is added. The class must already be registered
// with ClassDB, and the call must come from an editor build; outside one the
// engine-side registration would silently do nothing, so the operation
// panics instead. Additions left registered at editor shutdown are removed
// automatically during editor-level deinitialization.
func AddEditorPlugin[T GDClass](inst T) {
	editorPluginRegistryInstance.add(inst.GetClassName(), editorBuildAvailable())
}

// RemoveEditorPlugin removes a class previously added with AddEditorPlugin,
// the Go equivalent of godot-cpp EditorPlugins::remove_by_type<T>(). Removing
// a class that is not currently registered, or calling outside an editor
// build, panics.
func RemoveEditorPlugin[T GDClass](inst T) {
	editorPluginRegistryInstance.remove(inst.GetClassName(), editorBuildAvailable())
}

// editorBuildAvailable reports whether this extension runs inside an editor
// session. The EditorInterface singleton is created by register_editor_types()
// before extension editor-level initialization runs and is unregistered only
// after editor-level deinitialization, so it is present at every legal call
// site of the plugin API (see openspec change support-editor-plugins design
// D3). Runtime builds never provide it -- including TOOLS-enabled builds
// running the project as a game, which still dispatch the EDITOR init level;
// extensions gate their plugin registration on the editor session (for
// example Engine.is_editor_hint) for exactly this reason.
func editorBuildAvailable() bool {
	return GetSingleton("EditorInterface") != nil
}

// editorPluginRegistry is the ordered set of plugin class names added to the
// editor. The engine crossing is injectable so the guard, ordering, and
// drain behavior stay testable without an engine.
type editorPluginRegistry struct {
	mu               sync.Mutex
	classes          []string
	addToEngine      func(className string)
	removeFromEngine func(className string)
}

var editorPluginRegistryInstance = &editorPluginRegistry{
	addToEngine:      callEditorAddPlugin,
	removeFromEngine: callEditorRemovePlugin,
}

// add registers className as an editor plugin. Duplicate adds are rejected
// (godot-cpp ERR_FAIL) before anything crosses to the engine.
func (r *editorPluginRegistry) add(className string, inEditor bool) {
	if !inEditor {
		log.Panic("editor plugin registration requires an editor build",
			zap.String("operation", "add"),
			zap.String("class", className),
		)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.classes {
		if existing == className {
			log.Panic("editor plugin already registered",
				zap.String("class", className),
			)
		}
	}
	// godot-cpp order: record, then cross to the engine.
	r.classes = append(r.classes, className)
	r.addToEngine(className)
}

// remove unregisters a previously added plugin class. Removing a class that
// was never added is rejected before any engine call.
func (r *editorPluginRegistry) remove(className string, inEditor bool) {
	if !inEditor {
		log.Panic("editor plugin registration requires an editor build",
			zap.String("operation", "remove"),
			zap.String("class", className),
		)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, existing := range r.classes {
		if existing == className {
			r.classes = append(r.classes[:i], r.classes[i+1:]...)
			r.removeFromEngine(className)
			return
		}
	}
	log.Panic("editor plugin is not registered",
		zap.String("class", className),
	)
}

// drain removes every remaining plugin class in insertion order and empties
// the registry. It runs from editor-level deinitialization after the user
// terminator (mirroring godot-cpp GDExtensionBinding::deinitialize_level:
// user terminate callback, then EditorPlugins::deinitialize). The engine
// remove call is safe whether or not EditorNode is still alive.
func (r *editorPluginRegistry) drain() {
	r.mu.Lock()
	remaining := make([]string, len(r.classes))
	copy(remaining, r.classes)
	r.classes = r.classes[:0]
	r.mu.Unlock()
	for _, className := range remaining {
		r.removeFromEngine(className)
	}
}

func callEditorAddPlugin(className string) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	sn := NewStringNameWithLatin1Chars(className)
	pinner.Pin(&sn)
	CallFunc_GDExtensionInterfaceEditorAddPlugin(sn.AsGDExtensionConstStringNamePtrPinned(&pinner))
	// The engine copies the queued name; no exit-time orphan StringName.
	sn.Destroy()
}

func callEditorRemovePlugin(className string) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	sn := NewStringNameWithLatin1Chars(className)
	pinner.Pin(&sn)
	CallFunc_GDExtensionInterfaceEditorRemovePlugin(sn.AsGDExtensionConstStringNamePtrPinned(&pinner))
	sn.Destroy()
}
