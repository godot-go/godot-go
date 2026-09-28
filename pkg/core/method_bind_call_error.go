package core

import (
	"fmt"

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
// the bound signature and, when it does not, the argument count reported to the
// caller via GDExtensionCallError.expected.
//
// Both arity errors report the declared argument count, matching godot-cpp's
// call_with_variant_args_dv, which sets expected = sizeof...(P) for too-many and
// too-few alike; the required minimum stays derivable from the registered
// default-argument count.
//
// The too-few decision mirrors the exact slot-fill satisfiability of
// GoMethodMetadata.Call's trailing-default fill, which fills slot i iff
// i < supplied || i >= declared - defaults. An unfilled slot exists iff
// supplied < declared - defaults, i.e. declared - supplied > defaults, so arity
// validation and default application can never disagree. This coupling is
// deliberate: the condition must equal Call's fill satisfiability. Under the
// trailing convention a partial-default binding (0 < defaults < declared) is
// satisfiable as long as the caller supplies the leading (declared - defaults)
// arguments. Moving the fill model away from the trailing convention requires
// changing this check and TestClassifyVarcallArityMatchesCallFill in the same
// commit. Variadic methods skip both bounds: the declared count is the single
// variadic slice slot.
func (md *GoMethodMetadata) classifyVarcallArity(supplied int) (varcallArity, int) {
	if md.IsVariadic {
		return varcallArityOK, 0
	}
	declared := len(md.gdeArgumentTypes)
	defArgs := len(md.gdeDefaultArgumentPtrs)
	if supplied > declared {
		return varcallArityTooMany, declared
	}
	if declared-supplied > defArgs {
		return varcallArityTooFew, declared
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

// VarcallArgDecodeError reports that a varcall argument cleared variant-level
// validation but could not be decoded into the bound Go parameter type.
//
// It exists so the failure can travel from the decoder, which knows the argument
// position and the declared type, up to the varcall callback, which owns the
// engine's GDExtensionCallError slot. The callback reads Index directly instead
// of parsing it back out of a message string.
//
// Only failures a caller can induce are carried by this type. Authoring faults --
// a missing Ref constructor, an unsupported Go kind -- remain fatal and never
// become a VarcallArgDecodeError, per the caller-induced rule in
// method-call-error-reporting.
type VarcallArgDecodeError struct {
	// Index is the zero-based position of the argument that failed to decode.
	Index int
	// ParamType is the declared Go type the argument could not satisfy.
	ParamType string
	// Err is the underlying decoder failure.
	Err error
}

func (e *VarcallArgDecodeError) Error() string {
	return fmt.Sprintf("varcall argument %d failed to decode into %s: %v", e.Index, e.ParamType, e.Err)
}

// Unwrap keeps errors.Is and errors.As working through this wrapper.
func (e *VarcallArgDecodeError) Unwrap() error { return e.Err }
