package core

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/core
#include <godot/gdextension_interface.h>
#include "classdb_callback.h"
#include "method_bind.h"
*/
import "C"
import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	. "github.com/godot-go/godot-go/pkg/util"
	"go.uber.org/zap"
)

type InternalImpl struct {
	// GDClassInstances is used as a re-entrant FreeInstance guard.
	// Destroy() may trigger Godot to call FreeInstance again on the
	// same object. This map catches the double-call (panics instead
	// of corrupting the C heap with a double C++ destructor invocation).
	GDClassInstances      *SyncMap[GDObjectInstanceID, GDClass]
	GDRegisteredGDClasses *SyncMap[string, *ClassInfo]
	GDClassConstructors   *SyncMap[string, GDClassGoConstructorFromOwner]
}

var (
	nullptr                              = unsafe.Pointer(nil)
	Internal                             InternalImpl
	GDExtensionBindingInitCallbacks      [GDEXTENSION_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
	GDExtensionBindingTerminateCallbacks [GDEXTENSION_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
	pnr                                  runtime.Pinner
)

func CreateGDClassInstance[T GDClass]() T {
	t := reflect.TypeFor[T]()
	objectInst := reflect.Zero(t).Interface().(GDClass)
	inst := GDClass(objectInst)

	// Register this class within our plugin
	className := inst.GetClassName()
	parentName := inst.GetParentClassName()

	log.Debug("CreateGDClassInstance called",
		zap.String("class_name", className),
		zap.Any("parent_name", parentName),
	)
	snParentName := NewStringNameWithLatin1Chars(parentName)
	defer snParentName.Destroy()
	owner := CallFunc_GDExtensionInterfaceClassdbConstructObject2(
		snParentName.AsGDExtensionConstStringNamePtr(),
	)
	if owner == nil {
		log.Panic("owner is nil", zap.String("class_name", className))
	}
	// create GDClass
	object := (*GodotObject)(owner)
	inst.SetGodotObjectOwner(object)
	cbs, ok := GDExtensionBindingGDExtensionInstanceBindingCallbacks.Get(className)
	if !ok {
		log.Panic("missing instance binding callbacks", zap.String("class", className))
	}
	SetConstructInfo(inst.(Wrapped), className, cbs)
	id := CallFunc_GDExtensionInterfaceObjectGetInstanceId((GDExtensionConstObjectPtr)(unsafe.Pointer(owner)))
	Internal.GDClassInstances.Set(id, inst)
	log.Info("GDClass instance created",
		zap.Any("object_id", id),
		zap.String("class_name", className),
		zap.Any("parent_name", parentName),
		zap.String("inst", fmt.Sprintf("%p", inst)),
		zap.String("owner", fmt.Sprintf("%p", owner)),
		zap.String("object", fmt.Sprintf("%p", object)),
		zap.String("inst.GetGodotObjectOwner", fmt.Sprintf("%p", inst.GetGodotObjectOwner())),
	)
	return inst.(T)
}

func CreateGDClassInstance2(tn string) GDClass {
	ci, ok := Internal.GDRegisteredGDClasses.Get(tn)
	if !ok {
		log.Panic("type not found",
			zap.String("name", tn),
		)
	}
	log.Debug("CreateGDClassInstance2 called",
		zap.String("class_name", tn),
		zap.Any("parent_name", ci.ParentName),
	)
	snParentName := NewStringNameWithLatin1Chars(ci.ParentName)
	defer snParentName.Destroy()
	owner := CallFunc_GDExtensionInterfaceClassdbConstructObject2(
		snParentName.AsGDExtensionConstStringNamePtr(),
	)
	if owner == nil {
		log.Panic("owner is nil", zap.String("type_name", tn))
	}
	reflectedInst := reflect.New(ci.ClassType)
	inst, ok := reflectedInst.Interface().(GDClass)
	if !ok {
		log.Panic("instance not a GDClass", zap.String("type_name", tn))
	}
	object := (*GodotObject)(owner)
	inst.SetGodotObjectOwner(object)
	WrappedPostInitialize(tn, inst)
	id := CallFunc_GDExtensionInterfaceObjectGetInstanceId((GDExtensionConstObjectPtr)(unsafe.Pointer(owner)))
	Internal.GDClassInstances.Set(id, inst)
	log.Info("GDClass instance created",
		zap.Any("object_id", id),
		zap.String("class_name", tn),
		zap.Any("parent_name", ci.ParentName),
		zap.String("inst", fmt.Sprintf("%p", inst)),
		zap.String("owner", fmt.Sprintf("%p", owner)),
		zap.String("object", fmt.Sprintf("%p", object)),
		zap.String("inst.GetGodotObjectOwner", fmt.Sprintf("%p", inst.GetGodotObjectOwner())),
	)
	return inst
}
