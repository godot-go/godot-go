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
	GetGodotObjectOwner() *GodotObject
	SetGodotObjectOwner(owner *GodotObject)
	GetClassName() string
	GetParentClassName() string
	CastTo(className string) Wrapped
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

func (w *WrappedImpl) CastTo(className string) Wrapped {
	owner := w.Owner

	tag := GDExtensionInterface_classdb_get_class_tag(
		internal.gdnInterface,
		NewStringNameWithLatin1Chars(className).AsGDExtensionStringNamePtr(),
	)

	if tag == nil {
		log.Panic("classTag unexpectedly came back nil", zap.String("type", className))
	}

	casted := GDExtensionInterface_object_cast_to(
		internal.gdnInterface,
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

	ret := GDExtensionInterface_object_get_instance_binding(
		internal.gdnInterface,
		casted,
		internal.token,
		&cbs)

	return *(*Wrapped)(ret)
}
