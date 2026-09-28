package core

import (
	"reflect"
	"strings"

	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassinit"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// godotClassNameForObjectType resolves the Godot class name that an OBJECT-variant
// Go type should advertise in method metadata.
//
// Godot treats PropertyInfo.class_name as authoritative for OBJECT parameters
// and returns: GDScript static typing validates every call site against it. The
// binder advertised the *owning* class for all of them, so a method bound on
// Example that takes a Node told GDScript the parameter was an Example, and every
// correctly typed caller was rejected at parse time.
//
// Every step is registry-backed rather than string-sniffed: a name is accepted
// only where some registry already knows it.
func godotClassNameForObjectType(t reflect.Type) (string, bool) {
	// A pointer type has no name of its own; the class name lives on the type it
	// points at. Concrete Go class implementations are pointers -- *ExampleRef,
	// for instance -- so without this the lookup would run against an empty
	// string and every such return would look unresolvable.
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	tn := t.Name()

	// Engine classes are keyed by their exact Godot name.
	if _, ok := GDNativeConstructors.Get(tn); ok {
		return tn, true
	}

	// User-defined extension classes registered with this binding.
	if _, ok := Internal.GDRegisteredGDClasses.Get(tn); ok {
		return tn, true
	}

	// Smart pointers advertise the pointee: RefShape2D resolves to Shape2D.
	// Checked after the plain-name lookups so an extension class actually named
	// RefFoo keeps its own name instead of being stripped to "Foo".
	if len(tn) > 3 && strings.HasPrefix(tn, "Ref") {
		if _, ok := GDClassRefConstructors.Get(tn[3:]); ok {
			return tn[3:], true
		}
	}

	// The generic object interfaces carry no class information of their own, so
	// they advertise the base. This also catches a bare Ref, whose prefix strip
	// would otherwise yield an empty class name.
	switch tn {
	case "Object", "Ref", "GDClass", "GDExtensionClass":
		return "Object", true
	}

	return "", false
}

// objectClassForMetadata is the class name a bound method advertises for one
// argument or for its return.
//
// Only OBJECT-variant entries are resolved; every other variant keeps the owning
// class exactly as it was before, so value-type metadata is untouched.
//
// An OBJECT type that matches no registered class advertises "Object" rather than
// failing the bind. A Go type can be a perfectly valid object without being a
// registered Godot class: ExampleRef is a Go-defined RefCounted whose
// registration is disabled, and ReturnEmptyRef hands back a plain Go struct with
// no engine instance behind it at all. Advertising "Object" is truthful, since
// every engine object is an Object, so no legitimate call is rejected. The
// unresolved case is logged rather than silently absorbed.
func objectClassForMetadata(owning string, gdMethodName string, what string, t reflect.Type, variant GDExtensionVariantType) string {
	if variant != GDEXTENSION_VARIANT_TYPE_OBJECT {
		return owning
	}
	if name, ok := godotClassNameForObjectType(t); ok {
		return name
	}
	log.Warn("object-typed "+what+" resolves to no registered godot class; advertising Object",
		zap.String("class", owning),
		zap.String("method", gdMethodName),
		zap.String("go_type", t.String()),
	)
	return "Object"
}
