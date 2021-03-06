package classes

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/pinzolo/casee"
)

type ApiType int8

const (
	CoreApiType = iota
	ToolsApiType
)

func ParseApiType(value string) ApiType {
	switch value {
	case "core":
		return CoreApiType
	case "tools":
		return ToolsApiType
	}

	log.Panicf("unrecognized api type %s", value)
	return -1
}

var enumTypeRegex = regexp.MustCompile(`^enum\._?([\w\d_]+)::([\w\d_]+)$`)

func cType(value string) string {
	switch value {
	case "bool":
		return "char"
	case "float":
		return "double"
	}

	return value
}

// Usage is an enum type categorizing data types for parsing and generating NativeScript classes.
type Usage int8

//revive:disable:var-naming
const (
	USAGE_VOID Usage = iota
	USAGE_GO_PRIMATIVE
	USAGE_GDNATIVE_CONST_OR_ENUM
	USAGE_GODOT_STRING
	USAGE_GODOT_STRING_NAME
	USAGE_GDNATIVE_REF
	USAGE_GODOT_CONST_OR_ENUM
	USAGE_GODOT_CLASS
)
//revive:enable:var-naming

func (u Usage) String() string {
	switch u {
	case USAGE_VOID:
		return "USAGE_VOID"
	case USAGE_GO_PRIMATIVE:
		return "USAGE_GO_PRIMATIVE"
	case USAGE_GDNATIVE_CONST_OR_ENUM:
		return "USAGE_GDNATIVE_CONST_OR_ENUM"
	case USAGE_GODOT_STRING:
		return "USAGE_GODOT_STRING"
	case USAGE_GODOT_STRING_NAME:
		return "USAGE_GODOT_STRING_NAME"
	case USAGE_GDNATIVE_REF:
		return "USAGE_GDNATIVE_REF"
	case USAGE_GODOT_CONST_OR_ENUM:
		return "USAGE_GODOT_CONST_OR_ENUM"
	case USAGE_GODOT_CLASS:
		return "USAGE_GODOT_CLASS"
	default:
		log.Panicf("unknown usage value: %d", u)
	}

	return "___UNKNOWN"
}

func GoTypeAndUsage(value string) (string, Usage) {
	switch value {
	case "String":
		return "string", USAGE_GODOT_STRING
	case "StringName":
		return "string", USAGE_GODOT_STRING_NAME
	case "void":
		return "", USAGE_VOID
	case "bool":
		return "bool", USAGE_GO_PRIMATIVE
	case "int":
		return "int64", USAGE_GO_PRIMATIVE
	case "float":
		return "float32", USAGE_GO_PRIMATIVE
	case "double":
		return "float64", USAGE_GO_PRIMATIVE
	case "enum.Error":
		return "Error", USAGE_GDNATIVE_CONST_OR_ENUM
	case "AABB", "Array", "Basis", "Color", "Dictionary", "NodePath", "Plane",
		"PoolByteArray", "PoolColorArray", "PoolIntArray", "PoolRealArray",
		"PoolStringArray", "PoolVector2Array", "PoolVector3Array", "Quat", "Rect2",
		"RID", "Transform", "Transform2D", "Variant", "VariantType", "Vector2",
		"Vector3", "Vector3Axis":
		return value, USAGE_GDNATIVE_REF
	}

	matches := enumTypeRegex.FindAllStringSubmatch(value, 1)
	if len(matches) == 1 {
		tokens := matches[0]

		switch tokens[1] {
		case "AABB", "Array", "Basis", "Color", "Dictionary", "NodePath", "Plane",
			"PoolByteArray", "PoolColorArray", "PoolIntArray", "PoolRealArray",
			"PoolStringArray", "PoolVector2Array", "PoolVector3Array", "Quat", "Rect2",
			"RID", "Transform", "Transform2D", "Variant", "VariantType", "Vector2",
			"Vector3", "Vector3Axis":
			return fmt.Sprintf("%s%s", tokens[1], tokens[2]), USAGE_GODOT_CONST_OR_ENUM
		}

		return fmt.Sprintf("%s%s", tokens[1], tokens[2]), USAGE_GODOT_CONST_OR_ENUM
	}

	// TODO: whitelist the godot classes
	return fmt.Sprintf("%s", value), USAGE_GODOT_CLASS
}

// GoType converts types in the api.json into generated Go types.
func GoType(value string) string {
	t, _ := GoTypeAndUsage(value)
	return t
}

// GoTypeUsage converts types in the api.json into generated Go types.
func GoTypeUsage(value string) string {
	_, u := GoTypeAndUsage(value)
	return u.String()
}

// GDAPIDoc is a structure for parsed documentation.
type GDAPIDoc struct {
	Name        string        `xml:"name,attr"`
	Description string        `xml:"description"`
	Methods     []GDMethodDoc `xml:"methods>method"`
}

type GDMethodDoc struct {
	Name        string `xml:"name,attr"`
	Description string `xml:"description"`
}

type Constants map[string]int64

type GDAPIs []GDAPI

type ApiTypeBaseClassIndex map[string]map[string][]GDAPI

func (a GDAPIs) FilterForObject(index gdapiPathIndex) GDAPIs {
	arr := GDAPIs{}

	if index == nil {
		return arr
	}

	for _, api := range a {
		if !arrayContains(index[api.Name], "Object") && api.Name != "Object" {
			log.Printf("skipping class %s", api.Name)
			continue
		}

		arr = append(arr, api)
	}

	return arr
}

func (a GDAPIs) PartitionByBaseApiTypeAndClass() ApiTypeBaseClassIndex {
	parts := ApiTypeBaseClassIndex{
		"core":  map[string][]GDAPI{},
		"tools": map[string][]GDAPI{},
	}
	for _, api := range a {
		parentClasses, ok := parts[api.APIType]
		if !ok {
			log.Printf("api type '%v' unsupported", api.ApiType())
			continue
		}
		parentClasses[api.BaseClass] = append(parentClasses[api.BaseClass], api)
	}

	return parts
}

// GDAPI is a structure for parsed JSON from godot_api.json.
type GDAPI struct {
	APIType       string       `json:"api_type"`
	BaseClass     string       `json:"base_class"`
	Constants     Constants    `json:"constants"`
	Enums         []GDEnums    `json:"enums"`
	Methods       []GDMethod   `json:"methods"`
	Name          string       `json:"name"`
	Properties    []GDProperty `json:"properties"`
	Signals       []GDSignal   `json:"signals"`
	Singleton     bool         `json:"singleton"`
	SingletonName string       `json:"singleton_name"`
	Instanciable  bool         `json:"Instanciable"`
	IsReference   bool         `json:"is_reference"`
}

func (a GDAPI) HasEnumValue(value string) bool {
	for _, e := range a.Enums {
		for v := range e.Values {
			if v == value {
				return true
			}
		}
	}

	return false
}

func (a GDAPI) ParentInterface() string {
	if a.Name == "Object" {
		return "Class"
	}

	return a.BaseClass
}

func (a GDAPI) PrefixName() string {
	return strings.TrimLeft(a.Name, "_")
}

func (a GDAPI) ConstantPrefixName() string {
	return strings.ToUpper(casee.ToSnakeCase(strings.TrimLeft(a.Name, "_")))
}

func (a GDAPI) PrivatePrefixName() string {
	return casee.ToCamelCase(strings.TrimLeft(a.Name, "_"))
}

func (a GDAPI) ApiType() ApiType {
	return ParseApiType(a.APIType)
}

func (a GDAPI) GoMethods() []GDMethod {
	newMethods := make([]GDMethod, 0, len(a.Methods))
	for _, m := range a.Methods {
		// ignore methods with an underscore prefix
		if !strings.HasPrefix(m.Name, "_") && m.Name != "free" {
			newMethods = append(newMethods, m)
		}
	}

	return newMethods
}

// ByName is used for sorting GDAPI objects by name
type ByName []GDAPI

func (c ByName) Len() int           { return len(c) }
func (c ByName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByName) Less(i, j int) bool { return c[i].Name < c[j].Name }

type GDEnums struct {
	Name   string           `json:"name"`
	Values map[string]int64 `json:"values"`
}

// ByEnumName is used for sorting GDAPI objects by name
type ByEnumName []GDEnums

func (c ByEnumName) Len() int           { return len(c) }
func (c ByEnumName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByEnumName) Less(i, j int) bool { return c[i].Name < c[j].Name }

type GDMethod struct {
	Arguments    []GDArgument `json:"arguments"`
	HasVarargs   bool         `json:"has_varargs"`
	IsConst      bool         `json:"is_const"`
	IsEditor     bool         `json:"is_editor"`
	IsFromScript bool         `json:"is_from_script"`
	IsNoscript   bool         `json:"is_noscript"`
	IsReverse    bool         `json:"is_reverse"`
	IsVirtual    bool         `json:"is_virtual"`
	Name         string       `json:"name"`
	ReturnType   string       `json:"return_type"`
}

func (m GDMethod) CReturnType() string {
	return cType(m.ReturnType)
}

func (m GDMethod) GoName() string {
	return casee.ToPascalCase(m.Name)
}

func (m GDMethod) MethodBindName() string {
	return casee.ToPascalCase(m.Name) + "MethodBind"
}

// CToGoValueTypeMap proves a translation between C to Go
var CToGoValueTypeMap = map[string]string{
	"bool":               "bool",
	"uint8_t":            "uint8",
	"uint32_t":           "uint32",
	"uint64_t":           "uint64",
	"int64_t":            "int64",
	"double":             "float64",
	"wchar_t":            "int32",
	"char":               "C.char",
	"int":                "int32",
	"size_t":             "int64",
	"void":               "void",
	"string":             "string",
	"float":              "float32",
	"godot_real":         "float32",
	"godot_bool":         "bool",
	"godot_int":          "int32",
	"godot_object":       "GodotObject",
	"godot_vector3_axis": "Vector3Axis",
	"godot_variant_type": "VariantType",
	"godot_error":        "Error",
	"godot_string":       "String",
}

func (m GDMethod) IsBuiltinReturnType() bool {
	_, ok := CToGoValueTypeMap[m.ReturnType]
	return ok
}

// ByMethodName is used for sorting GDAPI objects by name
type ByMethodName []GDMethod

func (c ByMethodName) Len() int           { return len(c) }
func (c ByMethodName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByMethodName) Less(i, j int) bool { return c[i].Name < c[j].Name }

type GDArgument struct {
	DefaultValue    string `json:"default_value"`
	HasDefaultValue bool   `json:"has_default_value"`
	Name            string `json:"name"`
	Type            string `json:"type"`
}

func (m GDArgument) GoName() string {
	switch m.Name {
	case "var", "type", "interface", "default", "func", "range":
		return fmt.Sprintf("_%s", m.Name)
	}

	return m.Name
}

type GDProperty struct {
	Getter string `json:"getter"`
	Name   string `json:"name"`
	Setter string `json:"setter"`
	Type   string `json:"type"`
}

// ByPropertyName is used for sorting GDAPI objects by name
type ByPropertyName []GDProperty

func (c ByPropertyName) Len() int           { return len(c) }
func (c ByPropertyName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByPropertyName) Less(i, j int) bool { return c[i].Name < c[j].Name }

type GDSignal struct {
	Arguments []GDArgument `json:"arguments"`
	Name      string       `json:"name"`
}

// BySignalName is used for sorting GDAPI objects by name
type BySignalName []GDSignal

func (c BySignalName) Len() int           { return len(c) }
func (c BySignalName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c BySignalName) Less(i, j int) bool { return c[i].Name < c[j].Name }
