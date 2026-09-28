package core

import (
	"reflect"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// DecodePtrcallArgs decodes a ptrcall argument list against the declared Go
// parameter types and returns the values that would be handed to the callee.
//
// This is a thin, behaviour-free seam over the decoder used by the real ptrcall
// callback. It exists because that path is only reachable from GDScript through a
// statically-typed call site, and typed object call sites are gated by a separate
// change (fix-object-argument-type-metadata): without a seam, the object-argument
// decode could not be exercised or falsified on its own.
//
// Callers build the argument cells exactly as the engine would. For an object
// parameter the cell is a GDExtensionObjectPtr holding the object pointer, and
// the cell's address is what goes in the slice, matching the outbound encoding in
// cmd/generate/gdclassimpl/classes.go.tmpl.
func DecodePtrcallArgs(receiver GDClass, args []GDExtensionConstTypePtr, decls []reflect.Type) []reflect.Value {
	return reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs(receiver, args, decls)
}
