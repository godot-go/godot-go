package clang

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	. "github.com/alecthomas/participle/v2/lexer"
)

type CHeaderFileAST struct {
	Expr []Expr `parser:" @@* "`
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
	Comment  string           `parser:"   @Comment "`
	Enum     *TypedefEnum     `parser:" | @@ ';'   "`
	Alias    *TypedefAlias    `parser:" | @@ ';'   "`
	Function *TypedefFunction `parser:" | @@ ';'   "`
	Struct   *TypedefStruct   `parser:" | @@ ';'   "`
}

type TypedefEnum struct {
	Values []EnumValue `parser:" 'typedef' 'enum' '{' ( @@ ( ',' Comment? @@ Comment? )* ','? Comment? )? '}' "`
	Name   *string     `parser:" @Ident                                                              "`
}

type EnumValue struct {
	Name          string  `parser:" @Ident                     "`
	IntValue      *int    `parser:" ( '=' ( @Int               "`
	ConstRefValue *string `parser:"              | @Ident ) )? "`
}

type TypedefAlias struct {
	Type Type   `parser:" 'typedef' @@ "`
	Name string `parser:" @Ident       "`
}

type TypedefFunction struct {
	ReturnType Type       `parser:" 'typedef' @@                "`
	Name       string     `parser:" '(' '*' @Ident ')'          "`
	Arguments  []Argument `parser:" '(' ( @@ ( ',' @@ )* )? ')' "`
}

type TypedefStruct struct {
	Fields []StructField `parser:" 'typedef' 'struct' '{' @@* '}' "`
	Name   string        `parser:" @Ident                         "`
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
	Variable *StructVariable `parser:" ( @@       "`
	Function *StructFunction `parser:" | @@ ) ';' "`
}

type StructVariable struct {
	Type Type   `parser:" @@     "`
	Name string `parser:" @Ident "`
}

type Type struct {
	IsConst   bool   `parser:" @'const'? "`
	Name      string `parser:" @Ident    "`
	IsPointer bool   `parser:" @'*'?     "`
}

func (t Type) String() string {
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

type StructFunction struct {
	ReturnType Type       `parser:" @@                     "`
	Name       string     `parser:" '(' '*' @Ident ')'     "`
	Arguments  []Argument `parser:" '(' @@ ( ',' @@ )* ')' "`
	Comment    string     `parser:" @Comment?              "`
}

type Argument struct {
	Type Type   `parser:" @@     "`
	Name string `parser:" @Ident? "`
}

func ParseCString(s string) (CHeaderFileAST, error) {
	var (
		f CHeaderFileAST
	)

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

	parser, err := participle.Build(
		&CHeaderFileAST{},
		participle.Lexer(headerFileLexer),
		participle.UseLookahead(4),
		participle.Elide("Whitespace", "Comment"),
	)

	if err != nil {
		return f, err
	}

	err = parser.ParseString("", s, &f)

	if err != nil {
		return f, err
	}

	return f, nil
}
