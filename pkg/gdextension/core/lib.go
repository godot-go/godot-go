package core

/*
#cgo CFLAGS: -I${SRCDIR}/../../../godot_headers -I${SRCDIR}/../../../pkg/log -I${SRCDIR}/../../../pkg/gdextension/core
#include <godot/gdextension_interface.h>
#include "classdb_callback.h"
#include "method_bind.h"
*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextension/builtin"
	. "github.com/godot-go/godot-go/pkg/gdextension/ffi"
	. "github.com/godot-go/godot-go/pkg/gdextension/nativestructure"
	. "github.com/godot-go/godot-go/pkg/util"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type InternalImpl struct {
	GDNativeInstances     *SyncMap[ObjectID, GDExtensionClass]
	GDClassInstances      *SyncMap[GDObjectInstanceID, GDClass]
	GDRegisteredGDClasses *SyncMap[string, *ClassInfo]
}

var (
	nullptr = unsafe.Pointer(nil)
	Internal InternalImpl
	GDExtensionBindingInitCallbacks      [GDEXTENSION_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
	GDExtensionBindingTerminateCallbacks [GDEXTENSION_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
)

func CreateGDClassInstance(tn string) GDClass {
	ci, ok := Internal.GDRegisteredGDClasses.Get(tn)

	if !ok {
		log.Panic("type not found",
			zap.String("name", tn),
		)
	}

	log.Debug("CreateGDClassInstance called",
		zap.String("class_name", tn),
		zap.Any("parent_name", ci.ParentName),
	)

	snParentName := NewStringNameWithLatin1Chars(ci.ParentName)
	defer snParentName.Destroy()

	// create inherited GDExtensionClass first
	owner := CallFunc_GDExtensionInterfaceClassdbConstructObject(
		snParentName.AsGDExtensionConstStringNamePtr(),
	)

	if owner == nil {
		log.Panic("owner is nil", zap.String("type_name", tn))
	}

	// create GDClass
	reflectedInst := reflect.New(ci.ClassType)

	inst, ok := reflectedInst.Interface().(GDClass)

	if !ok {
		log.Panic("instance not a GDClass", zap.String("type_name", tn))
	}

	object := (*GodotObject)(unsafe.Pointer(owner))

	id := CallFunc_GDExtensionInterfaceObjectGetInstanceId((GDExtensionConstObjectPtr)(unsafe.Pointer(owner)))

	inst.SetGodotObjectOwner(object)

	WrappedPostInitialize(tn, inst)

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
