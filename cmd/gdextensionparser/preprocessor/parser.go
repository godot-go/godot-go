package preprocessor

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/participle/v2"
	. "github.com/alecthomas/participle/v2/lexer"
)

type PreprocVars map[string]struct{}

type PreprocessorHeaderFileAST struct {
	Directives []Directive `parser:" @@* "`
}

func (f PreprocessorHeaderFileAST) Eval(isCpp bool) string {
	vars := PreprocVars{}

	if isCpp {
		vars["__cplusplus"] = struct{}{}
	}

	return f.eval(vars)
}

func (f PreprocessorHeaderFileAST) eval(vars PreprocVars) string {
	sb := strings.Builder{}

	for _, d := range f.Directives {
		sb.WriteString(d.Eval(vars))
		sb.WriteString("\n")
	}

	return sb.String()
}

type Directive struct {
	Ifndef  *IfndefDirective  `parser:" EOL* ( @@          "`
	Ifdef   *IfdefDirective   `parser:"   | @@        "`
	Define  *DefineDirective  `parser:"   | @@        "`
	Include *IncludeDirective `parser:"   | @@        "`
	Source  string            `parser:"   | @Source ) "`
}

func (d Directive) Eval(vars PreprocVars) string {
	if d.Ifndef != nil {
		return d.Ifndef.Eval(vars)
	} else if d.Ifdef != nil {
		return d.Ifdef.Eval(vars)
	} else if d.Define != nil {
		return d.Define.Eval(vars)
	} else if d.Include != nil {
		return d.Include.Eval(vars)
	} else {
		return d.Source
	}
}

type IfndefDirective struct {
	Name       string      `parser:" @Ifndef EOL      "`
	Directives []Directive `parser:" @@* '#endif' EOL "`
}

func (d IfndefDirective) Eval(vars PreprocVars) string {
	if len(d.Name) == 0 {
		panic("#ifndef missing variable")
	}

	sb := strings.Builder{}

	if _, ok := vars[d.Name]; !ok {
		for _, c := range d.Directives {
			sb.WriteString(c.Eval(vars))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

type IfdefDirective struct {
	Name       string      `parser:" @Ifdef EOL       "`
	Directives []Directive `parser:" @@* '#endif' EOL "`
}

func (d IfdefDirective) Eval(vars PreprocVars) string {
	if len(d.Name) == 0 {
		panic("#ifdef missing variable")
	}

	sb := strings.Builder{}

	if _, ok := vars[d.Name]; ok {
		for _, c := range d.Directives {
			sb.WriteString(c.Eval(vars))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

type DefineDirective struct {
	Name string `parser:" @Define EOL" `
}

func (d DefineDirective) Eval(vars PreprocVars) string {
	if len(d.Name) == 0 {
		panic("#define missing variable")
	}

	if _, ok := vars[d.Name]; !ok {
		vars[d.Name] = struct{}{}
	}

	return ""
}

type IncludeDirective struct {
	Name string `parser:" @Include EOL" `
}

func (d IncludeDirective) Eval(vars PreprocVars) string {
	return ""
}

func ParsePreprocessorString(s string) (*PreprocessorHeaderFileAST, error) {
	var preprocessorHeaderFileLexer = MustStateful(Rules{
		"Root": {
			{`Ifdef`, `#ifdef[ \t]+[a-zA-Z_][a-zA-Z0-9_]*`, Push("Root")},
			{`Ifndef`, `#ifndef[ \t]+[a-zA-Z_][a-zA-Z0-9_]*`, Push("Root")},
			{`Define`, `#define[ \t]+[a-zA-Z_][a-zA-Z0-9_]*`, nil},
			{`Include`, `#include[ \t]+<[A-Za-z0-9_]+\.h>`, nil},
			{`Endif`, `#endif`, Pop()},
			{`Whitespace`, `[ \t]+`, nil},
			{`Comment`, `(\/\/[^\n]*)|(\/\*(.|[\r\n])*?\*\/)`, nil},
			{`EOL`, `[\n\r]+`, nil},
			{`Source`, `[\n\r]*[^#]+[\n\r]*`, nil},
		},
	})

	parser, err := participle.Build[PreprocessorHeaderFileAST](
		participle.Lexer(preprocessorHeaderFileLexer),
		participle.Elide("Whitespace", "Comment"),
		directiveIdent("Ifdef", "Ifndef", "Define"),
		directiveFilename("Include"),
	)

	if err != nil {
		return nil, err
	}

	ast, err := parser.ParseString("", s)

	if err != nil {
		return nil, err
	}

	return ast, nil
}

func directiveIdent(types ...string) participle.Option {
	re := regexp.MustCompile("(#[A-Za-z_]+)[ \t]+([a-zA-Z_][a-zA-Z0-9_]*)")
	return participle.Map(func(token Token) (Token, error) {
		matches := re.FindAllStringSubmatch(token.Value, 1)

		if len(matches) == 0 {
			panic(fmt.Sprintf("no matches found for DirectiveIdent: %v", token))
		}

		token.Value = matches[0][2]

		return token, nil
	}, types...)
}

func directiveFilename(types ...string) participle.Option {
	re := regexp.MustCompile("(#[A-Za-z_]+)[ \t]+<([a-zA-Z_][a-zA-Z0-9_.]*)>")
	return participle.Map(func(token Token) (Token, error) {
		matches := re.FindAllStringSubmatch(token.Value, 1)

		if len(matches) == 0 {
			panic(fmt.Sprintf("no matches found for DirectiveFilename: %v", token))
		}

		token.Value = matches[0][2]

		return token, nil
	}, types...)
}
