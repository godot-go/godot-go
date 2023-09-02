package gdextension

import (
	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type GodotObject [0]byte

// Base for all engine classes, to contain the pointer to the engine instance.
type Wrapped interface {
	GetGodotObjectOwner() *GodotObject
	SetGodotObjectOwner(owner *GodotObject)
	GetClassName() string
	GetParentClassName() string
}

type WrappedImpl struct {
	// Must be public but you should not touch this.
	Owner *GodotObject
}

func (w *WrappedImpl) GetGodotObjectOwner() *GodotObject {
	return w.Owner
}

func (w *WrappedImpl) SetGodotObjectOwner(owner *GodotObject) {
	w.Owner = owner
}

// func (w *WrappedImpl) GetClassName() string {
// 	return "Wrapped"
// }

func ObjectCastTo(obj Object, className string) Object {
	if obj == nil {
		return nil
	}
	log.Info("ObjectCastTo called",
		zap.String("className", obj.GetClassName()),
		zap.String("otherClassName", className),
	)
	owner := obj.GetGodotObjectOwner()
	cn := NewStringNameWithLatin1Chars(className)
	defer cn.Destroy()
	tag := CallFunc_GDExtensionInterfaceClassdbGetClassTag(
		cn.AsGDExtensionConstStringNamePtr(),
	)
	if tag == nil {
		log.Panic("classTag unexpectedly came back nil", zap.String("type", className))
	}
	casted := CallFunc_GDExtensionInterfaceObjectCastTo(
		(GDExtensionConstObjectPtr)(owner),
		tag,
	)
	if casted == nil {
		return nil
	}
	cbs, ok := gdExtensionBindingGDExtensionInstanceBindingCallbacks.Get(className)
	if !ok {
		log.Warn("unable to find callbacks for Object")
		return nil
	}
	inst := CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		casted,
		FFI.Token,
		&cbs)
	wci := (*WrappedClassInstance)(inst)
	return wci.Instance
}

type WrappedClassInstance struct {
	Instance Object
}
