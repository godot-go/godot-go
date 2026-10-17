package ffi

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	ifaceDocStartRe  = regexp.MustCompile(`^\s*/\*\*`)
	ifaceNameTagRe   = regexp.MustCompile(`^\s*\*\s*@name\s+(\S+)`)
	ifaceTypedefIDRe = regexp.MustCompile(`\(\*(GDExtensionInterface[A-Za-z0-9_]+)\)\s*\(`)
)

// parseInterfaceNameMap scans raw gdextension_interface.h text and maps each
// interface-function typedef identifier to the registration name recorded in
// the `@name` tag of the doc block directly above it. The engine generates the
// header with that convention, and the @name value is the symbol the engine
// registers, unlike the snake-cased typedef identifier.
//
// Interface typedefs without their own @name are left absent so callers fall
// back to the legacy identifier-based name transform.
func parseInterfaceNameMap(headerText string) map[string]string {
	nameMap := make(map[string]string)
	pendingName := ""
	for _, line := range strings.Split(headerText, "\n") {
		if ifaceDocStartRe.MatchString(line) {
			pendingName = ""
		}
		if match := ifaceNameTagRe.FindStringSubmatch(line); match != nil {
			pendingName = match[1]
		}
		if match := ifaceTypedefIDRe.FindStringSubmatch(line); match != nil {
			if pendingName != "" {
				nameMap[match[1]] = pendingName
			}
			pendingName = ""
		}
	}
	return nameMap
}

// loadInterfaceNameMap builds the @name map from the interface header bundled
// with the project. A missing or unreadable header yields an empty map so
// generation against a tree without the header (for example generator tests on
// temp directories) falls back entirely to the legacy name transform.
func loadInterfaceNameMap(projectPath string) map[string]string {
	headerPath := filepath.Join(projectPath, "godot_headers", "godot", "gdextension_interface.h")
	data, err := os.ReadFile(headerPath)
	if err != nil {
		return map[string]string{}
	}
	return parseInterfaceNameMap(string(data))
}

// interfaceFunctionName returns the symbol name the engine registers for the
// given GDExtensionInterface* typedef: the header's @name value when known,
// otherwise the legacy snake-case identifier transform with fixups.
func interfaceFunctionName(nameMap map[string]string, typeName string) string {
	if name, ok := nameMap[typeName]; ok {
		return name
	}
	return loadProcAddressName(typeName)
}
