package core

// #include <godot/gdextension_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"runtime/cgo"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// GoCallback_MethodBindMethodCall is called when GDScript vararg methods calls into Go.
//
//export GoCallback_MethodBindMethodCall
func GoCallback_MethodBindMethodCall(
	methodUserData unsafe.Pointer,
	instPtr C.GDExtensionClassInstancePtr,
	argPtrs *C.GDExtensionVariantPtr,
	argumentCount C.GDExtensionInt,
	rReturn C.GDExtensionVariantPtr,
	rError *C.GDExtensionCallError,
) {
	// A null method handle or a null instance is an internal fault a caller
	// cannot induce, so those remain fatal. Every other mismatch (argument
	// count, argument type) is caller-induced and reported through rError
	// instead of a panic that would abort the process from inside a cgo
	// callback. This mirrors godot-cpp's call_with_variant_args validation.
	ud := (cgo.Handle)(methodUserData)
	bind, ok := ud.Value().(*GoMethodMetadata)
	if !ok || bind == nil {
		log.Panic("unable to retrieve methodUserData")
	}
	pnr.Pin(instPtr)
	inst := ObjectClassFromGDExtensionClassInstancePtr((GDExtensionClassInstancePtr)(instPtr))
	if inst == nil {
		log.Panic("GDExtensionClassInstancePtr canoot be null")
	}
	pnr.Pin(inst)

	// Reject argument-count mismatches before marshalling any arguments.
	if arity, expected := bind.classifyVarcallArity(int(argumentCount)); arity != varcallArityOK {
		rejectVarcallArity(rError, rReturn, arity, expected)
		return
	}

	argPtrSlice := unsafe.Slice((*GDExtensionConstVariantPtr)(argPtrs), int(argumentCount))
	// These are non-owning bit-copy views of the engine's borrowed arguments.
	// Nothing is allocated here, so the reject path frees nothing and must not
	// call Destroy on them (that would free engine-owned memory).
	args := make([]Variant, argumentCount)
	for i := range argPtrSlice {
		pnr.Pin(argPtrSlice[i])
		args[i] = NewVariantCopyWithGDExtensionConstVariantPtr(argPtrSlice[i])
	}

	// Reject the first argument whose variant type is not strictly convertible
	// to the bound parameter type (variant_can_convert_strict parity with
	// godot-cpp). This is a predicate only: no conversion is performed, so
	// rejection leaves no owned allocation to unwind.
	//
	// Variadic methods are exempt: their arguments are collected raw as Variants
	// (CallSlice) with no fixed per-parameter type to validate against — the
	// single declared slot is the slice itself, so checking it would wrongly
	// reject every argument (matching the arity pass, which also skips variadic).
	if !bind.IsVariadic {
		for i := range args {
			if !bind.varcallArgumentConvertible(args[i], i) {
				rejectVarcallInvalidArgument(rError, rReturn, i, int(bind.gdeArgumentTypes[i]))
				return
			}
		}
	}

	cn := inst.GetClass()
	defer cn.Destroy()
	log.Debug("GoCallback_MethodBindMethodCall called",
		zap.String("class", cn.ToUtf8()),
		zap.String("method", bind.GdMethodName),
		zap.String("bind", bind.String()),
	)
	retCall := bind.Call(inst, args...)
	*(*Variant)(unsafe.Pointer(rReturn)) = retCall
	pnr.Pin(rReturn)
	setCallErrorOK(rError)
}

// setCallErrorOK records a successful dispatch in the engine call-error slot.
func setCallErrorOK(rError *C.GDExtensionCallError) {
	rError.error = C.GDEXTENSION_CALL_OK
	rError.argument = -1
	rError.expected = -1
}

// rejectVarcallArity reports an argument-count mismatch and leaves a nil return.
func rejectVarcallArity(
	rError *C.GDExtensionCallError,
	rReturn C.GDExtensionVariantPtr,
	arity varcallArity,
	expected int,
) {
	writeNilCallReturn(rReturn)
	if arity == varcallArityTooMany {
		rError.error = C.GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS
	} else {
		rError.error = C.GDEXTENSION_CALL_ERROR_TOO_FEW_ARGUMENTS
	}
	rError.argument = -1
	rError.expected = C.int32_t(expected)
}

// rejectVarcallInvalidArgument reports a non-convertible argument and leaves a
// nil return. argument is the failing zero-based index; expected is the bound
// parameter's GDExtensionVariantType.
func rejectVarcallInvalidArgument(
	rError *C.GDExtensionCallError,
	rReturn C.GDExtensionVariantPtr,
	argument int,
	expected int,
) {
	writeNilCallReturn(rReturn)
	rError.error = C.GDEXTENSION_CALL_ERROR_INVALID_ARGUMENT
	rError.argument = C.int32_t(argument)
	rError.expected = C.int32_t(expected)
}

// writeNilCallReturn initializes the engine's return slot to nil so a rejected
// call is well-formed, matching the nil default godot-cpp assigns on error.
func writeNilCallReturn(rReturn C.GDExtensionVariantPtr) {
	nilReturn := NewVariantNil()
	*(*Variant)(unsafe.Pointer(rReturn)) = nilReturn
	pnr.Pin(rReturn)
}

// called when godot calls into golang code
//
//export GoCallback_MethodBindMethodPtrcall
func GoCallback_MethodBindMethodPtrcall(
	methodUserData unsafe.Pointer,
	instPtr C.GDExtensionClassInstancePtr,
	argPtrs *C.GDExtensionConstTypePtr,
	rReturn C.GDExtensionTypePtr,
) {
	ud := (cgo.Handle)(methodUserData)
	bind, ok := ud.Value().(*GoMethodMetadata)
	if !ok || bind == nil {
		log.Panic("unable to retrieve methodUserData")
	}
	inst := ObjectClassFromGDExtensionClassInstancePtr((GDExtensionClassInstancePtr)(instPtr))
	if inst == nil {
		log.Panic("GDExtensionClassInstancePtr canoot be null")
	}
	cn := inst.GetClass()
	defer cn.Destroy()
	log.Debug("GoCallback_MethodBindMethodPtrcall called",
		zap.String("class", cn.ToUtf8()),
		zap.String("method", bind.String()),
	)
	sliceLen := len(bind.GoArgumentTypes)
	argsSlice := unsafe.Slice((*GDExtensionConstTypePtr)(unsafe.Pointer(argPtrs)), sliceLen)
	bind.Ptrcall(
		inst,
		argsSlice,
		(GDExtensionUninitializedTypePtr)(rReturn),
	)
	pnr.Pin(rReturn)
}
