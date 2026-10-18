package core

import (
	"fmt"
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
	Name                  string
	NameStringName        StringName // stored by value — eliminates dangling pointer bug
	ParentName            string
	ParentNameStringName  StringName // stored by value
	Level                 GDExtensionInitializationLevel
	MethodMap             map[string]*MethodBindAndClassMethodInfo
	VirtualMethodMap      map[string]*MethodBindAndClassMethodInfo
	SignalNameSet         StringSet
	SignalNameStringNames map[string]StringName // signal name → StringName retained by Godot
	PropertyNameSet       StringSet
	ConstantNameSet       StringSet
	ParentPtr             *ClassInfo
	ClassType             reflect.Type
	InheritType           reflect.Type
	propertySpecs         []classPropertySpec // class-owned backing store for the exposed property list
	ValidateProperty      func(*GDExtensionPropertyInfo)
}

func (c *ClassInfo) HasError() error {
	if c == nil {
		return fmt.Errorf("class info is null")
	}
	if len(c.Name) == 0 {
		return fmt.Errorf("name is empty")
	}
	return nil
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
	// The exposed-array ownership split (openspec fix-class-property-list-lifecycle):
	// each query owns its C array (released by FreePropertyList2), the class owns
	// the string cells and objects torn down here exactly once. No nil-check
	// dance against query timing is needed anymore.
	for i := range c.propertySpecs {
		c.propertySpecs[i].destroy()
	}
	c.propertySpecs = nil

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
	properties []PropertyDescriptor,
	validateProperty func(*GDExtensionPropertyInfo),
) *ClassInfo {
	ret := &ClassInfo{
		Name:                  name,
		NameStringName:        NewStringNameWithLatin1Chars(name),
		ParentName:            parentName,
		ParentNameStringName:  NewStringNameWithLatin1Chars(parentName),
		Level:                 level,
		MethodMap:             map[string]*MethodBindAndClassMethodInfo{},
		SignalNameSet:         map[string]struct{}{},
		SignalNameStringNames: map[string]StringName{},
		VirtualMethodMap:      map[string]*MethodBindAndClassMethodInfo{},
		PropertyNameSet:       map[string]struct{}{},
		ConstantNameSet:       map[string]struct{}{},
		ParentPtr:             parentPtr,
		ClassType:             classType,
		InheritType:           inheritType,
		propertySpecs:         buildClassPropertySpecs(properties),
		ValidateProperty:      validateProperty,
	}
	pnr.Pin(ret)
	return ret
}

var (
	classdbCurrentLevel GDExtensionInitializationLevel = GDEXTENSION_INITIALIZATION_CORE
)
