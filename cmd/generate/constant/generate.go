package constant

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	_ "embed"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
)

var (
	//go:embed classes.constants.go.tmpl
	classesConstantsText string

	//go:embed classes.enums.go.tmpl
	classesEnumsText string

	//go:embed globalconstants.go.tmpl
	globalConstantsText string

	//go:embed globalenums.go.tmpl
	globalEnumsText string
)

// Generate will generate Go wrappers for all Godot base types
func Generate(projectPath string, eapi extensionapiparser.ExtensionApi) {
	var (
		err error
	)
	if err = GenerateClassConstants(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassEnums(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateGlobalConstants(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateGlobalEnums(projectPath, eapi); err != nil {
		panic(err)
	}
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

	filename := filepath.Join(projectPath, "pkg", "gdextension", "constant", fmt.Sprintf("classes.constants.gen.go"))

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", "constant", fmt.Sprintf("classes.enums.gen.go"))

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", "constant", fmt.Sprintf("globalconstants.gen.go"))

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", "constant", fmt.Sprintf("globalenums.gen.go"))

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
