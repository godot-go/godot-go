// Package types is responsible for parsing the Godot headers for type definitions
// and generating Go wrappers around that structure.
package types

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/pcting/godot-go/cmd/gdnativeapijson"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// TODO: Some headers have been removed to reduce compile time.
var ignoreHeaders = []string{
	"pluginscript/godot_pluginscript.h",
	"android/godot_android.h",
	"arvr/godot_arvr.h",
}

// These are being ignored for the time being
var ignoreStructs = []string{
	"godot_char_type",
	"godot_instance_create_func",
	"godot_instance_destroy_func",
	"godot_instance_method",
	"godot_method_attributes",
	"godot_property_get_func",
	"godot_property_set_func",
	"godot_property_usage_flags",
}

var ignoreMethods = []string{
	"godot_string_casecmp_to",
	"godot_string_nocasecmp_to",
	"godot_string_naturalnocasecmp_to",
	"godot_string_begins_with_char_array",
	"godot_string_findmk_from_in_place",
	"godot_string_format_with_custom_placeholder",
	"godot_string_get_slicec",
	"godot_string_hash",
	"godot_string_hex_encode_buffer",
	"godot_string_md5",
	"godot_string_num",
	"godot_string_num_int64",
	"godot_string_num_int64_capitalized",
	"godot_string_num_real",
	"godot_string_num_scientific",
	"godot_string_num_with_decimals",
	"godot_string_char_to_double",
	"godot_string_char_to_int",
	"godot_string_wchar_to_int",
	"godot_string_char_to_int_with_len",
	"godot_string_char_to_int64_with_len",
	"godot_string_unicode_char_to_double",
	"godot_string_char_lowercase",
	"godot_string_char_uppercase",
	"godot_string_chars_to_utf8",
	"godot_string_chars_to_utf8_with_len",
	"godot_string_hash_chars",
	"godot_string_hash_chars_with_len",
	"godot_string_hash_utf8_chars",
	"godot_string_hash_utf8_chars_with_len",
	"godot_string_humanize_size",
	"godot_get_global_constants",
	"godot_register_native_call_type",
	"godot_get_class_constructor",
}

// View is a structure that holds the api struct, so it can be used inside our template.
type View struct {
	TemplateName string
	TypeDefs     []gdnativeapijson.GoTypeDef
	Globals      []gdnativeapijson.GoMethod
}

func arrayContains(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}

	return false
}

// Generate will generate Go wrappers for all Godot base types
func Generate(packagePath string) {
	// parseGodotHeaders all available receiverMethods
	gdnativeAPI := gdnativeapijson.ParseGdnativeApiJson(packagePath)

	// Convert the API definitions into a method struct
	constructors := ConstructorIndex{}
	globalMethods := make([]gdnativeapijson.GoMethod, 0, len(gdnativeAPI.Core.API))
	receiverMethods := MethodIndex{}
	core := &gdnativeAPI.Core
	
	for core != nil {
		apiNameKey := fmt.Sprintf("%s:%d.%d", core.Type, core.Version.Major, core.Version.Minor)
		apiMetadata, ok := gdnativeapijson.ApiNameMap[apiNameKey]

		if ok {
			for _, api := range core.API {
				if arrayContains(ignoreMethods, string(api.Name)) {
					continue
				}
	
				if len(api.Arguments) == 0 {
					log.Panicf("unable to handle C function %s with empty arguments", api.Name)
				}

				method := api.ToGoMethod(apiMetadata)
	
				switch method.GoMethodType {
				case gdnativeapijson.ConstructorGoMethodType:	
					if cons, ok := constructors[method.ReturnType.CName]; ok {
						constructors[method.ReturnType.CName] = append(cons, method)
					} else {
						constructors[method.ReturnType.CName] = []gdnativeapijson.GoMethod{method}
					}

				case gdnativeapijson.GlobalGoMethodType:
					globalMethods = append(globalMethods, method)

				case gdnativeapijson.ReceiverGoMethodType:
					if ms, ok := receiverMethods[method.Receiver.Type.CName]; ok {
						receiverMethods[method.Receiver.Type.CName] = append(ms, method)
					} else {
						receiverMethods[method.Receiver.Type.CName] = []gdnativeapijson.GoMethod{method}
					}
				}
			}
		}

		core = core.Next
	}

	// parseGodotHeaders the Godot header files for type definitions
	index := parseGodotHeaders(packagePath, constructors, receiverMethods, ignoreHeaders, ignoreStructs)

	headers := make([]string, 0, len(index))
	for h := range index {
		headers = append(headers, h)
	}
	sort.Strings(headers)

	// Loop through each header name and generate the Go code in a file based
	// on the header name.
	log.Println("Generating Go wrappers for Godot base types...")
	for _, filePath := range headers {
		typeDefMap := index[filePath]

		typeDefs := make([]gdnativeapijson.GoTypeDef, 0, len(typeDefMap))
		for _, td := range typeDefMap {
			typeDefs = append(typeDefs, td)
		}

		// Convert the header name into the Go filename
		filename := filepath.Base(filePath)
		outFileName := filename[:len(filename)-len(".h")] + "_typegen.go"
		if strings.Index(outFileName, "godot_") == 0 {
			outFileName = outFileName[len("godot_"):]
		}

		log.Printf("  Generating Go code for %s...\n", outFileName)

		// Create a structure for our template view. This will contain all of
		// the data we need to construct our Go wrappers.
		view := View{
			TemplateName: "type.go.tmpl",
			TypeDefs:     typeDefs,
		}

		// Write the file using our template.
		outPath := filepath.Join(packagePath, "pkg", "gdnative")
		md5Path := filepath.Join(packagePath, "tmp")
		ret := writeTemplate(
			filepath.Join(packagePath, "cmd", "generate", "types", "type.go.tmpl"),
			filepath.Join(outPath, outFileName),
			filepath.Join(md5Path, outFileName + ".pre-fmt.md5"),
			view,
		)

		if ret {
			// Run gofmt on the generated Go file.
			goFmt(filepath.Join(outPath, outFileName))
	
			goImports(filepath.Join(outPath, outFileName))
		} else {
			log.Printf("No changes found for %s; skipping go-fmt and go-imports", outFileName)
		}
	}

	if len(globalMethods) > 0 {
		// Create a structure for our template view. This will contain all of
		// the data we need to construct our Go wrappers.
		view := View{
			TemplateName: "type.go.tmpl",
			Globals:      globalMethods,
		}

		// Write the file using our template.
		outPath := filepath.Join(packagePath, "pkg", "gdnative")
		md5Path := filepath.Join(packagePath, "tmp")
		ret := writeTemplate(
			filepath.Join(packagePath, "cmd", "generate", "types", "type.go.tmpl"),
			filepath.Join(outPath, "globals_typegen.go"),
			filepath.Join(md5Path, "globals_typegen.go.pre-fmt.md5"),
			view,
		)

		if ret {
			// Run gofmt on the generated Go file.
			goFmt(filepath.Join(outPath, "globals_typegen.go"))

			goImports(filepath.Join(outPath, "globals_typegen.go"))
		} else {
			log.Printf("No changes found for globals_typegen.go; skipping go-fmt and go-imports")
		}
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

func goFmt(filePath string) {
	cmd := exec.Command("gofmt", "-w", filePath)
	log.Println("Running:", cmd)
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	err := cmd.Run()
	if err != nil {
		log.Println("Error running gofmt:", err)
		panic(stdErr.String())
	}
}

func goImports(filePath string) {
	cmd := exec.Command("goimports", "-w", filePath)
	log.Println("Running:", cmd)
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	err := cmd.Run()
	if err != nil {
		log.Println("Error running goimports:", err)
		panic(stdErr.String())
	}
}
