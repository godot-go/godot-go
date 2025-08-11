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
	// TODO: implement rError checking
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
	cn := inst.GetClass()
	// defer cn.Destroy()
	log.Debug("GoCallback_MethodBindMethodCall called",
		zap.String("class", cn.ToUtf8()),
		zap.String("method", bind.GdMethodName),
		zap.String("bind", bind.String()),
	)
	argPtrSlice := unsafe.Slice((*GDExtensionConstVariantPtr)(argPtrs), int(argumentCount))
	args := make([]Variant, argumentCount)
	for i := range argPtrSlice {
		pnr.Pin(argPtrSlice[i])
		args[i] = NewVariantCopyWithGDExtensionConstVariantPtr(argPtrSlice[i])
	}
	retCall := bind.Call(inst, args...)
	*(*Variant)(unsafe.Pointer(rReturn)) = retCall
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
