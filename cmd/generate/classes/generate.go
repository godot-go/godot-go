package classes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pinzolo/casee"
)

type convertClassView struct {
	PackageName       string
	TemplateName      string
	GeneratedFileName string
	Apis              []GDAPI
}

type view struct {
	PackageName  string
	TemplateName string
	Apis         GDAPIs
}

type gdapiTree struct {
	node     GDAPI
	children map[string]gdapiTree
}

// returns a map of GDAPI.Name to the lineage as a []string
type gdapiPathIndex map[string][]string

// returns a map of GDAPI.Name to the path of ancestors as a []string
func (t *gdapiTree) BuildPathIndex(acc *gdapiPathIndex, lineage []string) {
	var x []string
	if t.node.BaseClass == "" {
		x = lineage[:]
	} else {
		x = append(lineage[:], t.node.BaseClass)
	}

	(*acc)[t.node.Name] = x

	for _, v := range t.children {
		v.BuildPathIndex(acc, x)
	}
}

func arrayContains(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}

	return false
}

func buildAncestryTree(parts ApiTypeBaseClassIndex, apiType, parent, class string) gdapiTree {
	var n *GDAPI

	parentClasses, ok := parts[apiType]

	if !ok {
		log.Panicf("api type '%s' not found", apiType)
	}

	for _, a := range parentClasses[parent] {
		if a.Name == class {
			n = &a
			break
		}
	}

	if n == nil {
		log.Panicf("could not find class %s", class)
	}

	tree := gdapiTree{
		node:     *n,
		children: map[string]gdapiTree{},
	}
	for _, a := range parentClasses[class] {
		log.Printf("inserting %s", a.Name)
		tree.children[a.Name] = buildAncestryTree(parts, apiType, class, a.Name)
	}

	return tree
}

func AssertIsDirectory(dirPath string) error {
	fi, err := os.Stat(dirPath)

	if err != nil {
		return fmt.Errorf("directory '%s' not found", dirPath)
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		break
	default:
		return fmt.Errorf("'%s' is not a directory", dirPath)
	}

	return nil
}

// Generate reads in Godot's api.json and generates codegen of all the classes
func Generate(packagePath string) {
	outputPackageDirectoryPath := filepath.Join(packagePath, "pkg", "gdnative")
	apiJsonFilePath := filepath.Join(packagePath, "godot_headers", "api.json")
	templateFilePath := filepath.Join(packagePath, "cmd", "generate", "classes", "class.go.tmpl")

	var (
		outputFile      *os.File
		err             error
		tmpl            *template.Template
		templateContent []byte
	)

	// pre-condition check output directory
	if err = AssertIsDirectory(outputPackageDirectoryPath); err != nil {
		log.Panicln("pre-condition error:", err)
	}

	funcMap := template.FuncMap{
		"ToSnakeCase":                  casee.ToSnakeCase,
		"ToUpper":                      strings.ToUpper,
		"GoType":                       GoType,
		"GoTypeUsage":                  GoTypeUsage,
		"USAGE_VOID":                   func() string { return "USAGE_VOID" },
		"USAGE_GO_PRIMATIVE":           func() string { return "USAGE_GO_PRIMATIVE" },
		"USAGE_GDNATIVE_CONST_OR_ENUM": func() string { return "USAGE_GDNATIVE_CONST_OR_ENUM" },
		"USAGE_GODOT_STRING":           func() string { return "USAGE_GODOT_STRING" },
		"USAGE_GODOT_STRING_NAME":      func() string { return "USAGE_GODOT_STRING_NAME" },
		"USAGE_GDNATIVE_REF":           func() string { return "USAGE_GDNATIVE_REF" },
		"USAGE_GODOT_CONST_OR_ENUM":    func() string { return "USAGE_GODOT_CONST_OR_ENUM" },
		"USAGE_GODOT_CLASS":            func() string { return "USAGE_GODOT_CLASS" },
	}

	templateContent, err = ioutil.ReadFile(templateFilePath)
	if err != nil {
		log.Panicln("error reading template file:", err)
	}

	if tmpl, err = template.New(templateFilePath).Funcs(funcMap).Parse(string(templateContent)); err != nil {
		log.Panicln("error parsing template:", err)
	}

	body, err := ioutil.ReadFile(apiJsonFilePath)
	if err != nil {
		log.Panic(err)
	}

	var apis GDAPIs
	if err := json.Unmarshal(body, &apis); err != nil {
		panic(err)
	}

	parts := apis.PartitionByBaseApiTypeAndClass()

	objectTree := buildAncestryTree(parts, "core", "", "Object")
	index := gdapiPathIndex{}
	objectTree.BuildPathIndex(&index, []string{})

	// Open the output file for writing
	outputPath := filepath.Join(outputPackageDirectoryPath, "classes.gen.go")

	if outputFile, err = os.Create(outputPath); err != nil {
		panic(err)
	}

	defer outputFile.Close()

	// write header
	WriteTemplate(
		outputFile,
		tmpl,
		view{
			PackageName:  filepath.Base(outputPackageDirectoryPath),
			TemplateName: "class_header.go.tmpl",
			Apis:         apis.FilterForObject(index),
		},
	)

	log.Println("done generating classes")
}

// WriteTemplate writes the output of the specified template to outfile.
func WriteTemplate(outfile *os.File, tmpl *template.Template, view interface{}) {
	var (
		err          error
		generatedBuf bytes.Buffer
	)

	// Write the template with the given view to a buffer.
	if err = tmpl.Execute(&generatedBuf, view); err != nil {
		panic(err)
	}

	generatedBytes := generatedBuf.Bytes()

	outfile.Write(generatedBytes)
}
