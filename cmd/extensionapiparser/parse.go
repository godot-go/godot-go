package extensionapiparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func strictUnmarshal(data []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// ParseGdextensionApiJson parses gdextension_api.json into a APIJson struct.
func ParseExtensionApiJson(projectPath string) (ExtensionApi, error) {
	filename := projectPath + "/godot_headers/extension_api.json"
	// Open the gdextension_api.json file that defines the GDExtension APIVersion.
	body, err := os.ReadFile(filename)
	if err != nil {
		return ExtensionApi{}, err
	}

	// Unmarshal the JSON into our struct.
	var extensionApiJson ExtensionApi
	if err := strictUnmarshal(body, &extensionApiJson); err != nil {
		return ExtensionApi{}, err
	}

	return extensionApiJson, nil
}

func GenerateExtensionAPI(projectPath, buildConfig string) (ExtensionApi, error) {
	var (
		eapi ExtensionApi
		err  error
	)
	if eapi, err = ParseExtensionApiJson(projectPath); err != nil {
		return ExtensionApi{}, err
	}
	if !eapi.HasBuildConfiguration(buildConfig) {
		return ExtensionApi{}, fmt.Errorf(`unable to find build configuration "%s"`, buildConfig)
	}
	eapi.BuildConfig = buildConfig
	eapi.Classes = eapi.FilteredClasses()

	return eapi, nil
}
