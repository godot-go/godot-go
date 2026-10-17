package core

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/godot-go/godot-go/pkg/ffi"
)

// Engine-free tests for the editor plugin registry seams (openspec change
// support-editor-plugins): guard order, registry mutation only on accepted
// calls, insertion-order drain, and the terminator-before-drain shutdown
// sequence. The engine crossing is replaced with a recorder; nothing here
// requires a running engine.

func newRecordingRegistry() (*editorPluginRegistry, *[]string) {
	var calls []string
	r := &editorPluginRegistry{
		addToEngine:      func(className string) { calls = append(calls, "add:"+className) },
		removeFromEngine: func(className string) { calls = append(calls, "remove:"+className) },
	}
	return r, &calls
}

func registryClasses(r *editorPluginRegistry) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.classes...)
}

func TestEditorPluginRegistryAddAccepted(t *testing.T) {
	r, calls := newRecordingRegistry()
	r.add("MyPlugin", true)
	if got := registryClasses(r); !reflect.DeepEqual(got, []string{"MyPlugin"}) {
		t.Fatalf("registry after add = %v, want [MyPlugin]", got)
	}
	if want := []string{"add:MyPlugin"}; !reflect.DeepEqual(*calls, want) {
		t.Fatalf("engine calls = %v, want %v", *calls, want)
	}
}

func TestEditorPluginRegistryDuplicateAddRejected(t *testing.T) {
	r, calls := newRecordingRegistry()
	r.add("MyPlugin", true)
	msg, fired := recoverFatal(func() { r.add("MyPlugin", true) })
	if !fired {
		t.Fatal("duplicate add must panic")
	}
	if !strings.Contains(msg, "already registered") {
		t.Fatalf("panic message = %q, want it to mention already registered", msg)
	}
	if got := registryClasses(r); !reflect.DeepEqual(got, []string{"MyPlugin"}) {
		t.Fatalf("registry mutated on rejection: %v", got)
	}
	if len(*calls) != 1 {
		t.Fatalf("duplicate add reached the engine: %v", *calls)
	}
}

func TestEditorPluginRegistryUnknownRemoveRejected(t *testing.T) {
	r, calls := newRecordingRegistry()
	msg, fired := recoverFatal(func() { r.remove("NeverAdded", true) })
	if !fired {
		t.Fatal("unknown remove must panic")
	}
	if !strings.Contains(msg, "not registered") {
		t.Fatalf("panic message = %q, want it to mention not registered", msg)
	}
	if len(*calls) != 0 {
		t.Fatalf("unknown remove reached the engine: %v", *calls)
	}
}

func TestEditorPluginRegistryRemoveAccepted(t *testing.T) {
	r, calls := newRecordingRegistry()
	r.add("A", true)
	r.add("B", true)
	r.remove("A", true)
	if got := registryClasses(r); !reflect.DeepEqual(got, []string{"B"}) {
		t.Fatalf("registry after remove = %v, want [B]", got)
	}
	if want := []string{"add:A", "add:B", "remove:A"}; !reflect.DeepEqual(*calls, want) {
		t.Fatalf("engine calls = %v, want %v", *calls, want)
	}
}

func TestEditorPluginRegistryOutsideEditorRejected(t *testing.T) {
	for _, op := range []string{"add", "remove"} {
		t.Run(op, func(t *testing.T) {
			r, calls := newRecordingRegistry()
			r.add("Present", true)
			*calls = (*calls)[:0]
			var msg string
			var fired bool
			if op == "add" {
				msg, fired = recoverFatal(func() { r.add("Ghost", false) })
			} else {
				msg, fired = recoverFatal(func() { r.remove("Present", false) })
			}
			if !fired {
				t.Fatalf("%s outside an editor build must panic", op)
			}
			if !strings.Contains(msg, "editor build") {
				t.Fatalf("panic message = %q, want it to mention the editor build requirement", msg)
			}
			if len(*calls) != 0 {
				t.Fatalf("%s outside an editor build reached the engine: %v", op, *calls)
			}
			if got := registryClasses(r); !reflect.DeepEqual(got, []string{"Present"}) {
				t.Fatalf("registry mutated outside editor: %v", got)
			}
		})
	}
}

func TestEditorPluginRegistryDrainRemovesInInsertionOrder(t *testing.T) {
	r, calls := newRecordingRegistry()
	r.add("First", true)
	r.add("Second", true)
	r.add("Third", true)
	*calls = (*calls)[:0]

	r.drain()

	if want := []string{"remove:First", "remove:Second", "remove:Third"}; !reflect.DeepEqual(*calls, want) {
		t.Fatalf("drain engine calls = %v, want %v", *calls, want)
	}
	if got := registryClasses(r); len(got) != 0 {
		t.Fatalf("registry not empty after drain: %v", got)
	}
	*calls = (*calls)[:0]
	r.drain()
	if len(*calls) != 0 {
		t.Fatalf("second drain must be a no-op, got %v", *calls)
	}
}

func TestDeinitializeBindingLevelDrainsAfterTerminator(t *testing.T) {
	original := *editorPluginRegistryInstance
	originalTerminator := GDExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_EDITOR]
	t.Cleanup(func() {
		*editorPluginRegistryInstance = original
		GDExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_EDITOR] = originalTerminator
	})

	var events []string
	*editorPluginRegistryInstance = editorPluginRegistry{
		addToEngine:      func(className string) { events = append(events, "add:"+className) },
		removeFromEngine: func(className string) { events = append(events, "remove:"+className) },
	}
	GDExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_EDITOR] = func() {
		events = append(events, "terminator")
		// Spec: a user terminator still observes its plugins registered.
		if got := registryClasses(editorPluginRegistryInstance); len(got) != 1 {
			t.Errorf("terminator must see live registration, registry = %v", got)
		}
	}

	editorPluginRegistryInstance.add("LivePlugin", true)
	deinitializeBindingLevel(GDEXTENSION_INITIALIZATION_EDITOR)

	if want := []string{"add:LivePlugin", "terminator", "remove:LivePlugin"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("shutdown event order = %v, want %v", events, want)
	}
	if got := registryClasses(editorPluginRegistryInstance); len(got) != 0 {
		t.Fatalf("registry not empty after editor deinit: %v", got)
	}
}

func TestDeinitializeBindingLevelNonEditorDoesNotDrain(t *testing.T) {
	original := *editorPluginRegistryInstance
	originalTerminator := GDExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_SCENE]
	t.Cleanup(func() {
		*editorPluginRegistryInstance = original
		GDExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_SCENE] = originalTerminator
	})

	var events []string
	*editorPluginRegistryInstance = editorPluginRegistry{
		addToEngine:      func(className string) { events = append(events, "add:"+className) },
		removeFromEngine: func(className string) { events = append(events, "remove:"+className) },
	}

	editorPluginRegistryInstance.add("Survivor", true)
	deinitializeBindingLevel(GDEXTENSION_INITIALIZATION_SCENE)

	if want := []string{"add:Survivor"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("scene deinit must not drain plugins: %v", events)
	}
	if got := registryClasses(editorPluginRegistryInstance); len(got) != 1 {
		t.Fatalf("scene deinit emptied the plugin registry: %v", got)
	}
}
