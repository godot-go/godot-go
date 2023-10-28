package builtin

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
	"github.com/godot-go/godot-go/cmd/gdextensionparser/clang"
	"github.com/iancoleman/strcase"
)

var (
	//go:embed builtinclasses.bindings.go.tmpl
	builtinClassesBindingsText string

	//go:embed builtinclasses.go.tmpl
	builtinClassesText string

	//go:embed variant.go.tmpl
	variantGoText string

	//go:embed classes.constants.go.tmpl
	classesConstantsText string

	//go:embed classes.enums.go.tmpl
	classesEnumsText string

	//go:embed classes.interfaces.go.tmpl
	classesInterfacesText string

	//go:embed classes.ref.interfaces.go.tmpl
	classesRefInterfacesText string
)

func Generate(projectPath string, ast clang.CHeaderFileAST, eapi extensionapiparser.ExtensionApi) {
	err := GenerateBuiltinClasses(projectPath, eapi)
	if err != nil {
		panic(err)
	}
	if err = GenerateBuiltinClassBindings(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassConstants(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassEnums(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassInterfaces(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassRefInterfaces(projectPath, eapi); err != nil {
		panic(err)
	}
	err = GenerateVariantGoFile(projectPath, ast)
	if err != nil {
		panic(err)
	}
}

func GenerateBuiltinClasses(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("builtinclasses.gen.go").
		Funcs(template.FuncMap{
			"contains":                 strings.Contains,
			"upper":                    strings.ToUpper,
			"lowerFirstChar":           lowerFirstChar,
			"upperFirstChar":           upperFirstChar,
			"lowerCamel":               strcase.ToLowerCamel,
			"screamingSnake":           screamingSnake,
			"snakeCase":                snakeCase,
			"goMethodName":             goMethodName,
			"goArgumentName":           goArgumentName,
			"goArgumentType":           goArgumentType,
			"goHasArgumentTypeEncoder": goHasArgumentTypeEncoder,
			"goReturnType":             goReturnType,
			"goDecodeNumberType":       goDecodeNumberType,
			"getOperatorIdName":        getOperatorIdName,
			"typeHasPtr":               typeHasPtr,
			"goEncoder":                goEncoder,
			"goEncodeIsReference":      goEncodeIsReference,
			"coalesce":                 coalesce,
		}).
		Parse(builtinClassesText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "builtin", fmt.Sprintf("builtinclasses.gen.go"))

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

func GenerateBuiltinClassBindings(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("builtinclasses.bindings.gen.go").
		Funcs(template.FuncMap{
			"contains":                 strings.Contains,
			"upper":                    strings.ToUpper,
			"lowerFirstChar":           lowerFirstChar,
			"upperFirstChar":           upperFirstChar,
			"lowerCamel":               strcase.ToLowerCamel,
			"screamingSnake":           screamingSnake,
			"snakeCase":                snakeCase,
			"goMethodName":             goMethodName,
			"goArgumentName":           goArgumentName,
			"goArgumentType":           goArgumentType,
			"goHasArgumentTypeEncoder": goHasArgumentTypeEncoder,
			"goReturnType":             goReturnType,
			"goDecodeNumberType":       goDecodeNumberType,
			"getOperatorIdName":        getOperatorIdName,
			"typeHasPtr":               typeHasPtr,
			"goEncoder":                goEncoder,
			"goEncodeIsReference":      goEncodeIsReference,
			"coalesce":                 coalesce,
		}).
		Parse(builtinClassesBindingsText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "builtin", fmt.Sprintf("builtinclasses.bindings.gen.go"))

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

func GenerateClassConstants(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.constants.gen.go").
		Funcs(template.FuncMap{
			"goVariantConstructor": goVariantConstructor,
			"goMethodName":         goMethodName,
			"goArgumentName":       goArgumentName,
			"goArgumentType":       goArgumentType,
			"goReturnType":         goReturnType,
			"goClassEnumName":      goClassEnumName,
			"goClassConstantName":  goClassConstantName,
			"goClassStructName":    goClassStructName,
			"goClassInterfaceName": goClassInterfaceName,
			"coalesce":             coalesce,
		}).
		Parse(classesConstantsText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "builtin", fmt.Sprintf("classes.constants.gen.go"))

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

func GenerateClassEnums(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.enums.gen.go").
		Funcs(template.FuncMap{
			"goVariantConstructor": goVariantConstructor,
			"goMethodName":         goMethodName,
			"goArgumentName":       goArgumentName,
			"goArgumentType":       goArgumentType,
			"goReturnType":         goReturnType,
			"goClassEnumName":      goClassEnumName,
			"goClassStructName":    goClassStructName,
			"goClassInterfaceName": goClassInterfaceName,
			"coalesce":             coalesce,
		}).
		Parse(classesEnumsText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "builtin", fmt.Sprintf("classes.enums.gen.go"))

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

func GenerateClassInterfaces(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.interfaces.gen.go").
		Funcs(template.FuncMap{
			"goVariantConstructor": goVariantConstructor,
			"goMethodName":         goMethodName,
			"goArgumentName":       goArgumentName,
			"goArgumentType":       goArgumentType,
			"goReturnType":         goReturnType,
			"goClassEnumName":      goClassEnumName,
			"goClassStructName":    goClassStructName,
			"goClassInterfaceName": goClassInterfaceName,
			"coalesce":             coalesce,
		}).
		Parse(classesInterfacesText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "builtin", fmt.Sprintf("classes.interfaces.gen.go"))

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

func GenerateClassRefInterfaces(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.ref.instances.gen.go").
	Funcs(template.FuncMap{
		"isSetterMethodName":   isSetterMethodName,
		"goVariantConstructor": goVariantConstructor,
		"goMethodName":         goMethodName,
		"goArgumentName":       goArgumentName,
		"goArgumentType":       goArgumentType,
		"goVariantFunc":        goVariantFunc,
		"goReturnType":         goReturnType,
		"goClassEnumName":      goClassEnumName,
		"goClassStructName":    goClassStructName,
		"goClassInterfaceName": goClassInterfaceName,
		"goEncoder":            goEncoder,
		"goEncodeIsReference":  goEncodeIsReference,
		"coalesce":             coalesce,
	}).
		Parse(classesRefInterfacesText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "builtin", fmt.Sprintf("classes.ref.interfaces.gen.go"))

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

func GenerateVariantGoFile(projectPath string, ast clang.CHeaderFileAST) error {
	funcs := template.FuncMap{
		"snakeCase":           strcase.ToSnake,
		"camelCase":           strcase.ToCamel,
	}

	tmpl, err := template.New("variant.gen.go").
		Funcs(funcs).
		Parse(variantGoText)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, ast)
	if err != nil {
		return err
	}

	goFileName := filepath.Join(projectPath, "pkg", "builtin", "variant.gen.go")
	f, err := os.Create(goFileName)
	f.Write(b.Bytes())
	f.Close()
	return nil
}