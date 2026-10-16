package ffi

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleInterfaceHeader mirrors the real gdextension_interface.h doc-comment
// convention: one /** block with @name directly above each interface typedef.
const sampleInterfaceHeader = `
typedef void (*GDExtensionInterfaceFunctionPtr)();

/**
 * @name callable_custom_get_userdata
 * @since 4.2
 *
 * Retrieves the userdata pointer from a custom Callable.
 *
 * @param p_callable A pointer to a Callable.
 *
 * @return The userdata pointer given when creating this custom Callable.
 */
typedef void *(*GDExtensionInterfaceCallableCustomGetUserData)(GDExtensionConstTypePtr p_callable, void *p_token);

/**
 * @name variant_get_ptr_internal_getter
 * @since 4.4
 */
typedef GDExtensionVariantGetInternalPtrFunc (*GDExtensionInterfaceGetVariantGetInternalPtrFunc)(GDExtensionVariantType p_type);

/**
 * @name classdb_register_extension_class5
 * @since 4.4
 */
typedef void (*GDExtensionInterfaceClassdbRegisterExtensionClass5)(
	GDExtensionClassLibraryPtr p_library,
	GDExtensionConstStringNamePtr p_class_name,
	const GDExtensionClassCreationInfo5 *p_create_info);

/* No doc block: must not steal the previous typedef's @name. */
typedef void (*GDExtensionInterfaceNoDocBlock)(void);

/* Non-interface typedefs are ignored entirely. */
typedef void (*GDExtensionVariantGetInternalPtrFunc)(GDExtensionVariantPtr);

/**
 * @name some_helper
 * @since 4.1
 */
typedef void (*NotAnInterfaceTypedef)(void);
`

func TestParseInterfaceNameMap(t *testing.T) {
	nameMap := parseInterfaceNameMap(sampleInterfaceHeader)

	assert.Equal(t, "callable_custom_get_userdata", nameMap["GDExtensionInterfaceCallableCustomGetUserData"])
	assert.Equal(t, "variant_get_ptr_internal_getter", nameMap["GDExtensionInterfaceGetVariantGetInternalPtrFunc"])
	assert.Equal(t, "classdb_register_extension_class5", nameMap["GDExtensionInterfaceClassdbRegisterExtensionClass5"])
	assert.NotContains(t, nameMap, "GDExtensionInterfaceNoDocBlock", "typedef without its own @name must be absent, not inherit the previous name")
	assert.NotContains(t, nameMap, "GDExtensionInterfaceFunctionPtr")
	assert.NotContains(t, nameMap, "NotAnInterfaceTypedef")
}

func TestLoadInterfaceNameMapMissingHeader(t *testing.T) {
	// Matches the existing TestGenerate contract: a temp project tree without
	// the header must not fail generation.
	nameMap := loadInterfaceNameMap(filepath.Join(t.TempDir(), "nonexistent"))
	require.NotNil(t, nameMap)
	require.Empty(t, nameMap)
}

func TestInterfaceFunctionNameMetadataWinsOverIdentifier(t *testing.T) {
	nameMap := map[string]string{
		"GDExtensionInterfaceGetVariantGetInternalPtrFunc":   "variant_get_ptr_internal_getter",
		"GDExtensionInterfaceClassdbRegisterExtensionClass5": "classdb_register_extension_class5",
	}

	// The typedef identifier keeps the legacy name; the @name value is what the
	// engine registers, so it must be emitted verbatim.
	assert.Equal(t, "variant_get_ptr_internal_getter",
		interfaceFunctionName(nameMap, "GDExtensionInterfaceGetVariantGetInternalPtrFunc"))
	assert.Equal(t, "classdb_register_extension_class5",
		interfaceFunctionName(nameMap, "GDExtensionInterfaceClassdbRegisterExtensionClass5"))
}

func TestInterfaceFunctionNameFallsBackToLegacyTransform(t *testing.T) {
	nameMap := map[string]string{}
	typeName := "GDExtensionInterfaceObjectDestroy"

	assert.Equal(t, loadProcAddressName(typeName), interfaceFunctionName(nameMap, typeName))
}

// loaderExemptInterfaceTypedefs are typedef identifiers that match the
// interface-typedef shape but never get a LoadProcAddress call:
// GDExtensionInterfaceFunctionPtr is a plain function-pointer alias excluded
// from the loader by the parser's legacy list, and
// GDExtensionInterfaceGetProcAddress is obtained from the extension entry
// point directly.
var loaderExemptInterfaceTypedefs = map[string]bool{
	"GDExtensionInterfaceFunctionPtr":    true,
	"GDExtensionInterfaceGetProcAddress": true,
}

// uncoveredInterfaceTypedefs lists interface-function typedef identifiers in
// the header text that have no @name mapping, excluding loader exemptions.
func uncoveredInterfaceTypedefs(t *testing.T, headerText string) []string {
	t.Helper()
	nameMap := parseInterfaceNameMap(headerText)
	scan := regexp.MustCompile(`\(\*(GDExtensionInterface[A-Za-z0-9_]+)\)\s*\(`)
	seen := map[string]bool{}
	var missing []string
	for _, line := range strings.Split(headerText, "\n") {
		for _, match := range scan.FindAllStringSubmatch(line, -1) {
			id := match[1]
			if seen[id] || loaderExemptInterfaceTypedefs[id] {
				seen[id] = true
				continue
			}
			seen[id] = true
			if _, ok := nameMap[id]; !ok {
				missing = append(missing, id)
			}
		}
	}
	return missing
}

func TestBundledHeaderInterfaceNameCoverage(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..")
	headerBytes, err := os.ReadFile(filepath.Join(repoRoot, "godot_headers", "godot", "gdextension_interface.h"))
	require.NoError(t, err)

	assert.Empty(t, uncoveredInterfaceTypedefs(t, string(headerBytes)),
		"every interface-function typedef in the bundled header must carry @name metadata")
}

func TestCoverageGuardDetectsMissingName(t *testing.T) {
	// Mutation check: strip the @name line from the custom-callable doc block;
	// the guard must report exactly that typedef as uncovered.
	mutated := strings.Replace(
		sampleInterfaceHeader,
		" * @name callable_custom_get_userdata\n",
		"",
		1,
	)
	require.NotEqual(t, sampleInterfaceHeader, mutated)

	assert.Equal(t,
		[]string{"GDExtensionInterfaceCallableCustomGetUserData", "GDExtensionInterfaceNoDocBlock"},
		uncoveredInterfaceTypedefs(t, mutated))
}

func TestBundledHeaderFixedLoadNames(t *testing.T) {
	nameMap := loadInterfaceNameMap(filepath.Join("..", "..", ".."))

	// The four historically wrong LoadProcAddress strings, regression-pinned to
	// the names the engine actually registers.
	assert.Equal(t, "callable_custom_get_userdata", nameMap["GDExtensionInterfaceCallableCustomGetUserData"])
	assert.Equal(t, "classdb_register_extension_class5", nameMap["GDExtensionInterfaceClassdbRegisterExtensionClass5"])
	assert.Equal(t, "classdb_register_extension_class6", nameMap["GDExtensionInterfaceClassdbRegisterExtensionClass6"])
	assert.Equal(t, "variant_get_ptr_internal_getter", nameMap["GDExtensionInterfaceGetVariantGetInternalPtrFunc"])
}
