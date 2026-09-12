package core

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// varcallArity classifies how a varcall's supplied argument count compares to
// the bound method signature. It is the internal mirror of the engine's
// GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS / TOO_FEW_ARGUMENTS outcomes.
type varcallArity uint8

const (
	varcallArityOK varcallArity = iota
	varcallArityTooMany
	varcallArityTooFew
)

// classifyVarcallArity reports whether the supplied argument count satisfies
// the bound signature and, when it does not, the argument count the caller must
// supply (the value written to GDExtensionCallError.expected).
//
// The too-few decision reproduces the exact slot-fill condition used by
// GoMethodMetadata.Call, which fills slot i iff i < supplied || i < defaults.
// Therefore an unsatisfiable slot exists iff max(supplied, defaults) < declared,
// so arity validation and default application can never disagree. Variadic
// methods skip both bounds: the declared count is the single variadic slice slot.
func (md *GoMethodMetadata) classifyVarcallArity(supplied int) (varcallArity, int) {
	if md.IsVariadic {
		return varcallArityOK, 0
	}
	declared := len(md.gdeArgumentTypes)
	defArgs := len(md.gdeDefaultArgumentPtrs)
	if supplied > declared {
		return varcallArityTooMany, declared
	}
	fill := supplied
	if defArgs > fill {
		fill = defArgs
	}
	if fill < declared {
		required := declared - defArgs
		if required < 0 {
			required = 0
		}
		return varcallArityTooFew, required
	}
	return varcallArityOK, 0
}

// varcallArgumentConvertible reports whether the supplied marshalled argument
// is strictly convertible to the bound parameter's variant type at index, using
// the same engine predicate as godot-cpp's VariantCasterAndValidate
// (variant_can_convert_strict). Indices outside the declared range are treated
// as convertible (variadic excess is handled by the arity pass and CallSlice).
//
// The argument is a non-owning bit-copy view of the engine's borrowed value;
// this only reads its type and never frees it.
func (md *GoMethodMetadata) varcallArgumentConvertible(arg Variant, index int) bool {
	if index < 0 || index >= len(md.gdeArgumentTypes) {
		return true
	}
	from := arg.GetType()
	to := md.gdeArgumentTypes[index]
	return CallFunc_GDExtensionInterfaceVariantCanConvertStrict(from, to) != 0
}
