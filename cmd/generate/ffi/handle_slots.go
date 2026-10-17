package ffi

import (
	"strings"

	"github.com/godot-go/godot-go/cmd/gdextensionparser/clang"
)

// handleReturningFunctions are the interface functions whose void* result is
// not a memory address but the value stored in an instance-binding (or
// callable-user) slot — a cgo.Handle packed by this extension. Returning
// them through uintptr_t keeps the small integer out of GC pointer slots
// (openspec fix-instance-handle-lifecycle).
var handleReturningFunctions = map[string]bool{
	"GDExtensionInterfaceObjectGetInstanceBinding":  true,
	"GDExtensionInterfaceCallableCustomGetUserData": true,
}

// handleInstanceFunctions are interface functions whose p_instance parameter
// receives the extension instance value written by this extension — a packed
// cgo.Handle, not a pointer to C memory.
var handleInstanceFunctions = map[string]bool{
	"GDExtensionInterfaceObjectSetInstance": true,
}

func isVoidPointer(t *clang.PrimativeType) bool {
	return t != nil && t.IsPointer && t.Name == "void"
}

// markHandleSlots annotates the parsed typedefs so the wrapper templates
// render those slots as uintptr_t on the C side and cgo.Handle on the Go
// side. The annotations ride on the shared *PrimativeType nodes, so the
// Collect* value copies produced afterwards for template execution carry
// them.
func markHandleSlots(ast clang.CHeaderFileAST) {
	for _, e := range ast.Expr {
		fn := e.Function
		if fn == nil {
			continue
		}

		if handleReturningFunctions[fn.Name] && isVoidPointer(&fn.ReturnType) {
			fn.ReturnType.IsHandleSlot = true
		}

		for i := range fn.Arguments {
			t := fn.Arguments[i].Type.Primative
			if t == nil {
				continue
			}
			name := strings.TrimSpace(fn.Arguments[i].Name)
			switch {
			case name == "p_binding" && isVoidPointer(t):
				// Coupled to gdextension_interface.h's naming convention:
				// every void* parameter called p_binding is a cgo.Handle
				// slot.
				t.IsHandleSlot = true
			case handleInstanceFunctions[fn.Name] && name == "p_instance" &&
				t.Name == "GDExtensionClassInstancePtr":
				t.IsHandleSlot = true
			}
		}
	}
}
