package gdextension

// #include <godot/gdextension_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
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
	bind := (*MethodBindImpl)(methodUserData)
	if bind == nil {
		log.Panic("unable to retrieve methodUserData")
	}
	inst := ObjectClassFromGDExtensionClassInstancePtr((GDExtensionClassInstancePtr)(instPtr))
	if inst == nil {
		log.Panic("GDExtensionClassInstancePtr canoot be null")
	}
	cn := inst.GetClass()
	defer cn.Destroy()
	log.Debug("GoCallback_MethodBindMethodCall called",
		zap.String("class", cn.ToUtf8()),
		zap.String("method", bind.MethodName),
		zap.String("bind", bind.String()),
	)
	argPtrSlice := unsafe.Slice((*GDExtensionConstVariantPtr)(argPtrs), int(argumentCount))
	args := make([]Variant, argumentCount)
	for i := range argPtrSlice {
		args[i] = NewVariantCopyWithGDExtensionConstVariantPtr(argPtrSlice[i])
	}
	retCall := bind.Call(inst, args)
	copyVariantWithGDExtensionTypePtr((GDExtensionUninitializedVariantPtr)(rReturn), retCall.nativeConstPtr())
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
	bind := (*MethodBindImpl)(methodUserData)
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
	argsSlice := unsafe.Slice((*GDExtensionConstTypePtr)(unsafe.Pointer(argPtrs)), len(bind.MethodMetadata.GoArgumentTypes))
	bind.Ptrcall(
		inst,
		argsSlice,
		(GDExtensionUninitializedTypePtr)(rReturn),
	)
}
