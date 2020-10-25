package classes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pinzolo/casee"
)

type convertClassView struct {
	PackageName       string
	TemplateName      string
	GeneratedFileName string
	Apis              []GDAPI
}

type view struct {
	PackageName       string
	TemplateName      string
	Apis              GDAPIs
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

func assertIsDirectory(dirPath string) error {
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
		f               *os.File
		err             error
		templateContent []byte
	)

	// pre-condition check output directory
	if err = assertIsDirectory(outputPackageDirectoryPath); err != nil {
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

	tmpl, err := template.New(templateFilePath).Funcs(funcMap).Parse(string(templateContent))
	if err != nil {
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

	if f, err = os.Create(outputPath); err != nil {
		panic(err)
	}

	// write header
	writeTemplate(
		f,
		tmpl,
		view{
			PackageName:  filepath.Base(outputPackageDirectoryPath),
			TemplateName: "class_header.go.tmpl",
			Apis:         apis.FilterForObject(index),
		},
	)

	f.Close()

	log.Println("done generating classes")
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

// returns true if there were changes
func writeTemplate(f *os.File, tmpl *template.Template, v view) {
	var (
		err error
		generatedBuf bytes.Buffer
	)

	// Write the template with the given view to a buffer.
	err = tmpl.Execute(&generatedBuf, v)
	if err != nil {
		panic(err)
	}

	generatedBytes := generatedBuf.Bytes()

	f.Write(generatedBytes)
}
