package clang

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	. "github.com/alecthomas/participle/v2/lexer"
	"golang.org/x/exp/slices"
)

var (
	legacyGDExtentionInterfaceFunctionNames []string = []string {
		"GDExtensionInterfaceFunctionPtr",
	}
)

type CHeaderFileAST struct {
	Expr []Expr `parser:" @@* " json:",omitempty"`
}

func (a CHeaderFileAST) FindVariantEnumType() *TypedefEnum {
	for _, e := range a.Expr {
		if e.Enum != nil && e.Enum.Name != nil && *e.Enum.Name == "GDExtensionVariantType" {
			return e.Enum
		}
	}
	return nil
}

func (a CHeaderFileAST) CollectFunctions() []TypedefFunction {
	// there's a duplicate of GDExtensionClassGetPropertyList
	distinct := map[string]struct{}{}
	fns := make([]TypedefFunction, 0, len(a.Expr))

	for _, e := range a.Expr {
		if e.Function != nil {
			if _, ok := distinct[e.Function.Name]; !ok {
				fns = append(fns, *e.Function)
				distinct[e.Function.Name] = struct{}{}
			}
		}
	}
	return fns
}

func (a CHeaderFileAST) CollectGDExtensionInterfaceFunctions() []TypedefFunction {
	allFns := a.CollectFunctions()

	fns := make([]TypedefFunction, 0, len(allFns))

	for _, fn := range allFns {
		if strings.HasPrefix(fn.Name, "GDExtensionInterface") &&
			!slices.Contains(legacyGDExtentionInterfaceFunctionNames, fn.Name) {
			fns = append(fns, fn)
		}
	}

	return fns
}

func (a CHeaderFileAST) CollectNonGDExtensionInterfaceFunctions() []TypedefFunction {
	allFns := a.CollectFunctions()

	fns := make([]TypedefFunction, 0, len(allFns))

	for _, fn := range allFns {
		if !strings.HasPrefix(fn.Name, "GDExtensionInterface") {
			fns = append(fns, fn)
		}
	}

	return fns
}

func (a CHeaderFileAST) CollectStructs() []TypedefStruct {
	structs := make([]TypedefStruct, 0, len(a.Expr))
	for _, e := range a.Expr {
		if e.Struct != nil {
			structs = append(structs, *e.Struct)
		}
	}
	return structs
}

func (a CHeaderFileAST) CollectAliases() []TypedefAlias {
	aliases := make([]TypedefAlias, 0, len(a.Expr))
	for _, e := range a.Expr {
		if e.Alias != nil {
			aliases = append(aliases, *e.Alias)
		}
	}
	return aliases
}

func (a CHeaderFileAST) CollectEnums() []TypedefEnum {
	enums := make([]TypedefEnum, 0, len(a.Expr))
	for _, e := range a.Expr {
		if e.Enum != nil {
			enums = append(enums, *e.Enum)
		}
	}
	return enums
}

type Expr struct {
	Comment  string           `parser:"   @Comment " json:",omitempty"`
	Enum     *TypedefEnum     `parser:" | @@ ';'   " json:",omitempty"`
	Alias    *TypedefAlias    `parser:" | @@ ';'   " json:",omitempty"`
	Function *TypedefFunction `parser:" | @@ ';'   " json:",omitempty"`
	Struct   *TypedefStruct   `parser:" | @@ ';'   " json:",omitempty"`
}

type TypedefEnum struct {
	Values []EnumValue `parser:" 'typedef' 'enum' '{' ( @@ ( ',' Comment? @@ Comment? )* ','? Comment? )? '}' " json:",omitempty"`
	Name   *string     `parser:" @Ident                                                                       " json:",omitempty"`
}

type EnumValue struct {
	Name          string  `parser:" @Ident                     " json:",omitempty"`
	IntValue      *int    `parser:" ( '=' ( @Int               " json:",omitempty"`
	ConstRefValue *string `parser:"              | @Ident ) )? " json:",omitempty"`
}

type TypedefAlias struct {
	Type PrimativeType `parser:" 'typedef' @@ " json:",omitempty"`
	Name string        `parser:" @Ident       " json:",omitempty"`
}

type TypedefFunction struct {
	ReturnType PrimativeType `parser:" 'typedef' @@                " json:",omitempty"`
	Name       string        `parser:" '(' '*' @Ident ')'          " json:",omitempty"`
	Arguments  []Argument    `parser:" '(' ( @@ ( ',' @@ )* )? ')' " json:",omitempty"`
}

type TypedefStruct struct {
	Fields []StructField `parser:" 'typedef' 'struct' '{' @@* '}' " json:",omitempty"`
	Name   string        `parser:" @Ident                         " json:",omitempty"`
}

func (t TypedefStruct) CollectFunctions() []StructFunction {
	fns := make([]StructFunction, 0, len(t.Fields))
	for _, f := range t.Fields {
		if f.Function != nil {
			fns = append(fns, *f.Function)
		}
	}
	return fns
}

type StructField struct {
	Variable *StructVariable `parser:" ( @@       " json:",omitempty"`
	Function *StructFunction `parser:" | @@ ) ';' " json:",omitempty"`
}

type StructVariable struct {
	Type PrimativeType `parser:" @@     " json:",omitempty"`
	Name string        `parser:" @Ident " json:",omitempty"`
}

type FunctionType struct {
	ReturnType PrimativeType `parser:" @@                          " json:",omitempty"`
	Name       string        `parser:" '(' '*' @Ident ')'          " json:",omitempty"`
	Arguments  []Argument    `parser:" '(' ( @@ ( ',' @@ )* )? ')' " json:",omitempty"`
}

func (t FunctionType) CStyleString() string {
	sb := strings.Builder{}
	sb.WriteString(t.ReturnType.CStyleString())
	sb.WriteString("(*")
	sb.WriteString(t.Name)
	sb.WriteString(")(")
	for i := 0; i < len(t.Arguments); i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(t.Arguments[i].Type.Primative.CStyleString())
	}
	sb.WriteString(")")
	return sb.String()
}

type PrimativeType struct {
	IsConst   bool   `parser:" @'const'? " json:",omitempty"`
	Name      string `parser:" @Ident    " json:",omitempty"`
	IsPointer bool   `parser:" @'*'?     " json:",omitempty"`
}

func (t PrimativeType) CStyleString() string {
	sb := strings.Builder{}

	if t.IsConst {
		sb.WriteString("const ")
	}

	sb.WriteString(t.Name)

	if t.IsPointer {
		sb.WriteString(" * ")
	}

	return sb.String()
}

type Type struct {
	Function  *FunctionType  `parser:" ( @@   " json:",omitempty"`
	Primative *PrimativeType `parser:" | @@ ) " json:",omitempty"`
}

func (t Type) CStyleString() string {
	if t.Primative != nil {
		return t.Primative.CStyleString()
	} else if t.Function != nil {
		return t.Function.CStyleString()
	}

	panic("unhandled type")
}

type StructFunction struct {
	ReturnType PrimativeType `parser:" @@                     " json:",omitempty"`
	Name       string        `parser:" '(' '*' @Ident ')'     " json:",omitempty"`
	Arguments  []Argument    `parser:" '(' @@ ( ',' @@ )* ')' " json:",omitempty"`
	Comment    string        `parser:" @Comment?              " json:",omitempty"`
}

// void (*p_func)(void *, uint32_t)
type Argument struct {
	Type      Type       `parser:" @@                               " json:",omitempty"`
	Name      string     `parser:" ( @Ident | '(' '*' @Ident ')' )? " json:",omitempty"`
}

func (a Argument) IsPinnable() bool {
	switch {
	case a.Type.Function != nil:
		return false
	case a.Type.Primative != nil:
		switch a.Type.Primative.Name {
		case "char":
			return false
		default:
			return a.Type.Primative.IsPointer
		}
	}

	return false
}

func (a Argument) CStyleString(i int) string {
	if a.Type.Function != nil {
		return a.Type.CStyleString()
	}

	name := a.ResolvedName(i)

	return fmt.Sprintf("%s %s", a.Type.CStyleString(), name)
}

func (a Argument) ResolvedName(i int) string {
	if a.Type.Function != nil && a.Type.Function.Name != "" {
		return a.Type.Function.Name
	}

	if a.Name != "" {
		return a.Name
	}
	return fmt.Sprintf("arg_%d", i)
}

func ParseCString(s string) (CHeaderFileAST, error) {
	var headerFileLexer = MustStateful(Rules{
		"Root": {
			{`Typedef`, `typedef`, nil},
			{`Struct`, `struct`, nil},
			{`{`, `{`, nil},
			{`}`, `}`, nil},
			{`;`, `;`, nil},
			{`,`, `,`, nil},
			{`"`, `"`, nil},
			{`(`, `\(`, nil},
			{`)`, `\)`, nil},
			{`*`, `\*`, nil},
			{`=`, `=`, nil},
			{`Const`, `const`, nil},
			{`Ident`, `[a-zA-Z_][a-zA-Z0-9_]*`, nil},
			{`Int`, `[+-]?\d+`, nil},
			{`Comment`, `[ \t\r\n]*(\/\/[^\n]*)|(\/\*(.|[\r\n])*?\*\/)[ \t\r\n]*`, nil},
			{`Whitespace`, `[ \t\r\n]+`, nil},
		},
	})

	parser, err := participle.Build[CHeaderFileAST](
		participle.Lexer(headerFileLexer),
		participle.UseLookahead(20),
		participle.Elide("Whitespace", "Comment"),
	)

	if err != nil {
		return CHeaderFileAST{}, err
	}

	ast, err := parser.ParseString("", s)

	if err != nil {
		return CHeaderFileAST{}, err
	}

	return *ast, nil
}
