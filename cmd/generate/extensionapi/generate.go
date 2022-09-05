// Package types is responsible for parsing the Godot headers for type definitions
// and generating Go wrappers around that structure.
package extensionapi

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	_ "embed"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
	"github.com/iancoleman/strcase"
)

var (
	//go:embed globalconstants.go.tmpl
	globalConstantsText string

	//go:embed globalenums.go.tmpl
	globalEnumsText string

	//go:embed utilityfunctions.go.tmpl
	utilityFunctionsText string

	//go:embed builtinclasses.go.tmpl
	builtinClassesText string

	//go:embed classes.go.tmpl
	classesText string

	//go:embed classes.h.tmpl
	cHeaderClassesText string

	//go:embed classes.c.tmpl
	cClassesText string

	//go:embed nativestructures.go.tmpl
	nativeStructuresText string
)

// Generate will generate Go wrappers for all Godot base types
func Generate(projectPath string) {
	eapi, err := extensionapiparser.ParseExtensionApiJson(projectPath)

	if err != nil {
		panic(err)
	}

	err = GenerateGlobalConstants(projectPath, eapi)

	if err != nil {
		panic(err)
	}

	err = GenerateGlobalEnums(projectPath, eapi)

	if err != nil {
		panic(err)
	}

	err = GenerateNativeStrucutres(projectPath, eapi)

	if err != nil {
		panic(err)
	}

	err = GenerateBuiltinClasses(projectPath, eapi)

	if err != nil {
		panic(err)
	}

	err = GenerateCHeaderClasses(projectPath, eapi)

	if err != nil {
		panic(err)
	}

	err = GenerateCClasses(projectPath, eapi)

	if err != nil {
		panic(err)
	}

	err = GenerateClasses(projectPath, eapi)

	if err != nil {
		panic(err)
	}

	err = GenerateUtilityFunctions(projectPath, eapi)

	if err != nil {
		panic(err)
	}
}

func GenerateGlobalConstants(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("globalconstants.gen.go").
		Parse(globalConstantsText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdnative", fmt.Sprintf("globalconstants.gen.go"))

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

func GenerateNativeStrucutres(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("nativestructures.gen.go").
		Funcs(template.FuncMap{
			"nativeStructureFormatToFields": nativeStructureFormatToFields,
			"hasPrefix":                     strings.HasPrefix,
		}).
		Parse(nativeStructuresText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdnative", fmt.Sprintf("nativeStructures.gen.go"))

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

func GenerateGlobalEnums(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("globalenums.gen.go").
		Parse(globalEnumsText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdnative", fmt.Sprintf("globalenums.gen.go"))

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

func GenerateBuiltinClasses(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("builtinclasses.gen.go").
		Funcs(template.FuncMap{
			"upper":                    strings.ToUpper,
			"lowerFirstChar":           lowerFirstChar,
			"upperFirstChar":           upperFirstChar,
			"lowerCamel":               strcase.ToLowerCamel,
			"screamingSnake":           screamingSnake,
			"goMethodName":             goMethodName,
			"goArgumentName":           goArgumentName,
			"goArgumentType":           goArgumentType,
			"goHasArgumentTypeEncoder": goHasArgumentTypeEncoder,
			"goReturnType":             goReturnType,
			"getOperatorIdName":        getOperatorIdName,
			"isCopyConstructor":        isCopyConstructor,
			"typeHasPtr":               typeHasPtr,
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

	filename := filepath.Join(projectPath, "pkg", "gdnative", fmt.Sprintf("builtinclasses.gen.go"))

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

func GenerateClasses(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.gen.go").
		Funcs(template.FuncMap{
			"goVariantConstructor": goVariantConstructor,
			"goMethodName":         goMethodName,
			"goArgumentName":       goArgumentName,
			"goArgumentType":       goArgumentType,
			"goReturnType":         goReturnType,
			"goClassEnumName":      goClassEnumName,
		}).
		Parse(classesText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdnative", fmt.Sprintf("classes.gen.go"))

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

func GenerateCHeaderClasses(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.gen.h").
		Funcs(template.FuncMap{
			"goMethodName":    goMethodName,
			"goArgumentName":  goArgumentName,
			"goArgumentType":  goArgumentType,
			"goReturnType":    goReturnType,
			"goClassEnumName": goClassEnumName,
		}).
		Parse(cHeaderClassesText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdnative", fmt.Sprintf("classes.gen.h"))

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

func GenerateCClasses(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.gen.c").
		Funcs(template.FuncMap{
			"goMethodName":    goMethodName,
			"goArgumentName":  goArgumentName,
			"goArgumentType":  goArgumentType,
			"goReturnType":    goReturnType,
			"goClassEnumName": goClassEnumName,
		}).
		Parse(cClassesText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdnative", fmt.Sprintf("classes.gen.c"))

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

func GenerateUtilityFunctions(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("utilityfunctions.gen.go").
		Funcs(template.FuncMap{
			"camelCase":      strcase.ToCamel,
			"goArgumentType": goArgumentType,
			"goReturnType":   goReturnType,
		}).
		Parse(utilityFunctionsText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdnative", fmt.Sprintf("utilityfunctions.gen.go"))

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
