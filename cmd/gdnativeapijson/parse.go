package gdnativeapijson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/pinzolo/casee"
)

type Argument []string

func (a *Argument) FunctionArgument() string {
	if strings.HasSuffix((*a)[0], "**") {
		return fmt.Sprintf("%s* %s[]", (*a)[0][:len((*a)[0])-2], (*a)[1])
	}

	return fmt.Sprintf("%s %s", (*a)[0], (*a)[1])
}

type Arguments []Argument

func (as *Arguments) DataTypeRegexMatch(r *regexp.Regexp) bool {
	for _, a := range *as {
		if r.MatchString(a[0]) {
			return true
		}
	}
	return false
}

var cTypeRegex = regexp.MustCompile(`(const)?\s*([\w_][\w_\d]*)\s*(\**)`)

func parseGoType(cTypeName string) GoType {
	matches := cTypeRegex.FindAllStringSubmatch(cTypeName, 1)

	if len(matches) == 0 {
		panic(fmt.Errorf("unrecognized argument: %+q", cTypeName))
	}

	tokens := matches[0]

	var (
		hasConst bool
		rt       ReferenceType
	)

	hasConst = tokens[1] == "const"

	switch tokens[3] {
	case "*":
		rt = PointerReferenceType
	case "**":
		rt = PointerArrayReferenceType
	}

	return NewGoType(hasConst, rt, tokens[2])
}

func parseArgument(argument Argument) GoArgument {
	goType := parseGoType(([]string)(argument)[0])

	return GoArgument{
		Type: goType,
		Name: ([]string)(argument)[1],
	}
}

func (as Arguments) ToDestAndArguments() (*GoArgument, []GoArgument) {
	receiver := parseArgument(as[0])

	args := make([]GoArgument, len(as)-1)

	for i, a := range as[1:] {
		args[i] = parseArgument(a)
	}
	return &receiver, args
}

type ApiFunctionName string

type ApiFunction struct {
	Name       ApiFunctionName `json:"name"`
	ReturnType string          `json:"return_type"`
	Arguments  Arguments       `json:"arguments"`
}

type ApiFunctions []ApiFunction

func (n ApiFunctionName) toPascal() string {
	result := casee.ToPascalCase(stripGodotPrefix(string(n)))

	return fixPascalCase(result)
}

//ToGoMethodName returns the Go method name and a flag specifying if the method is global
func (n ApiFunctionName) toGoMethodName(receiver *GoArgument) (string, bool) {
	pascalCase := n.toPascal()

	if !strings.HasPrefix(pascalCase, receiver.Type.GoName) {
		return pascalCase, true
	}

	// remove the receiver name from the prefix of the method name
	method := pascalCase[len(receiver.Type.GoName):]

	return method, false
}

var cConstructorFunctionRegex = regexp.MustCompile(`(godot_[\w_][\w_\d]*)_new(_(?:[\w_][\w_\d]*))?`)

func (n ApiFunctionName) toGoConstructorName() (string, string, bool) {
	matches := cConstructorFunctionRegex.FindAllStringSubmatch(string(n), 1)

	if len(matches) == 0 {
		return "", "", false
	}

	typeName, _ := ToGoTypeName(ToPascalCase(stripGodotPrefix(matches[0][1])))
	method := fmt.Sprintf("New%s%s", typeName, ToPascalCase(matches[0][2]))

	return typeName, method, true
}

func (f ApiFunction) IsConstructor() bool {
	ret := f.Arguments[0][1] == "r_dest"

	if ret {
		if f.ReturnType != "void" {
			log.Panicf("C constructor function %s is expected to have a void return type; however, actual type is %s", f.Name, f.ReturnType)
		}

		if !strings.Contains(string(f.Name), "_new_") && !strings.HasSuffix(string(f.Name), "_new") {
			log.Panicf("C constructor function %s is expected to have \"_new_\" in function name", f.Name)
		}
	}

	return ret
}

func (f ApiFunction) ToGoMethod(apiMetadata ApiMetadata) GoMethod {
	returnTypeName, methodName, isConstructor := f.Name.toGoConstructorName()

	if isConstructor {
		dest, args := f.Arguments.ToDestAndArguments()

		destTypeName := dest.Type.GoName

		if returnTypeName != destTypeName {
			log.Panicf("C constructor function %s prefix %s is expected to match 1st argument %s", f.Name, returnTypeName, destTypeName)
		}

		return GoMethod{
			Name:         methodName,
			GoMethodType: ConstructorGoMethodType,
			ReturnType:   dest.Type,
			Receiver:     nil,
			Arguments:    args,
			CName:        string(f.Name),
			ApiMetadata:  apiMetadata,
		}
	}

	returnType := parseGoType(f.ReturnType)
	receiver, args := f.Arguments.ToDestAndArguments()
	methodName, isGlobal := f.Name.toGoMethodName(receiver)

	if isGlobal {
		return GoMethod{
			Name:         methodName,
			GoMethodType: GlobalGoMethodType,
			ReturnType:   returnType,
			Receiver:     nil,
			Arguments:    append([]GoArgument{*receiver}, args...),
			CName:        string(f.Name),
			ApiMetadata:  apiMetadata,
		}
	}

	return GoMethod{
		Name:         methodName,
		GoMethodType: ReceiverGoMethodType,
		ReturnType:   returnType,
		Receiver:     receiver,
		Arguments:    args,
		CName:        string(f.Name),
		ApiMetadata:  apiMetadata,
	}
}

// APIVersion is a single APIVersion definition in `gdnative_api.json`
type APIVersion struct {
	Name    *string `json:"name"`
	Type    string  `json:"type"`
	Version struct {
		Major int `json:"major"`
		Minor int `json:"minor"`
	} `json:"version"`
	Next           *APIVersion  `json:"next"`
	API            ApiFunctions `json:"api"`
	isFirstVersion bool
}

func (a *APIVersion) StructTypeAndVersion(structType string) string {
	if a.isFirstVersion {
		return structType
	}

	return fmt.Sprintf("%s_%d_%d", structType, a.Version.Major, a.Version.Minor)
}

func (a *APIVersion) AllVersions() []APIVersion {
	if a == nil {
		return nil
	}
	a.isFirstVersion = true
	arr := []APIVersion{*a}

	for n := a.Next; n != nil; n = n.Next {
		a.isFirstVersion = false
		if (*n).Type != a.Type {
			panic("next type is assumed to match type")
		}
		arr = append(arr, *n)
	}
	return arr
}

// ApiVersions is a structure based on `gdnative_api.json` in `godot_headers`.
type APIJson struct {
	Core       APIVersion   `json:"core"`
	Extensions []APIVersion `json:"extensions"`
}

// ParseGdnativeApiJson parses gdnative_api.json into a APIJson struct.
func ParseGdnativeApiJson(packagePath string) APIJson {
	filename := packagePath + "/godot_headers/gdnative_api.json"
	// Open the gdnative_api.json file that defines the GDNative APIVersion.
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	// Unmarshal the JSON into our struct.
	var apiJson APIJson
	if err := json.Unmarshal(body, &apiJson); err != nil {
		panic(err)
	}

	return apiJson
}
