package gdnativeapijson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type Argument []string

func (a *Argument) FunctionArgument() string {
	if strings.HasSuffix((*a)[0], "**") {
		return fmt.Sprintf("%s* %s[]", (*a)[0][:len((*a)[0])-2], (*a)[1])
	} else {
		return fmt.Sprintf("%s %s", (*a)[0], (*a)[1])
	}
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

type ApiFunctions []struct {
	Name       string    `json:"name"`
	ReturnType string    `json:"return_type"`
	Arguments  Arguments `json:"arguments"`
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
