package core

import (
	"reflect"
	"strings"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// PropertySetGet holds metadata of the getter and setting functions of a Godot property.
// type PropertySetGet struct {
// 	Index   int
// 	Setter  string
// 	Getter  string
// 	_setptr *GoMethodMetadata
// 	_getptr *GoMethodMetadata
// 	Type    GDExtensionVariantType
// }

type MethodBindAndClassMethodInfo struct {
	GoMethodMetadata *GoMethodMetadata
	ClassMethodInfo  *GDExtensionClassMethodInfo
}

type StringSet map[string]struct{}

type ClassInfo struct {
	Name                 string
	NameStringName       StringName // stored by value — eliminates dangling pointer bug
	ParentName           string
	ParentNameStringName StringName // stored by value
	Level                GDExtensionInitializationLevel
	MethodMap                 map[string]*MethodBindAndClassMethodInfo
	VirtualMethodMap          map[string]*MethodBindAndClassMethodInfo
	SignalNameSet             StringSet
	SignalNameStringNames     map[string]StringName // signal name → StringName retained by Godot
	PropertyNameSet           StringSet
	ConstantNameSet           StringSet
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
	// Destroy StringNames in the property list only if FreePropertyList2
	// hasn't already been called (which sets PropertyList to nil).
	// If FreePropertyList2 was called, it already destroyed all StringNames
	// in the property list — destroying them here would be a double-free.
	if c.PropertyList != nil {
		for i := range c.PropertyList {
			c.PropertyList[i].Destroy()
		}
		c.PropertyList = nil
	}

	// Destroy signal name StringNames retained by Godot.
	for _, sn := range c.SignalNameStringNames {
		sn.Destroy()
	}
	c.SignalNameStringNames = nil

	// Copy to local vars to isolate from ClassInfo (which has Go pointers).
	// This avoids "Go pointer to unpinned Go pointer" cgo panic.
	name := c.NameStringName
	parentName := c.ParentNameStringName
	pnr.Pin(&name)
	pnr.Pin(&parentName)
	name.Destroy()
	parentName.Destroy()
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
		Name:                    name,
		NameStringName:          NewStringNameWithLatin1Chars(name),
		ParentName:              parentName,
		ParentNameStringName:    NewStringNameWithLatin1Chars(parentName),
		Level:                   level,
		MethodMap:               map[string]*MethodBindAndClassMethodInfo{},
		SignalNameSet:           map[string]struct{}{},
		SignalNameStringNames:   map[string]StringName{},
		VirtualMethodMap:        map[string]*MethodBindAndClassMethodInfo{},
		PropertyNameSet:         map[string]struct{}{},
		ConstantNameSet:         map[string]struct{}{},
		ParentPtr:               parentPtr,
		ClassType:               classType,
		InheritType:             inheritType,
		PropertyList:            propertyList,
		ValidateProperty:        validateProperty,
	}
	pnr.Pin(ret)
	return ret
}

var (
	classdbCurrentLevel GDExtensionInitializationLevel = GDEXTENSION_INITIALIZATION_CORE
)
