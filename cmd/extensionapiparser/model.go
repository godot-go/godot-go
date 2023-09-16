package extensionapiparser

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

type Header struct {
	VersionMajor    int    `json:"version_major"`
	VersionMinor    int    `json:"version_minor"`
	VersionPatch    int    `json:"version_patch"`
	VersionStatus   string `json:"version_status"`
	VersionBuild    string `json:"version_build"`
	VersionFullName string `json:"version_full_name"`
}

type BuiltinClassSizeDetail struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

type BuiltinClassSize struct {
	BuildConfiguration string                   `json:"build_configuration"`
	Sizes              []BuiltinClassSizeDetail `json:"sizes"`
}

func (sz BuiltinClassSize) FindSize(name string) int {
	for _, sz := range sz.Sizes {
		if sz.Name == name {
			return sz.Size
		}
	}

	panic(fmt.Sprintf("could not find size for %s", name))
}

type BuiltinClassMemberOffsetClassMember struct {
	Member string `json:"member"`
	Offset int    `json:"offset"`
	Meta   string `json:"meta"`
}

type BuiltinClassMemberOffsetClass struct {
	Name    string                                `json:"name"`
	Members []BuiltinClassMemberOffsetClassMember `json:"members"`
}

type BuiltinClassMemberOffset struct {
	BuildConfiguration string                          `json:"build_configuration"`
	Classes            []BuiltinClassMemberOffsetClass `json:"classes"`
}

type GlobalConstant struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type EnumValue struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type Enum struct {
	Name       string      `json:"name"`
	IsBitfield *bool       `json:"is_bitfield"`
	Values     []EnumValue `json:"values"`
}

func (e Enum) GoName() string {
	return strings.Replace(e.Name, ".", "", -1)
}

type Argument struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DefaultValue string `json:"default_value"`
	Meta         string `json:"meta"`
}

func (a Argument) HasDestroy() bool {
	switch a.Type {
	case "String", "StringName":
		return true
	}

	return false
}

type UtilityFunction struct {
	Name              string     `json:"name"`
	ReturnType        string     `json:"return_type"`
	Category          string     `json:"category"`
	IsVararg          bool       `json:"is_vararg"`
	Hash              int        `json:"hash"`
	HashCompatibility []int      `json:"hash_compatibility"`
	Arguments         []Argument `json:"arguments"`
}

type ClassOperator struct {
	Name       string `json:"name"`
	RightType  string `json:"right_type"`
	ReturnType string `json:"return_type"`
}

type ClassConstructor struct {
	Index     int        `json:"index"`
	Arguments []Argument `json:"arguments"`
	Name      string     `json:"name"`
}

type BuiltInClassMethod struct {
	Name              string     `json:"name"`
	ReturnType        string     `json:"return_type"`
	IsConst           bool       `json:"is_const"`
	IsVararg          bool       `json:"is_vararg"`
	IsStatic          bool       `json:"is_static"`
	Hash              int        `json:"hash"`
	HashCompatibility []int      `json:"hash_compatibility"`
	Arguments         []Argument `json:"arguments"`
}

type ClassMember struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type BuiltInClassConstant struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ClassConstant struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type BuiltinClass struct {
	Name               string                 `json:"name"`
	IndexingReturnType string                 `json:"indexing_return_type"`
	IsKeyed            bool                   `json:"is_keyed"`
	Members            []ClassMember          `json:"members"`
	Constants          []BuiltInClassConstant `json:"constants"`
	Enums              []Enum                 `json:"enums"`
	Operators          []ClassOperator        `json:"operators"`
	Methods            []BuiltInClassMethod   `json:"methods"`
	Constructors       []ClassConstructor     `json:"constructors"`
	HasDestructor      bool                   `json:"has_destructor"`
}

func (a BuiltinClass) FilteredConstructors() []ClassConstructor {
	switch a.Name {
	case "String":
		values := make([]ClassConstructor, 0, len(a.Constructors))

		for _, c := range a.Constructors {
			if len(c.Arguments) == 1 && c.Arguments[0].Type == "String" {
				continue
			}

			values = append(values, c)
		}

		return values
	default:
		return a.Constructors
	}
}

func (a BuiltinClass) FilteredMethods() []BuiltInClassMethod {
	switch a.Name {
	case "Signal":
		excludes := []string{
			"emit",
		}

		values := make([]BuiltInClassMethod, 0, len(a.Methods))

		for _, m := range a.Methods {
			if slices.Contains(excludes, m.Name) {
				continue
			}

			values = append(values, m)
		}

		return values
	case "Callable":
		excludes := []string{
			"call",
			"call_deferred",
			"rpc",
			"rpc_id",
			"bind",
		}

		values := make([]BuiltInClassMethod, 0, len(a.Methods))

		for _, m := range a.Methods {
			if slices.Contains(excludes, m.Name) {
				continue
			}

			values = append(values, m)
		}

		return values
	default:
		return a.Methods
	}
}

type ClassMethodReturnValue struct {
	Type string `json:"type"`
	Meta string `json:"meta"`
}

type ClassMethod struct {
	Name              string                 `json:"name"`
	ReturnValue       ClassMethodReturnValue `json:"return_value"`
	IsConst           bool                   `json:"is_const"`
	IsVararg          bool                   `json:"is_vararg"`
	IsVirtual         bool                   `json:"is_virtual"`
	IsStatic          bool                   `json:"is_static"`
	Hash              int                    `json:"hash"`
	HashCompatibility []int                  `json:"hash_compatibility"`
	Arguments         []Argument             `json:"arguments"`
}

type ClassSignal struct {
	Name      string     `json:"name"`
	Arguments []Argument `json:"arguments"`
}

type ClassProperty struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Setter string `json:"setter"`
	Getter string `json:"getter"`
	Index  int    `json:"index"`
}

type Class struct {
	Name           string          `json:"name"`
	IsRefcounted   bool            `json:"is_refcounted"`
	IsInstantiable bool            `json:"is_instantiable"`
	Inherits       string          `json:"inherits"`
	ApiType        string          `json:"api_type"`
	Constants      []ClassConstant `json:"constants"`
	Enums          []Enum          `json:"enums"`
	Methods        []ClassMethod   `json:"methods"`
	Signals        []ClassSignal   `json:"signals"`
	Properties     []ClassProperty `json:"properties"`
}

func (a Class) FilteredEnums() []Enum {
	values := make([]Enum, 0, len(a.Enums))

	for _, e := range a.Enums {
		switch fmt.Sprintf("%s%s", a.Name, e.GoName()) {
		case "GDExtensionInitializationLevel":
			continue
		default:
			values = append(values, e)
		}
	}

	return values
}

type Singleton struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type NativeStructure struct {
	Name   string `json:"name"`
	Format string `json:"format"`
}

type ExtensionApi struct {
	Header                    Header                     `json:"header"`
	BuiltinClassSizes         []BuiltinClassSize         `json:"builtin_class_sizes"`
	BuiltinClassMemberOffsets []BuiltinClassMemberOffset `json:"builtin_class_member_offsets"`
	GlobalConstants           []GlobalConstant           `json:"global_constants"`
	GlobalEnums               []Enum                     `json:"global_enums"`
	UtilityFunctions          []UtilityFunction          `json:"utility_functions"`
	BuiltinClasses            []BuiltinClass             `json:"builtin_classes"`
	Classes                   []Class                    `json:"classes"`
	Singletons                []Singleton                `json:"singletons"`
	NativeStructures          []NativeStructure          `json:"native_structures"`
	BuildConfig               string                     `json:"-"`
}

func (a ExtensionApi) HasBuildConfiguration(buildConfig string) bool {
	for _, sz := range a.BuiltinClassSizes {
		if sz.BuildConfiguration == buildConfig {
			return true
		}
	}
	return false
}

func (a ExtensionApi) Float64BuiltinClassSize() *BuiltinClassSize {
	for _, sz := range a.BuiltinClassSizes {
		if sz.BuildConfiguration == a.BuildConfig {
			return &sz
		}
	}
	return nil
}

func (a ExtensionApi) ContainsClassName(name string) bool {
	for _, c := range a.Classes {
		if c.Name == name {
			return true
		}
	}

	return false
}

func (a ExtensionApi) IsRefcountedClassName(name string) bool {
	for _, c := range a.Classes {
		if c.Name == name {
			return c.IsRefcounted
		}
	}

	return false
}

func (a ExtensionApi) FilteredClasses() []Class {
	values := make([]Class, 0, len(a.Classes))

	for _, c := range a.Classes {
		values = append(values, c)
	}

	return values
}

func (a ExtensionApi) FilteredRefcountedClasses() []Class {
	values := make([]Class, 0, len(a.Classes))

	for _, c := range a.Classes {
		if c.IsRefcounted {
			values = append(values, c)
		}
	}

	return values
}

func (a ExtensionApi) FilteredBuiltinClasses() []BuiltinClass {
	values := make([]BuiltinClass, 0, len(a.BuiltinClasses))

	for _, c := range a.BuiltinClasses {
		switch c.Name {
		case
			"Nil",
			"int",
			"float",
			"bool":
			continue
		default:
			values = append(values, c)
		}
	}

	return values
}
