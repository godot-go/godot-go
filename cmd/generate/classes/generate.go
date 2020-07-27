package classes

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
	GeneratedFileName string
	Api               GDAPI
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

func Generate(packagePath string) {
	outputPackageDirectoryPath := filepath.Join(packagePath, "pkg", "gdnative")
	md5path := filepath.Join(packagePath, "tmp")
	apiJsonFilePath := filepath.Join(packagePath, "godot_headers", "api.json")
	templateFilePath := filepath.Join(packagePath, "cmd", "generate", "classes", "class.go.tmpl")

	var (
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
		"USAGE_GDNATIVE_RAW":           func() string { return "USAGE_GDNATIVE_RAW" },
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

	hasChanges := false

	// Loop through all of the ApiVersions and generate packages for them.
	for i, api := range apis {
		log.Printf("Generating %d/%d %s...\n", i+1, len(apis), api.PrefixName())
		if !arrayContains(index[api.Name], "Object") && api.Name != "Object" {
			log.Printf("skipping class %s", api.Name)
			continue
		}

		// Get the package name to generate
		// Note: Some entities are prefixed with an underscore that we want to remove.
		outFileName := fmt.Sprintf("%s.classgen.go", strings.TrimLeft(strings.ToLower(api.Name), "_"))

		v := view{
			PackageName:       filepath.Base(outputPackageDirectoryPath),
			TemplateName:      "class.go.tmpl",
			GeneratedFileName: outFileName,
			Api:               api,
		}

		if writeTemplate(
			tmpl,
			filepath.Join(outputPackageDirectoryPath, outFileName),
			filepath.Join(md5path, outFileName + ".pre-fmt.md5"),
			v,
		) {
			hasChanges = true
		}
	}

	log.Println("done generating files.")

	// apply go fmt and go imports on the generated files
	// filepath.Join(outputPackageDirectoryPath, "...")

	if hasChanges {
		log.Println("running go fmt on files.")
		execGoFmt(outputPackageDirectoryPath)

		// log.Println("running goimports on files.")
		// execGoImports(outputPackageDirectoryPath)
	}

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
func writeTemplate(tmpl *template.Template, outputPath string, md5Path string, v view) bool {
	var (
		err error
		generatedBuf bytes.Buffer
		generatedMd5Bytes []byte
		md5Bytes []byte
		isSame int = -1
	)

	// Write the template with the given view to a buffer.
	err = tmpl.Execute(&generatedBuf, v)
	if err != nil {
		panic(err)
	}

	generatedBytes := generatedBuf.Bytes()


	// skip if the output is unchanged
	if fileExists(outputPath) && fileExists(md5Path) {
		md5Bytes, err = ioutil.ReadFile(md5Path)
		
		gsum := md5.Sum(generatedBytes)
		generatedMd5Bytes = gsum[:]
		isSame = bytes.Compare(generatedMd5Bytes, md5Bytes)

		// log.Printf("%s: generated checksum %v, file checksum %v\n", outputPath, generatedMd5Bytes, md5Bytes)
	}

	if isSame == 0 {
		log.Printf("No changes found; skip generating type %s\n", outputPath)
		return false
	}

	// Open the output file for writing
	f, err := os.Create(outputPath)
	f.Write(generatedBytes)
	defer f.Close()
	if err != nil {
		panic(err)
	}

	if err := os.MkdirAll(filepath.Dir(md5Path), os.ModePerm); err !=nil {
		panic(err)
	}

	fmd5, err := os.Create(md5Path)
	fmd5.Write(generatedMd5Bytes)
	defer fmd5.Close()
	if err != nil {
		panic(err)
	}

	return true
}

func execGoFmt(filePath string) {
	cmd := exec.Command("gofmt", "-w", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Panic(fmt.Errorf("error running gofmt: \n%s\n%w", string(output), err))
	}
}

func execGoImports(filePath string) {
	cmd := exec.Command("goimports", "-w", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Panic(fmt.Errorf("error running goimports: \n%s\n%w", string(output), err))
	}
}
