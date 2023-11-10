package gdclassimpl

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
	//go:embed classes.go.tmpl
	classesText string

	//go:embed classes.refs.go.tmpl
	classesRefsText string
)

// Generate will generate Go wrappers for all Godot base types
func Generate(projectPath string, eapi extensionapiparser.ExtensionApi) {
	var (
		err error
	)
	if err = GenerateClasses(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassRefs(projectPath, eapi); err != nil {
		panic(err)
	}
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

	filename := filepath.Join(projectPath, "pkg", "gdextension", "gdclassimpl", fmt.Sprintf("classes.gen.go"))

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

	filename := filepath.Join(projectPath, "pkg", "gdextension", "gdclassimpl", fmt.Sprintf("classes.refs.gen.go"))

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
