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

	//go:embed builtinclasses.bindings.go.tmpl
	builtinClassesBindingsText string

	//go:embed builtinclasses.go.tmpl
	builtinClassesText string

	//go:embed classes.interfaces.go.tmpl
	classesInterfacesText string

	//go:embed classes.enums.go.tmpl
	classesEnumsText string

	//go:embed classes.constants.go.tmpl
	classesConstantsText string

	//go:embed classes.go.tmpl
	classesText string

	//go:embed classes.refs.go.tmpl
	classesRefsText string

	//go:embed classes.init.go.tmpl
	classesInitText string

	//go:embed classes.callbacks.h.tmpl
	cHeaderClassesText string

	//go:embed classes.callbacks.c.tmpl
	cClassesText string

	//go:embed nativestructures.go.tmpl
	nativeStructuresText string
)

// Generate will generate Go wrappers for all Godot base types
func Generate(projectPath, buildConfig string) {
	var (
		eapi extensionapiparser.ExtensionApi
		err error
	)
	if eapi, err = extensionapiparser.ParseExtensionApiJson(projectPath); err != nil {
		panic(err)
	}
	if !eapi.HasBuildConfiguration(buildConfig) {
		panic(fmt.Sprintf(`unable to find build configuration "%s"`, buildConfig))
	}
	eapi.BuildConfig = buildConfig
	eapi.Classes = eapi.FilteredClasses()
	if err = GenerateGlobalConstants(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateGlobalEnums(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateNativeStrucutres(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateBuiltinClasses(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateBuiltinClassBindings(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateCHeaderClassCallbacks(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateCClassCallbacks(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassInterfaces(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassEnums(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassConstants(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClasses(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassRefs(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassInit(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateUtilityFunctions(projectPath, eapi); err != nil {
		panic(err)
	}
}

func GenerateGlobalConstants(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	if len(extensionApi.GlobalConstants) == 0 {
		return nil
	}

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("globalconstants.gen.go"))

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("nativeStructures.gen.go"))

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
	if len(extensionApi.GlobalEnums) == 0 {
		return nil
	}

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("globalenums.gen.go"))

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("builtinclasses.gen.go"))

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("builtinclasses.bindings.gen.go"))

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("classes.interfaces.gen.go"))

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("classes.enums.gen.go"))

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("classes.constants.gen.go"))

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
		Parse(classesText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("classes.gen.go"))

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

func GenerateClassRefs(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.refs.gen.go").
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
		Parse(classesRefsText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("classes.refs.gen.go"))

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

func GenerateClassInit(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.init.gen.go").
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
		Parse(classesInitText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("classes.init.gen.go"))

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

func GenerateCHeaderClassCallbacks(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.callbacks.gen.h").
		Funcs(template.FuncMap{
			"goMethodName":    goMethodName,
			"goArgumentName":  goArgumentName,
			"goArgumentType":  goArgumentType,
			"goReturnType":    goReturnType,
			"goClassEnumName": goClassEnumName,
			"coalesce":        coalesce,
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

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("classes.callbacks.gen.h"))

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

func GenerateCClassCallbacks(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.callbacks.gen.c").
		Funcs(template.FuncMap{
			"goMethodName":    goMethodName,
			"goArgumentName":  goArgumentName,
			"goArgumentType":  goArgumentType,
			"goReturnType":    goReturnType,
			"goClassEnumName": goClassEnumName,
			"coalesce":        coalesce,
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

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("classes.callbacks.gen.c"))

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
			"goArgumentName": goArgumentName,
			"goArgumentType": goArgumentType,
			"goEncoder":      goEncoder,
			"goReturnType":   goReturnType,
			"coalesce":       coalesce,
			"goEncodeIsReference": goEncodeIsReference,
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

	filename := filepath.Join(projectPath, "pkg", "gdextension", fmt.Sprintf("utilityfunctions.gen.go"))

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
