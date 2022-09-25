package gdextension

import (
	.
	"github.com/godot-go/godot-go/pkg/gdnative"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type GodotObject [0]byte

// Base for all engine classes, to contain the pointer to the engine instance.
type Wrapped interface {
	// GetExtensionClass() TypeName
	GetGodotObjectOwner() *GodotObject
	SetGodotObjectOwner(owner *GodotObject)
	GetClassName() TypeName
	GetParentClassName() TypeName
	CastTo(className TypeName) Wrapped
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

func (w *WrappedImpl) CastTo(className TypeName) Wrapped {
	owner := w.Owner

	tag := GDNativeInterface_classdb_get_class_tag(
		internal.gdnInterface,
		string(className),
	)

	if tag == nil {
		log.Panic("classTag unexpectedly came back nil", zap.String("type", string(className)))
	}

	casted := GDNativeInterface_object_cast_to(
		internal.gdnInterface,
		(GDNativeObjectPtr)(owner),
		tag,
	)

	if casted == nil {
		return nil
	}

	cbs, ok := gdExtensionBindingGDNativeInstanceBindingCallbacks.Get(className)

	if !ok {
		log.Warn("unable to find callbacks for Object")
		return nil
	}

	ret := GDNativeInterface_object_get_instance_binding(
		internal.gdnInterface,
		casted,
		internal.token,
		&cbs)

	return *(*Wrapped)(ret)
}
