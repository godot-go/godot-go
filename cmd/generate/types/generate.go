// Package types is responsible for parsing the Godot headers for type definitions
// and generating Go wrappers around that structure.
package types

import (
	"bytes"
	"fmt"
	"github.com/godot-go/godot-go/cmd/gdnativeapijson"
	"log"
	"os"
	"path/filepath"
	"sort"
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
	// Create a structure for our template view. This will contain all of
	// the data we need to construct our Go wrappers.
	view := View{
		TemplateName: "type.go.tmpl",
	}

	// parseGodotHeaders all available receiverMethods
	gdnativeAPI := gdnativeapijson.ParseGdnativeApiJson(packagePath)

	// Convert the API definitions into a method struct
	constructors := ConstructorIndex{}
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
					view.Globals = append(view.Globals, method)

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

		typeDefMapKeys := make([]string, 0, len(typeDefMap))
		for d := range typeDefMap {
			typeDefMapKeys = append(typeDefMapKeys, d)
		}
		sort.Strings(typeDefMapKeys)

		for _, k := range typeDefMapKeys {
			view.TypeDefs = append(view.TypeDefs, typeDefMap[k])
		}
	}

	// Open the output file for writing
	outputPackageDirectoryPath := filepath.Join(packagePath, "pkg", "gdnative")
	outputPath := filepath.Join(outputPackageDirectoryPath, "types.gen.go")

	var (
		f *os.File
		err error
	)

	if f, err = os.Create(outputPath); err != nil {
		panic(err)
	}

	writeTemplate(
		f,
		filepath.Join(packagePath, "cmd", "generate", "types", "type.go.tmpl"),
		view,
	)
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
func writeTemplate(f *os.File, templatePath string, view View) {
	var (
		err error
		generatedBuf bytes.Buffer
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

	f.Write(generatedBytes)
}
