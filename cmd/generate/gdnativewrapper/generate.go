// Package gdnativewrapper generates C code to wrap all of the gdnative
// methods to call functions on the gdnative_api_structs to work
// around the cgo C function pointer limitation.
package gdnativewrapper

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/godot-go/godot-go/cmd/gdnativeparser/clang"
	"github.com/godot-go/godot-go/cmd/gdnativeparser/preprocessor"
	"github.com/iancoleman/strcase"
)

var (
	//go:embed gdnative_wrapper.h.tmpl
	gdnativeWrapperHeaderFileText string

	//go:embed gdnative_wrapper.c.tmpl
	gdnativeWrapperSrcFileText string

	//go:embed gdnative_wrapper.go.tmpl
	gdnativeWrapperGoFileText string
)

func Generate(projectPath string) {
	ast, err := generateGDExtensionInterfaceAST(projectPath)

	if err != nil {
		panic(err)
	}

	err = GenerateGDExtensionWrapperHeaderFile(projectPath, ast)

	if err != nil {
		panic(err)
	}

	err = GenerateGDExtensionWrapperSrcFile(projectPath, ast)

	if err != nil {
		panic(err)
	}

	err = GenerateGDExtensionWrapperGoFile(projectPath, ast)

	if err != nil {
		panic(err)
	}
}

func GenerateGDExtensionWrapperHeaderFile(projectPath string, ast clang.CHeaderFileAST) error {
	tmpl, err := template.New("gdnative_wrapper.gen.h").
		Funcs(template.FuncMap{
			"snakeCase": strcase.ToSnake,
		}).
		Parse(gdnativeWrapperHeaderFileText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, ast)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdnative", "gdnative_wrapper.gen.h")

	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(b.Bytes())

	if err != nil {
		return err
	}

	return nil
}

func GenerateGDExtensionWrapperSrcFile(projectPath string, ast clang.CHeaderFileAST) error {
	tmpl, err := template.New("gdnative_wrapper.gen.c").
		Funcs(template.FuncMap{
			"snakeCase": strcase.ToSnake,
		}).
		Parse(gdnativeWrapperSrcFileText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, ast)

	if err != nil {
		return err
	}

	headerFileName := filepath.Join(projectPath, "pkg", "gdnative", "gdnative_wrapper.gen.c")

	f, err := os.Create(headerFileName)
	f.Write(b.Bytes())
	f.Close()

	return nil
}

func GenerateGDExtensionWrapperGoFile(projectPath string, ast clang.CHeaderFileAST) error {
	tmpl, err := template.New("gdnative_wrapper.gen.go").
		Funcs(template.FuncMap{
			"snakeCase":          strcase.ToSnake,
			"camelCase":          strcase.ToCamel,
			"goReturnType":       goReturnType,
			"goArgumentType":     goArgumentType,
			"goEnumValue":        goEnumValue,
			"add":                add,
			"cgoCastArgument":    cgoCastArgument,
			"cgoCastReturnType":  cgoCastReturnType,
			"cgoCleanUpArgument": cgoCleanUpArgument,
		}).
		Parse(gdnativeWrapperGoFileText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, ast)

	if err != nil {
		return err
	}

	headerFileName := filepath.Join(projectPath, "pkg", "gdnative", "gdnative_wrapper.gen.go")

	f, err := os.Create(headerFileName)
	f.Write(b.Bytes())
	f.Close()

	return nil
}

func generateGDExtensionInterfaceAST(projectPath string) (clang.CHeaderFileAST, error) {
	n := filepath.Join(projectPath, "/godot_headers/godot/gdextension_interface.h")
	b, err := os.ReadFile(n)

	if err != nil {
		return clang.CHeaderFileAST{}, fmt.Errorf("error reading %s: %w", n, err)
	}

	preprocFile, err := preprocessor.ParsePreprocessorString((string)(b))

	if err != nil {
		return clang.CHeaderFileAST{}, fmt.Errorf("error preprocessing %s: %w", n, err)
	}

	preprocText := preprocFile.Eval(false)

	ast, err := clang.ParseCString(preprocText)

	if err != nil {
		return clang.CHeaderFileAST{}, fmt.Errorf("error parsing %s: %w", n, err)
	}

	return ast, nil
}
