package core

import (
	"fmt"
	"reflect"

	. "github.com/godot-go/godot-go/pkg/builtin"
	"github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// PropertySpec describes a property to register on a Go-defined Godot class
// with full inspector metadata.
//
// Zero-value semantics:
//   - Hint zero means PROPERTY_HINT_NONE.
//   - Usage zero is a sentinel meaning PROPERTY_USAGE_DEFAULT; explicit
//     reduced flag sets must OR real flags (e.g. PROPERTY_USAGE_STORAGE).
//   - ClassName empty means the class qualifier is derived; see
//     deriveClassName.
type PropertySpec struct {
	// VariantType is the property's Godot variant type.
	VariantType GDExtensionVariantType
	// Name is the property name as seen by GDScript.
	Name string
	// Setter is the bound method name used to set the property; empty means
	// the property is read-only.
	Setter string
	// Getter is the bound method name used to get the property; required.
	Getter string
	// Hint is the inspector hint (PROPERTY_HINT_*).
	Hint constant.PropertyHint
	// HintString carries hint-specific data (range string, enum names,
	// resource class name, ...).
	HintString string
	// Usage is the property usage flag set (PROPERTY_USAGE_*); zero means
	// PROPERTY_USAGE_DEFAULT.
	Usage constant.PropertyUsageFlags
	// ClassName is an explicit class qualifier for object-typed properties.
	// When empty, the qualifier is derived from the hint or the owning class.
	ClassName string
}

// deriveClassName resolves the class_name field of the registered property
// info. Behavior ported from godot-cpp: a resource-type hint names the
// resource class through the hint string; otherwise an explicit qualifier
// wins, and otherwise the owning class name is used.
func deriveClassName(spec PropertySpec, owningClassName string) string {
	// behavior ported from godot-cpp
	if spec.Hint == constant.PROPERTY_HINT_RESOURCE_TYPE {
		return spec.HintString
	}
	if spec.ClassName != "" {
		return spec.ClassName
	}
	return owningClassName
}

// validatePropertySpec performs the authoring-time checks that need nothing
// but the spec itself. Failures are authoring faults and fatal, matching the
// repo convention that caller-inducible errors alone are reported softly.
func validatePropertySpec(spec PropertySpec) {
	if spec.Getter == "" {
		log.Panic(`Getter method must be specified.`,
			zap.String("property", spec.Name),
		)
	}
	if spec.Hint == constant.PROPERTY_HINT_RESOURCE_TYPE && spec.HintString == "" {
		panic(fmt.Sprintf(`Property "%s" uses PROPERTY_HINT_RESOURCE_TYPE but declares an empty hint string.`, spec.Name))
	}
}

// ClassDBAddPropertyInfo registers a property on a Go-defined class with the
// full hint/hint string/usage metadata from the spec.
//
// All Godot string temporaries are created and owned here: they are pinned
// for the duration of the registration call and destroyed after it returns,
// because Godot copies the PropertyInfo pointer data during registration
// (same lifecycle precedent as ClassDBAddSignal).
func ClassDBAddPropertyInfo(inst GDClass, spec PropertySpec) {
	t := reflect.TypeOf(inst)
	cn := inst.GetClassName()
	ci, ok := Internal.GDRegisteredGDClasses.Get(cn)
	if !ok {
		log.Panic("Trying to add property to non-existing class.",
			zap.String("property", spec.Name),
			zap.String("p_class", (string)(cn)),
			zap.Any("type", t),
		)
	}
	if _, ok := ci.PropertyNameSet[spec.Name]; ok {
		panic(fmt.Sprintf(`Property "%s" already exists in class "%s".`, spec.Name, cn))
	}
	validatePropertySpec(spec)
	mci, ok := ci.MethodMap[spec.Getter]
	if !ok {
		log.Panic("unable to find getter",
			zap.String("getter", spec.Getter),
		)
	}
	getter := mci.GoMethodMetadata
	if len(getter.GoArgumentTypes) != 0 {
		panic(fmt.Sprintf(`getter method "%s" must take a single argument.`, spec.Getter))
	}
	// specifying a setter is optional
	var setter *GoMethodMetadata
	if len(spec.Setter) > 0 {
		mci, ok := ci.MethodMap[spec.Setter]
		if !ok {
			log.Panic("unable to find setter",
				zap.String("setter", spec.Setter),
			)
		}
		setter = mci.GoMethodMetadata
		if len(setter.GoArgumentTypes) != 1 {
			panic(fmt.Sprintf(`Setter method "%s" must take a single argument.`, spec.Setter))
		}
	}
	// register property with plugin
	ci.PropertyNameSet[spec.Name] = struct{}{}

	// Build the refcounted temporaries that back the PropertyInfo pointers.
	// Godot copies the pointer data during registration, so these are
	// destroyed once the call has returned.
	propName := NewStringNameWithLatin1Chars(spec.Name)
	defer propName.Destroy()
	hintString := NewStringWithUtf8Chars(spec.HintString)
	defer hintString.Destroy()
	className := deriveClassName(spec, cn)
	classNameSn := NewStringNameWithLatin1Chars(className)
	defer classNameSn.Destroy()
	usage := uint32(constant.PROPERTY_USAGE_DEFAULT)
	if spec.Usage != 0 {
		usage = uint32(spec.Usage)
	}
	pnr.Pin(&propName)
	pnr.Pin(&hintString)
	pnr.Pin(&classNameSn)
	prop_info := NewGDExtensionPropertyInfo(
		classNameSn.AsGDExtensionConstStringNamePtr(),
		spec.VariantType,
		propName.AsGDExtensionConstStringNamePtr(),
		uint32(spec.Hint),
		hintString.AsGDExtensionConstStringPtr(),
		usage,
	)
	pnr.Pin(&prop_info)
	// An absent setter is registered as an empty setter name (read-only
	// property), never as a nil dereference.
	setterMethodName := ""
	if setter != nil {
		setterMethodName = setter.GdMethodName
	}
	snSetterGDName := NewStringNameWithLatin1Chars(setterMethodName)
	defer snSetterGDName.Destroy()
	snGetterGDName := NewStringNameWithLatin1Chars(getter.GdMethodName)
	defer snGetterGDName.Destroy()
	log.Info("register property",
		zap.String("class", cn),
		zap.String("name", spec.Name),
		zap.Int("variant_type", int(spec.VariantType)),
		zap.Int("hint", int(spec.Hint)),
	)
	cnName := ci.NameStringName
	pnr.Pin(&cnName)
	pnr.Pin(&snSetterGDName)
	pnr.Pin(&snGetterGDName)
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertyIndexed(
		FFI.Library,
		cnName.AsGDExtensionConstStringNamePtr(),
		&prop_info,
		snSetterGDName.AsGDExtensionConstStringNamePtr(),
		snGetterGDName.AsGDExtensionConstStringNamePtr(),
		-1,
	)
}
