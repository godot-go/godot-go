package gdextension

import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type GodotObject unsafe.Pointer

// Base for all engine classes, to contain the pointer to the engine instance.
type Wrapped interface {
	HasDestructor
	GetGodotObjectOwner() *GodotObject
	SetGodotObjectOwner(owner *GodotObject)
	GetClassName() string
	GetParentClassName() string
	AsGDExtensionObjectPtr() GDExtensionObjectPtr
	AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr
	AsGDExtensionTypePtr() GDExtensionTypePtr
	AsGDExtensionConstTypePtr() GDExtensionConstTypePtr
}

type WrappedImpl struct {
	// Must be public but you should not touch this.
	Owner *GodotObject
}

func (w *WrappedImpl) GetGodotObjectOwner() *GodotObject {
	return w.Owner
}

func (w *WrappedImpl) AsGDExtensionObjectPtr() GDExtensionObjectPtr {
	return (GDExtensionObjectPtr)(unsafe.Pointer(w.Owner))
}

func (w *WrappedImpl) AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr {
	return (GDExtensionConstObjectPtr)(unsafe.Pointer(w.Owner))
}

func (w *WrappedImpl) AsGDExtensionTypePtr() GDExtensionTypePtr {
	return (GDExtensionTypePtr)(unsafe.Pointer(&w.Owner))
}

func (w *WrappedImpl) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return (GDExtensionConstTypePtr)(unsafe.Pointer(&w.Owner))
}

func (w *WrappedImpl) SetGodotObjectOwner(owner *GodotObject) {
	w.Owner = owner
}

func (w *WrappedImpl) IsNil() bool {
	return w == nil || w.Owner == nil
}

func (w *WrappedImpl) Destroy() {
}

func (cx *ObjectImpl) ToGoString() string {
	if cx == nil || cx.Owner == nil {
		return ""
	}
	gdstr := cx.ToString()
	defer gdstr.Destroy()
	return gdstr.ToUtf8()
}

func ObjectCastTo(obj Object, className string) Object {
	if obj == nil {
		return nil
	}
	gdStrCn := obj.GetClass()
	defer gdStrCn.Destroy()
	log.Info("ObjectCastTo called",
		zap.String("class", gdStrCn.ToUtf8()),
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
	// TODO: this is probably broken
	inst := CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		casted,
		FFI.Token,
		&cbs)
	wci := (*WrappedClassInstance)(inst)
	wrapperClassName := wci.Instance.GetClassName()
	gdStrClassName := wci.Instance.GetClass()
	defer gdStrClassName.Destroy()
	log.Info("ObjectCastTo casted",
		zap.String("class", gdStrClassName.ToUtf8()),
		zap.String("className", wrapperClassName),
	)
	return wci.Instance
}

type WrappedClassInstance struct {
	Instance Object
}
