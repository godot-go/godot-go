package core

import (
	"reflect"
	"strings"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// PropertySetGet holds metadata of the getter and setting functions of a Godot property.
type PropertySetGet struct {
	Index   int
	Setter  string
	Getter  string
	_setptr *GoMethodMetadata
	_getptr *GoMethodMetadata
	Type    GDExtensionVariantType
}

type MethodBindAndClassMethodInfo struct {
	GoMethodMetadata *GoMethodMetadata
	ClassMethodInfo  *GDExtensionClassMethodInfo
}

type ClassInfo struct {
	Name                      string
	NameAsStringNamePtr       GDExtensionConstStringNamePtr
	ParentName                string
	ParentNameAsStringNamePtr GDExtensionConstStringNamePtr
	Level                     GDExtensionInitializationLevel
	MethodMap                 map[string]*MethodBindAndClassMethodInfo
	SignalNameMap             map[string]struct{}
	VirtualMethodMap          map[string]*MethodBindAndClassMethodInfo
	PropertyNameMap           map[string]struct{}
	ConstantNameMap           map[string]struct{}
	ParentPtr                 *ClassInfo
	ClassType                 reflect.Type
	InheritType               reflect.Type
	PropertyList              []GDExtensionPropertyInfo
	ValidateProperty          func(*GDExtensionPropertyInfo)
}

func (c *ClassInfo) String() string {
	var sb strings.Builder
	sb.WriteString(c.Name)
	sb.WriteString("(")
	sb.WriteString(c.ParentName)
	sb.WriteString(")")
	return sb.String()
}

func (c *ClassInfo) Destroy() {
	name := (*StringName)(unsafe.Pointer(c.NameAsStringNamePtr))
	if name != nil {
		name.Destroy()
	}

	parentName := (*StringName)(unsafe.Pointer(c.ParentNameAsStringNamePtr))
	if parentName != nil {
		parentName.Destroy()
	}

	for _, v := range c.VirtualMethodMap {
		v.ClassMethodInfo.Destroy()
	}

	for _, v := range c.MethodMap {
		v.ClassMethodInfo.Destroy()
	}
}

func NewClassInfo(
	name, parentName string,
	level GDExtensionInitializationLevel,
	classType, inheritType reflect.Type,
	parentPtr *ClassInfo,
	propertyList []GDExtensionPropertyInfo,
	validateProperty func(*GDExtensionPropertyInfo),
) *ClassInfo {
	ret := &ClassInfo{
		Name:                name,
		NameAsStringNamePtr: NewStringNameWithLatin1Chars(name).AsGDExtensionConstStringNamePtr(),
		ParentName:          parentName,
		Level:               level,
		MethodMap:           map[string]*MethodBindAndClassMethodInfo{},
		SignalNameMap:       map[string]struct{}{},
		VirtualMethodMap:    map[string]*MethodBindAndClassMethodInfo{},
		PropertyNameMap:     map[string]struct{}{},
		ConstantNameMap:     map[string]struct{}{},
		ParentPtr:           parentPtr,
		ClassType:           classType,
		InheritType:         inheritType,
		PropertyList:        propertyList,
		ValidateProperty:    validateProperty,
	}
	pnr.Pin(ret)
	return ret
}

var (
	classdbCurrentLevel GDExtensionInitializationLevel = GDEXTENSION_INITIALIZATION_CORE
)
