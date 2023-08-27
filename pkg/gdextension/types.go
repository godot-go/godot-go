package gdextension

import (
	"reflect"
	"strings"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
)

// PropertySetGet holds metadata of the getter and setting functions of a Godot property.
type PropertySetGet struct {
	Index   int
	Setter  string
	Getter  string
	_setptr *MethodBind
	_getptr *MethodBind
	Type    GDExtensionVariantType
}

type ClassInfo struct {
	Name                      string
	NameAsStringNamePtr       GDExtensionConstStringNamePtr
	ParentName                string
	ParentNameAsStringNamePtr GDExtensionConstStringNamePtr
	Level                     GDExtensionInitializationLevel
	MethodMap                 map[string]*MethodBind
	SignalNameMap             map[string]struct{}
	VirtualMethodMap          map[string]*MethodBind
	PropertyNameMap           map[string]struct{}
	ConstantNameMap           map[string]struct{}
	ParentPtr                 *ClassInfo
	ClassType                 reflect.Type
	InheritType               reflect.Type
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

	for _, v := range c.MethodMap {
		v.Destroy()
	}
}

func NewClassInfo(name, parentName string, level GDExtensionInitializationLevel, classType, inheritType reflect.Type, parentPtr *ClassInfo) *ClassInfo {
	return &ClassInfo{
		Name:                name,
		NameAsStringNamePtr: NewStringNameWithUtf8Chars(name).AsGDExtensionConstStringNamePtr(),
		ParentName:          parentName,
		Level:               level,
		MethodMap:           map[string]*MethodBind{},
		SignalNameMap:       map[string]struct{}{},
		VirtualMethodMap:    map[string]*MethodBind{},
		PropertyNameMap:     map[string]struct{}{},
		ConstantNameMap:     map[string]struct{}{},
		ParentPtr:           parentPtr,
		ClassType:           classType,
		InheritType:         inheritType,
	}
}

var (
	classdbCurrentLevel GDExtensionInitializationLevel = GDEXTENSION_INITIALIZATION_CORE
)
