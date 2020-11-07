package gdnativeapijson

import (
	"strings"

	"github.com/godot-go/godot-go/cmd/generate/classes"

	"github.com/pinzolo/casee"
)

// ApiMetadata represents the Go GDNative API struct
type ApiMetadata struct {
	Name  string
	CType string
}

var ApiNameMap = map[string]ApiMetadata{
	"CORE:1.0":         {"CoreApi", "godot_gdnative_core_api_struct"},
	"CORE:1.1":         {"Core11Api", "godot_gdnative_core_1_1_api_struct"},
	"CORE:1.2":         {"Core12Api", "godot_gdnative_core_1_2_api_struct"},
	"NATIVESCRIPT:1.0": {"NativescriptApi", "godot_gdnative_ext_nativescript_api_struct"},
	"NATIVESCRIPT:1.1": {"Nativescript11Api", "godot_gdnative_ext_nativescript_1_1_api_struct"},
}

type GoProperty struct {
	Base      string // Base will let us know if this is a struct, int, etc.
	Name      string // The Go type name in camelCase
	CName     string // The C type name in snake_case
	Comment   string // Contains the comment on the line of the struct
	IsPointer bool   // Usually for properties; defines if it is a pointer type
}

// GoTypeDef represents the C Typedef in the godot_headers
type GoTypeDef struct {
	Base            string        // C Typedef tag
	Comment         string        // Contains the comment on the line of the struct
	Name            string        // The Go type name in camelCase
	CHeaderFilename string        // The header filename this type shows up in
	IsPointer       bool          // Usually for properties; defines if it is a pointer type
	CName           string        // The C type name in snake_case
	Properties      []GoProperty  // Optional C struct fields
	Usage           classes.Usage // Usage of the type
	Constructors    []GoMethod
	Methods         []GoMethod
}

func (o GoTypeDef) HasDestroyMethod() bool {
	for _, m := range o.Methods {
		if m.Name == "Destroy" {
			return true
		}
	}

	return false
}

// GoMethodType represents categories for each gdnative function exposed.
type GoMethodType int8

const (
	// UnknownGoMethodType represents an unexpected parsing error.
	UnknownGoMethodType GoMethodType = iota

	// ConstructorGoMethodType is the constructor type.
	ConstructorGoMethodType

	// GlobalGoMethodType is a global function type.
	GlobalGoMethodType

	// ReceiverGoMethodType is a receiver method type.
	ReceiverGoMethodType
)

type GoMethod struct {
	Name         string
	GoMethodType GoMethodType
	ReturnType   GoType
	Receiver     *GoArgument
	Arguments    []GoArgument
	CName        string
	ApiMetadata  ApiMetadata
}

func (t *GoMethod) IsStringReceiver() bool {
	return t.Receiver != nil && t.Receiver.Name == "String"
}

func (t *GoMethod) IsStringNameReceiver() bool {
	return t.Receiver != nil && t.Receiver.Name == "StringName"
}

func (t *GoMethod) IsStringReturnType() bool {
	return t.Receiver != nil && t.ReturnType.GoName == "String"
}

func (t *GoMethod) IsStringNameReturnType() bool {
	return t.Receiver != nil && t.ReturnType.GoName == "StringName"
}

func (t *GoMethod) IsSetter() bool {
	return t.ReturnType.NoReturnValue()
}

type GoArgument struct {
	Type GoType
	Name string
}

type ReferenceType int8

const (
	NoneReferenceType ReferenceType = iota
	PointerReferenceType
	PointerArrayReferenceType
)

type GoType struct {
	HasConst      bool
	ReferenceType ReferenceType
	CName         string
	GoName        string
}

func NewGoType(hasConst bool, referenceType ReferenceType, cName string) GoType {
	goName, _ := ToGoTypeName(cName)
	return GoType{
		HasConst:      hasConst,
		ReferenceType: referenceType,
		CName:         cName,
		GoName:        goName,
	}
}

func fixPascalCase(value string) string {
	var (
		result string
	)

	// TODO: hack to align cForGo names with typeName
	result = strings.Replace(value, "Aabb", "AABB", 1)
	result = strings.Replace(result, "Rid", "RID", 1)

	return result
}

func ToPascalCase(value string) string {
	result := casee.ToPascalCase(value)

	return fixPascalCase(result)
}

var cgoBuiltInTypeMap = map[string]string{
	"bool":       "bool",
	"uint8_t":    "uint8",
	"uint32_t":   "uint32",
	"uint64_t":   "uint64",
	"int64_t":    "int64",
	"double":     "float64",
	"wchar_t":    "int32",
	"char":       "C.char",
	"int":        "int32",
	"size_t":     "int64", // size_t should probalby map to int
	"void":       "void",
	"string":     "string",
	"float":      "float32",
	"godot_real": "float32",
	"godot_bool": "bool",
	"godot_int":  "int32",
}

// ToGoTypeName convert the C type into the expected Go type
func ToGoTypeName(value string) (string, classes.Usage) {
	if ret, ok := cgoBuiltInTypeMap[value]; ok {
		return ret, classes.USAGE_GO_PRIMATIVE
	}

	goTypeName := ToPascalCase(stripGodotPrefix(value))

	t, u := classes.GoTypeAndUsage(goTypeName)

	switch u {
	case classes.USAGE_GODOT_STRING:
		return "String", u
	case classes.USAGE_GODOT_STRING_NAME:
		return "StringName", u
	}

	return t, u
}

func stripGodotPrefix(name string) string {
	if name == "godot_object" {
		return name
	} else if strings.HasPrefix(name, "godot_") {
		return name[len("godot_"):]
	}

	return name
}

func (t GoType) IsGodotObjectPointer() bool {
	ret := t.CName == "godot_object"

	if ret && t.ReferenceType != PointerReferenceType {
		panic("type must be a pointer to a godot_object")
	}

	return ret
}

func (t GoType) IsGoPrimative() bool {
	_, ok := cgoBuiltInTypeMap[t.CName]

	return ok || t.CName == "godot_variant_type"
}

func (t GoType) IsMethodBindPointer() bool {
	return t.GoName == "MethodBind" && t.ReferenceType == PointerReferenceType
}

func (t GoType) IsGodotVariantPointerArray() bool {
	return t.CName == "godot_variant" && t.ReferenceType == PointerArrayReferenceType
}

func (t GoType) IsCharPointer() bool {
	return t.CName == "char" && t.ReferenceType == PointerReferenceType
}

func (t GoType) IsVoidPointer() bool {
	return t.CName == "void" && t.ReferenceType == PointerReferenceType
}

func (t GoType) IsVoidPointerArray() bool {
	return t.CName == "void" && t.ReferenceType == PointerArrayReferenceType
}

func (t GoType) IsPointer() bool {
	return t.ReferenceType == PointerReferenceType
}

func (t GoType) IsPointerArray() bool {
	return t.ReferenceType == PointerArrayReferenceType
}

func (t GoType) NoReturnValue() bool {
	return t.CName == "void" && t.ReferenceType == NoneReferenceType
}

func (t GoType) IsString() bool {
	return t.CName == "godot_string"
}

func (t GoType) IsStringName() bool {
	return t.CName == "godot_string_name"
}
