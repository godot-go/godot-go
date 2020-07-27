// Package gdnativewrapper generates C code to wrap all of the gdnative
// methods to call functions on the gdnative_api_structs to work
// around the cgo C function pointer limitation. 
package gdnativewrapper

import (
	"path/filepath"
	"bytes"
	"crypto/md5"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/pcting/godot-go/cmd/gdnativeapijson"
)

// View is a structure that holds the api struct, so it can be used inside
// our template.
type View struct {
	ApiVersions []gdnativeapijson.APIVersion
	Type        string
	StructType  string
	Name        string
}

func (v View) ToCDefine() string {
	return strings.ToUpper(v.StructType)
}

// NotLastElement is a function we use inside the template to test whether or
// not the given element is the last in the slice or not. This is so we can
// correctly insert commas for argument lists.
func (v View) NotLastElement(n int, slice []gdnativeapijson.Argument) bool {
	if n != (len(slice) - 1) {
		return true
	}
	return false
}

// NotVoid checks to see if the return string is void or not. This is used inside
// our template so we can determine if we need to use the `return` keyword in
// the function body.
func (v View) NotVoid(ret string) bool {
	if ret != "void" {
		return true
	}
	return false
}

// HasArgs is a function we use inside the template to test whether or not the
// function has arguments. This is so we can determine if we need to place a
// comma.
func (v View) HasArgs(args []gdnativeapijson.Argument) bool {
	if len(args) != 0 {
		return true
	}
	return false
}

func filterApiFunctions(fs *gdnativeapijson.ApiFunctions, r *regexp.Regexp) (ret gdnativeapijson.ApiFunctions) {
	for _, f := range *fs {
		if r.MatchString(string(f.Name)) {
			log.Printf("function %s(%s) %s ignored because of function mame", f.Name, f.Arguments, f.ReturnType)
			continue
		} else if r.MatchString(f.ReturnType) {
			log.Printf("function %s(%s) %s ignored because of return type", f.Name, f.Arguments, f.ReturnType)
			continue
		} else if f.Arguments.DataTypeRegexMatch(r) {
			log.Printf("function %s(%s) %s ignored because of arguments", f.Name, f.Arguments, f.ReturnType)
			continue
		}

		ret = append(ret, f)
	}
	return
}

func Generate(packagePath string) {

	// Create a structure for our template view. This will contain all of
	// the data we need to construct our binding methods.
	var view View

	// Unmarshal the JSON into our struct.
	apis := gdnativeapijson.ParseGdnativeApiJson(packagePath)

	// Add the core APIVersion to our view first
	view.ApiVersions = apis.Core.AllVersions()
	view.Type = apis.Core.Type
	view.StructType = "core"

	// Generate the C bindings
	log.Println("Generating", view.StructType, "C headers...")
	writeTemplate(
		filepath.Join(packagePath,"cmd/generate/gdnativewrapper/gdnative.h.tmpl"),
		filepath.Join(packagePath,"pkg/gdnative/gdnative.wrappergen.h"),
		filepath.Join(packagePath,"tmp/gdnative.wrappergen.h.md5"),
		view,
	)

	log.Println("Generating", view.StructType, "C bindings...")
	writeTemplate(
		filepath.Join(packagePath, "cmd/generate/gdnativewrapper/gdnative.c.tmpl"),
		filepath.Join(packagePath,"pkg/gdnative/gdnative.wrappergen.c"),
		filepath.Join(packagePath,"tmp/gdnative.wrappergen.c.md5"),
		view,
	)

	log.Println("Generating", view.StructType, "Go bindings...")
	writeTemplate(
		filepath.Join(packagePath, "cmd/generate/gdnativewrapper/gdnative.go.tmpl"),
		filepath.Join(packagePath,"pkg/gdnative/gdnative.wrappergen.go"),
		filepath.Join(packagePath,"tmp/gdnative.wrappergen.go.md5"),
		view,
	)

	// Loop through all of our extensions and generate the bindings for those.
	for _, api := range apis.Extensions {
		view.ApiVersions = api.AllVersions()
		view.Name = *api.Name
		view.StructType = "ext_" + view.Name

		log.Println("Generating", view.StructType, "C headers...")
		writeTemplate(
			filepath.Join(packagePath, "cmd/generate/gdnativewrapper/gdnative.h.tmpl"),
			filepath.Join(packagePath, "pkg/gdnative/"+view.Name+".wrappergen.h"),
			filepath.Join(packagePath, "tmp/"+view.Name+".wrappergen.h.md5"),
			view,
		)

		log.Println("Generating", view.StructType, "C bindings...")
		writeTemplate(
			filepath.Join(packagePath, "cmd/generate/gdnativewrapper/gdnative.c.tmpl"),
			filepath.Join(packagePath, "pkg/gdnative/"+view.Name+".wrappergen.c"),
			filepath.Join(packagePath, "tmp/"+view.Name+".wrappergen.c.md5"),
			view,
		)

		log.Println("Generating", view.StructType, "Go bindings...")
		writeTemplate(
			filepath.Join(packagePath, "cmd/generate/gdnativewrapper/gdnative.go.tmpl"),
			filepath.Join(packagePath,"pkg/gdnative/"+view.Name+".wrappergen.go"),
			filepath.Join(packagePath,"tmp/"+view.Name+".wrappergen.go.md5"),
			view,
		)
	}
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
func writeTemplate(templatePath, outputPath string, md5Path string, view View) bool {
	var (
		err error
		generatedBuf bytes.Buffer
		generatedMd5Bytes []byte
		md5Bytes []byte
		isSame int = -1
		t *template.Template
	)

	// Create a template from our template file.
	t, err = template.ParseFiles(templatePath)
	if err != nil {
		log.Fatal("Error parsing template:", err)
	}

	// Write the template with the given view to a buffer.
	err = t.Execute(&generatedBuf, view)
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

		log.Printf("%s: generated checksum %v, file checksum %v", outputPath, generatedMd5Bytes, md5Bytes)
	}

	if isSame == 0 {
		log.Printf("No changes found; skip generating type %s", outputPath)
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
