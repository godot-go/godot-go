package gdutilfunc

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	_ "embed"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
	"github.com/iancoleman/strcase"
)

var (
	//go:embed utilityfunctions.go.tmpl
	utilityFunctionsText string
)

// Generate will generate Go wrappers for all Godot base types
func Generate(projectPath string, eapi extensionapiparser.ExtensionApi) {
	var (
		err error
	)
	if err = GenerateUtilityFunctions(projectPath, eapi); err != nil {
		panic(err)
	}
}

func GenerateUtilityFunctions(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("utilityfunctions.gen.go").
		Funcs(template.FuncMap{
			"camelCase":           strcase.ToCamel,
			"goArgumentName":      goArgumentName,
			"goArgumentType":      goArgumentType,
			"goEncoder":           goEncoder,
			"goReturnType":        goReturnType,
			"coalesce":            coalesce,
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

	filename := filepath.Join(projectPath, "pkg", "gdutilfunc", fmt.Sprintf("utilityfunctions.gen.go"))

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
