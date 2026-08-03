package builtin

import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type WrappedImpl struct {
	// Owner must be public but do not directly modify.
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

// NewWrapped creates a new WrappedImpl and its underlying GodotObject.
// Wrapped::Wrapped(const StringName &p_godot_class) in godot-cpp
// func NewWrappedFromClassName(className string) *WrappedImpl {
// 	log.Debug("NewWrappedFromClassName called")
// 	snclassName := NewStringNameWithLatin1Chars(className)
// 	defer snclassName.Destroy()
// 	snClassNamePtr := snclassName.AsGDExtensionConstStringNamePtr()
// 	owner := CallFunc_GDExtensionInterfaceClassdbConstructObject2(snClassNamePtr)
// 	w := &WrappedImpl{
// 		Owner: (*GodotObject)(unsafe.Pointer(owner)),
// 	}
// 	instHandle := cgo.NewHandle(w)

// 	cbs, ok := GDExtensionBindingGDExtensionInstanceBindingCallbacks.Get(className)
// 	if !ok {
// 		log.Warn("unable to find callbacks for Object",
// 			zap.String("name", className),
// 		)
// 		return nil
// 	}

// 	// Set the extension instance in the native Godot object.
// 	CallFunc_GDExtensionInterfaceObjectSetInstance(
// 		(GDExtensionObjectPtr)(owner),
// 		(GDExtensionConstStringNamePtr)(snClassNamePtr),
// 		(GDExtensionClassInstancePtr)(instHandle),
// 	)
// 	CallFunc_GDExtensionInterfaceObjectSetInstanceBinding(
// 		(GDExtensionObjectPtr)(owner),
// 		unsafe.Pointer(FFI.Token),
// 		instHandle,
// 		&cbs,
// 	)
// 	return w
// }

// NewWrappedFromGodotObject creates a new WrappedImpl from an existing GodotObject..
//
//	Wrapped::Wrapped(GodotObject *p_godot_object) in godot-cpp
func NewWrappedFromGodotObject(owner *GodotObject) *WrappedImpl {
	return &WrappedImpl{
		Owner: owner,
	}
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
	cbs, ok := GDExtensionBindingGDExtensionInstanceBindingCallbacks.Get(className)
	if !ok {
		log.Warn("unable to find callbacks for Object",
			zap.String("name", className),
		)
		return nil
	}
	cbsPtr := &cbs
	pnr.Pin(casted)
	pnr.Pin(cbsPtr)
	// TODO: validate this is working as expected
	inst := CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		casted,
		FFI.Token,
		cbsPtr)
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

func (w *WrappedClassInstance) GetClassName() string {
	return w.Instance.GetClassName()
}

func (w *WrappedClassInstance) GetParentClassName() string {
	return w.Instance.GetParentClassName()
}

func (w *WrappedClassInstance) GetGodotObjectOwner() *GodotObject {
	return w.Instance.GetGodotObjectOwner()
}

func (w *WrappedClassInstance) AsGDExtensionObjectPtr() GDExtensionObjectPtr {
	return (GDExtensionObjectPtr)(unsafe.Pointer(w.Instance.GetGodotObjectOwner()))
}

func (w *WrappedClassInstance) AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr {
	return (GDExtensionConstObjectPtr)(unsafe.Pointer(w.Instance.GetGodotObjectOwner()))
}

func (w *WrappedClassInstance) AsGDExtensionTypePtr() GDExtensionTypePtr {
	return (GDExtensionTypePtr)(unsafe.Pointer(w.Instance.GetGodotObjectOwner()))
}

func (w *WrappedClassInstance) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return (GDExtensionConstTypePtr)(unsafe.Pointer(w.Instance.GetGodotObjectOwner()))
}

func (w *WrappedClassInstance) SetGodotObjectOwner(owner *GodotObject) {
	w.Instance.SetGodotObjectOwner(owner)
}

func (w *WrappedClassInstance) IsNil() bool {
	return w == nil || w.Instance.GetGodotObjectOwner() == nil
}

func (w *WrappedClassInstance) Destroy() {
}
