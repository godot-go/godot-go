package gdextension

// #include <godot/gdextension_interface.h>
import "C"
import (
	"sync/atomic"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
)

func NewSimpleGDExtensionPropertyInfo(
	className string,
	variantType GDExtensionVariantType,
	name string,
) GDExtensionPropertyInfo {
	// TODO: move to alloc memory in C?
	return NewGDExtensionPropertyInfo(
		NewStringNameWithUtf8Chars(className).AsGDExtensionConstStringNamePtr(),
		variantType,
		NewStringNameWithUtf8Chars(name).AsGDExtensionConstStringNamePtr(),
		uint32(PROPERTY_HINT_NONE),
		NewStringWithUtf8Chars("").AsGDExtensionConstStringPtr(),
		uint32(PROPERTY_USAGE_DEFAULT),
	)
}

func (o ObjectID) compare(other ObjectID) int {
	if o.Id < other.Id {
		return -1
	} else if o.Id == other.Id {
		return 0
	}
	return 1
}

func (o ObjectID) IsRefCounted() bool {
	return (uint64(o.Id) & (uint64(1) << 63)) != 0
}

func (o ObjectID) IsValid() bool {
	return uint64(o.Id) != uint64(0)
}

func (o ObjectID) IsNull() bool {
	return uint64(o.Id) == uint64(0)
}

var (
	lastObjectIDValue uint64
)

func NewObjectID() ObjectID {
	v := atomic.AddUint64(&lastObjectIDValue, 1)

	return ObjectID{Id: v}
}

func objectDBGetInstance(p_object_id GDObjectInstanceID) *Object {
	obj := CallFunc_GDExtensionInterfaceObjectGetInstanceFromId((GDObjectInstanceID)((C.uint64_t)(p_object_id)))

	if obj == nil {
		return nil
	}

	cbs, ok := gdExtensionBindingGDExtensionInstanceBindingCallbacks.Get("Object")

	if !ok {
		log.Warn("unable to find callbacks for Object")
		return nil
	}

	binding := CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		obj,
		FFI.Token,
		&cbs,
	)

	return (*Object)(binding)
}
