package types

// type TypedefTag int8

// const (
// 	enumTypedefTag TypedefTag = iota
// 	structTypedefTag
// 	boolTypedefTag
// 	intTypedefTag
// 	floatTypedefTag
// 	voidTypedefTag
// )

type GoProperty struct {
	Base            string       // Base will let us know if this is a struct, int, etc.
	Name            string       // The Go type name in camelCase
	CName           string       // The C type name in snake_case
	Comment         string       // Contains the comment on the line of the struct
	IsPointer       bool         // Usually for properties; defines if it is a pointer type
}

// TODO: Remove unneeded fields
type GoTypeDef struct {
	Base            string       // C Typedef tag
	Comment         string       // Contains the comment on the line of the struct
	Name            string       // The Go type name in camelCase
	CHeaderFilename string       // The header filename this type shows up in
	IsPointer       bool         // Usually for properties; defines if it is a pointer type
	CName           string       // The C type name in snake_case
	Properties      []GoProperty // Optional C struct fields
	IsBuiltIn       bool         // Whether or not the definition is just one line long (e.g. bool, int, etc.)
	Constructors    []GoMethod
	Methods         []GoMethod
}

type GoMethod struct {
	Name        string
	ReturnType  GoType
	Receiver    *GoArgument
	Arguments   []GoArgument
	CName       string
	ApiMetadata ApiMetadata
}

func (t *GoMethod) IsSetter() bool {
	return t.ReturnType.IsVoid() && !t.ReturnType.HasPointer && !t.ReturnType.HasDoublePointer
}

type GoArgument struct {
	Type GoType
	Name string
}

type GoType struct {
	HasConst         bool
	HasPointer       bool
	HasDoublePointer bool
	IsBuiltIn        bool
	Name             string
	CName            string
}

func (t *GoType) IsObject() bool {
	if t.CName == "godot_object" && !t.IsBuiltIn {
		panic("godot_object must be a built-in type")
	}
	return t.CName == "godot_object"
}

func (t *GoType) IsVoid() bool {
	return t.Name == "void"
}
