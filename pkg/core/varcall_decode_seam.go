package core

import (
	"reflect"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

// DecodeVarcallArgs decodes a varcall argument list against the declared Go
// parameter types and returns the values that would be handed to the callee, or
// a *VarcallArgDecodeError naming the argument that failed.
//
// This is the varcall counterpart to DecodePtrcallArgs: a thin, behaviour-free
// seam over the decoder the real varcall callback uses. It exists because the
// failure contract cannot be observed from the GDScript side. A rejected varcall
// surfaces as a GDExtensionCallError, which GDScript has no way to catch -- it
// aborts the calling script instead, so a test written there cannot tell "the
// engine rejected my call" apart from "my test broke". Driving the decoder
// directly lets the three things that matter be asserted on their own: the
// failure comes back as an error rather than a panic, it names the argument
// index that failed, and the owned prefix decoded before the failure is released.
//
// The seam adds no policy of its own. The callback's decision about what to write
// into the engine's error slot is exercised separately.
func DecodeVarcallArgs(receiver GDClass, args []Variant, decls []reflect.Type) ([]reflect.Value, error) {
	return reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs(receiver, args, decls)
}
