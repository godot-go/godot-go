package globalstate

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
	//go:embed classes.init.go.tmpl
	classesInitText string

	//go:embed classes.callbacks.h.tmpl
	cHeaderClassesText string

	//go:embed classes.callbacks.c.tmpl
	cClassesText string
)

// Generate will generate Go wrappers for all Godot base types
func Generate(projectPath string, eapi extensionapiparser.ExtensionApi) {
	var (
		err error
	)
	if err = GenerateCHeaderClassCallbacks(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateCClassCallbacks(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassInit(projectPath, eapi); err != nil {
		panic(err)
	}
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

	filename := filepath.Join(projectPath, "pkg", "globalstate", fmt.Sprintf("classes.init.gen.go"))

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

	filename := filepath.Join(projectPath, "pkg", "globalstate", fmt.Sprintf("classes.callbacks.gen.h"))

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

	filename := filepath.Join(projectPath, "pkg", "globalstate", fmt.Sprintf("classes.callbacks.gen.c"))

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
