package gdextension

import (
	"reflect"

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
	CastTo(v Object) Object
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

func (w *WrappedImpl) CastTo(v Object) Object {
	owner := w.Owner

	t := reflect.TypeOf(v)

	className := t.Name()

	otherClassName := v.GetClassName()

	log.Info("WrappedImpl.CastTo called",
		zap.String("className", className),
		zap.String("otherClassName", otherClassName),
	)

	cn := NewStringNameWithUtf8Chars(className)
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

	ret := CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		casted,
		FFI.Token,
		&cbs)

	return *(*Object)(ret)
}

type WrappedClassInstance struct {
	Instance Object
}
