package gdnative

// #include <godot/gdnative_interface.h>
// #include "gdnative_wrapper.gen.h"
import "C"
import (
	"sync/atomic"

	"github.com/godot-go/godot-go/pkg/log"
)

type PropertyInfo struct {
	Type       GDNativeVariantType
	Name       PropertyName
	ClassName  TypeName
	Hint       PropertyHint
	HintString string
	Usage      PropertyUsageFlags
}

func NewPropertyInfo(
	p_type GDNativeVariantType,
	p_name PropertyName,
	p_hint PropertyHint,
	p_hint_string string,
	p_usage PropertyUsageFlags,
	p_class_name TypeName,
) *PropertyInfo {
	var (
		cn TypeName
	)

	// behavior ported from godot-cpp
	switch p_hint {
	case PROPERTY_HINT_RESOURCE_TYPE:
		cn = (TypeName)(p_hint_string)
	default:
		cn = p_class_name
	}

	return &PropertyInfo{
		Type:       p_type,
		Name:       p_name,
		ClassName:  cn,
		Hint:       p_hint,
		HintString: p_hint_string,
		Usage:      p_usage,
	}
}

// MethodInfo implements sorting operations in godot-cpp
type MethodInfo struct {
	Name             string
	ReturnVal        PropertyInfo
	Flags            MethodFlags
	Arguments        []PropertyInfo
	DefaultArguments []Variant
}

func NewMethodInfo(
	p_ret PropertyInfo,
	p_name string,
	args []PropertyInfo,
) *MethodInfo {
	return &MethodInfo{
		Name:             p_name,
		ReturnVal:        p_ret,
		Flags:            METHOD_FLAG_NORMAL,
		Arguments:        args,
		DefaultArguments: []Variant{},
	}
}

// type ObjectID uint64

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
	obj := GDNativeInterface_object_get_instance_from_id(internal.gdnInterface, (GDObjectInstanceID)((C.uint64_t)(p_object_id)))

	if obj == nil {
		return nil
	}

	cbs, ok := gdExtensionBindingGDNativeInstanceBindingCallbacks.Get("Object")

	if !ok {
		log.Warn("unable to find callbacks for Object")
		return nil
	}

	binding := GDNativeInterface_object_get_instance_binding(
		internal.gdnInterface,
		obj,
		internal.token,
		&cbs,
	)

	return (*Object)(binding)
}

func (cx *Object) CastTo(p_object *Object, classTag string) Wrapped {
	owner := p_object.GetGodotObjectOwner()
	tag := GDNativeInterface_classdb_get_class_tag(
		internal.gdnInterface,
		classTag,
	)

	casted := GDNativeInterface_object_cast_to(
		internal.gdnInterface,
		(GDNativeObjectPtr)(owner),
		tag,
	)

	if casted == nil {
		return nil
	}

	cbs, ok := gdExtensionBindingGDNativeInstanceBindingCallbacks.Get("Object")

	if !ok {
		log.Warn("unable to find callbacks for Object")
		return nil
	}

	ret := GDNativeInterface_object_get_instance_binding(
		internal.gdnInterface,
		casted,
		internal.token,
		&cbs)

	return (*wrapped)(ret)
}
