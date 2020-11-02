// Package types is a package that parses the GDNative headers for type definitions
// to create wrapper structures for Go.
package types

import (
	"fmt"
	"github.com/pinzolo/casee"
	"io/ioutil"
	"os"
	"regexp"
	"path/filepath"
	"strings"
	"github.com/godot-go/godot-go/cmd/gdnativeapijson"
)

var cTypeRegex = regexp.MustCompile(`(const)?\s*([\w_][\w_\d]*)\s*(\**)`)

// GlobalMethods contains the list of methods not associated with a GoTypeDef
type GlobalMethods []gdnativeapijson.GoMethod

// ConstructorIndex indexes by gdnativeapijson.GoTypeDef.CName
type ConstructorIndex map[string][]gdnativeapijson.GoMethod

// MethodIndex indexes by gdnativeapijson.GoTypeDef.CName
type MethodIndex map[string][]gdnativeapijson.GoMethod

// GoTypeDefIndex indexes by C header file name and then by C typedef name
type GoTypeDefIndex map[string]map[string]gdnativeapijson.GoTypeDef

// parseGodotHeaders will parse the GDNative headers. Takes a list of headers/structs to ignore.
// Definitions in the given headers and definitions
// with the given name will not be added to the returned list of type definitions.
// We'll need to manually create these structures.
func parseGodotHeaders(
	packagePath string,
	constructorIndex ConstructorIndex,
	methodIndex MethodIndex,
	excludeHeaders, excludeStructs []string) GoTypeDefIndex {
	var (
		index           = GoTypeDefIndex{}
		relPath         string
		err             error
		godotHeaderPath = filepath.Join(packagePath, "godot_headers")
	)

	// Walk through all of the godot filename files
	err = filepath.Walk(godotHeaderPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && filepath.Ext(path) == ".h" {
			relPath, err = filepath.Rel(godotHeaderPath, path)
			if err != nil {
				panic(err)
			}

			// Read the filename
			content, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}

			// Find all of the type definitions in the filename file
			// fmt.Println("Parsing File ", path, "...")
			foundTypesLines := findTypeDefs(content)

			// After extracting the lines, we can now parse the type definition to
			// a structure that we can use to build a Go wrapper.
			for _, foundTypeLines := range foundTypesLines {
				typeDef := parseTypeDef(foundTypeLines, relPath)
				typeDef.Constructors = constructorIndex[typeDef.CName]
				typeDef.Methods = methodIndex[typeDef.CName]

				// Only add the type if it's not in our exclude list.
				if !strInSlice(typeDef.CName, excludeStructs) && !strInSlice(typeDef.CHeaderFilename, excludeHeaders) {
					if tdMap, ok := index[relPath]; ok {
						tdMap[typeDef.CName] = typeDef
					} else {
						index[relPath] = map[string]gdnativeapijson.GoTypeDef{
							typeDef.CName: typeDef,
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return index
}

func parseTypeDef(typeLines []string, headerName string) gdnativeapijson.GoTypeDef {
	// Create a structure for our type definition.
	typeDef := gdnativeapijson.GoTypeDef{
		CHeaderFilename: headerName,
		Properties:      []gdnativeapijson.GoProperty{},
	}

	// Small function for splitting a line to get the uncommented line and
	// get the comment itself.
	getComment := func(line string) (def, comment string) {
		halves := strings.Split(line, "//")
		def = halves[0]
		if len(halves) > 1 {
			comment = strings.TrimSpace(halves[1])
		}
		if strings.HasPrefix(comment, "/") {
			comment = strings.Replace(comment, "/", "", 1)
		}

		return def, comment
	}

	// If the type definition is a single line, handle it a little differently
	if len(typeLines) == 1 {
		// Extract the comment if there is one.
		line, comment := getComment(typeLines[0])

		// Check to see if the property is a pointer type
		if strings.Contains(line, "*") {
			line = strings.Replace(line, "*", "", 1)
			typeDef.IsPointer = true
		}

		var err error



		// Get the words of the line
		words := strings.Split(line, " ")
		typeDef.CName = strings.Replace(words[len(words)-1], ";", "", 1)

		goTypeName, usage := gdnativeapijson.ToGoTypeName(typeDef.CName)

		typeDef.Name = goTypeName
		typeDef.Base = words[len(words)-2]
		typeDef.Comment = comment
		typeDef.Usage = usage

		if err != nil {
			panic(fmt.Errorf("%s\n%w", line, err))
		}

		return typeDef
	}

	// Extract the name of the type.
	lastLine := typeLines[len(typeLines)-1]
	words := strings.Split(lastLine, " ")
	typeDef.CName = strings.Replace(words[len(words)-1], ";", "", 1)

	var err error

	// Extract the base type
	firstLine := typeLines[0]
	words = strings.Split(firstLine, " ")
	typeDef.Base = words[1]

	if err != nil {
		panic(fmt.Errorf("%s\n%w", strings.Join(typeLines, "\n"), err))
	}

	// Convert the name of the type to a Go name
	typeDef.Name, _ = gdnativeapijson.ToGoTypeName(typeDef.CName)

	if len(typeDef.Name) == 0 {
		typeDef.Name = words[2]
	}

	// Extract the properties from the type
	var properties []string
	if strings.HasSuffix(strings.TrimSpace(firstLine), "{") {
		properties = typeLines[1 : len(typeLines)-1]
	} else {
		properties = typeLines[2 : len(typeLines)-1]
	}

	var accumLines string

	// Loop through each property line
	for _, line := range properties {
		if strings.HasPrefix(strings.TrimSpace(line), "//") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		if !strings.Contains(line, ";") && typeDef.Base != "enum" {
			accumLines += line
		} else {
			line = accumLines + line
			accumLines = ""
		}

		// Skip function definitions
		if strings.Contains(line, "(*") {
			continue
		}

		// Create a type definition for the property
		property := gdnativeapijson.GoProperty{}

		// Extract the comment if there is one.
		line, comment := getComment(line)
		property.Comment = comment

		// Sanitize the line
		line = strings.TrimSpace(line)
		line = strings.Split(line, ";")[0]
		line = strings.Replace(line, "unsigned ", "u", 1)
		line = strings.Replace(line, "const ", "", 1)

		// Split the line by spaces
		words = strings.Split(line, " ")

		// Check to see if the line is just a comment
		if words[0] == "//" || (strings.Index(line, "/*") == 0 && strings.Index(line, "*/") == (len(line)-2)) {
			continue
		}

		// Set the property details
		if typeDef.Base == "enum" {
			// Strip any commas in the name
			words[0] = strings.Replace(words[0], ",", "", 1)
			property.CName = words[0]
			property.Name = casee.ToPascalCase(strings.Replace(words[0], "GODOT_", "", 1))
		} else {
			if len(words) < 2 {
				fmt.Println("Skipping irregular line:", line)
				continue
			}
			property.Base = words[0]
			property.CName = words[1]
			property.Name = casee.ToPascalCase(strings.Replace(words[1], "godot_", "", 1))
		}

		// Check to see if the property is a pointer type
		if strings.Contains(property.CName, "*") {
			property.CName = strings.Replace(property.CName, "*", "", 1)
			property.Name = strings.Replace(property.Name, "*", "", 1)
			property.IsPointer = true
		}

		// Skip empty property names
		if property.Name == "" {
			continue
		}

		if strings.Contains(property.Name, "}") {
			panic(fmt.Errorf("malformed Name: %+v", property))
		}

		// Append the property to the type definition
		typeDef.Properties = append(typeDef.Properties, property)
	}

	return typeDef
}

type block int8

const (
	externBlock block = iota
	typedefBlock
	localStructBlock
	enumBlock
)

// findTypeDefs will return a list of type definition lines.
func findTypeDefs(content []byte) [][]string {
	lines := strings.Split(string(content), "\n")

	// Create a structure that will hold the lines that define the type.
	var (
		singleType []string
		foundTypes [][]string

		blocks []block
	)
	for i, line := range lines {

		if strings.Index(line, "extern \"C\" {") == 0 {
			// fmt.Println("Line", i ,": START EXTERN BLOCK")
			blocks = append(blocks, externBlock)
			continue
		} else if strings.Index(line, "struct ") == 0 {
			// fmt.Println("Line", i ,": START LOCAL STRUCT BLOCK")
			blocks = append(blocks, localStructBlock)
			continue
		} else if strings.Index(line, "enum ") == 0 {
			// fmt.Println("Line", i ,": START ENUM BLOCK")
			blocks = append(blocks, enumBlock)
			continue
		} else if strings.Index(line, "}") == 0 {
			if len(blocks) == 0 {
				panic(fmt.Sprintln("\tLine", i, ": extra closing bracket encountered", line))
			}
			n := len(blocks)-1
			b := blocks[n]
			blocks = blocks[:n]

			switch b {
			case localStructBlock:
				// fmt.Println("Line", i ,": END LOCAL STRUCT BLOCK")
				continue
			case externBlock:
				// fmt.Println("Line", i ,": END EXTERN BLOCK")
				continue
			case enumBlock:
				// fmt.Println("Line", i ,": END ENUM BLOCK")
				continue
			case typedefBlock:
				singleType = append(singleType, line)
				foundTypes = append(foundTypes, singleType)

				// fmt.Println("\tLine", i, ": Type found:\n", strings.Join(singleType, "\n"))

				// reset
				singleType = []string{}
			default:
				panic(fmt.Sprintln("\tLine", i, ": extra closing curly braket found"))
			}
		} else if strings.Index(line, "typedef ") == 0 {
			// Check to see if this is a single line type and avoid using
			// the blocks stack
			if strings.Contains(line, ";") {
				// Skip if this is a function definition
				if strings.Contains(line, ")") {
					fmt.Println("\tLine", i, ": skip function: ", line)
					continue
				}

				singleType = append(singleType, line)
				foundTypes = append(foundTypes, singleType)

				// fmt.Println("\tLine", i, ": Single line type found:\n", strings.Join(singleType, "\n"))

				// reset
				singleType = []string{}
			} else {
				blocks = append(blocks, typedefBlock)
				singleType = append(singleType, line)
			}
		} else if len(blocks) > 0 {
			b := blocks[len(blocks) - 1]
			if b == typedefBlock {
				singleType = append(singleType, line)
			}
		}

		// // If a type was found, keep appending our struct lines until we
		// // reach the end of the definition.
		// if accumulatingTypedefLines {
		// 	//fmt.Println("Line", i, ": Appending line for type found:", line)

		// 	// Keep adding the lines to our list of lines until we
		// 	// reach an end bracket.
		// 	singleType = append(singleType, line)

		// 	if strings.Contains(line, "}") {
		// 		//fmt.Println("Line", i, ": Found end of type definition.")
		// 		accumulatingTypedefLines = false
		// 		foundTypes = append(foundTypes, singleType)
		// 		singleType = []string{}
		// 	}
		// }
	}

	return foundTypes
}

func strInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
