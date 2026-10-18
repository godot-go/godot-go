package gdclassimpl

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// varargCensusFixturePath is the audited list of every is_vararg method in
// the engine API. It backs the audit table in
// openspec/changes/fix-vararg-return-buffer/design.md.
const varargCensusFixturePath = "vararg_census.json"

type varargCensusFixture struct {
	Comment string `json:"comment"`
	Counts  struct {
		Class   int `json:"class"`
		Builtin int `json:"builtin"`
		Utility int `json:"utility"`
	} `json:"counts"`
	Entries []varargCensusEntry `json:"entries"`
}

type varargCensusEntry struct {
	Source     string `json:"source"`
	Class      string `json:"class"`
	Method     string `json:"method"`
	ReturnType string `json:"return_type"`
}

func loadVarargCensus(t *testing.T) varargCensusFixture {
	t.Helper()
	body, err := os.ReadFile(varargCensusFixturePath)
	require.NoError(t, err)
	var census varargCensusFixture
	require.NoError(t, json.Unmarshal(body, &census))
	return census
}

// deriveVarargCensusEntries lists every vararg method across class methods,
// built-in class methods, and utility functions with its declared return
// type, sorted by source, class, and method. An empty declared type is
// normalized to "void".
func deriveVarargCensusEntries(eapi extensionapiparser.ExtensionApi) []varargCensusEntry {
	var entries []varargCensusEntry
	for _, c := range eapi.Classes {
		for _, m := range c.Methods {
			if m.IsVararg {
				entries = append(entries, varargCensusEntry{
					Source:     "class",
					Class:      c.Name,
					Method:     m.Name,
					ReturnType: coalesceVoid(typeOrMeta(m.ReturnValue.Meta, m.ReturnValue.Type)),
				})
			}
		}
	}
	for _, c := range eapi.BuiltinClasses {
		for _, m := range c.Methods {
			if m.IsVararg {
				entries = append(entries, varargCensusEntry{
					Source:     "builtin",
					Class:      c.Name,
					Method:     m.Name,
					ReturnType: coalesceVoid(m.ReturnType),
				})
			}
		}
	}
	for _, f := range eapi.UtilityFunctions {
		if f.IsVararg {
			entries = append(entries, varargCensusEntry{
				Source:     "utility",
				Class:      "",
				Method:     f.Name,
				ReturnType: coalesceVoid(f.ReturnType),
			})
		}
	}
	sourceRank := map[string]int{"class": 0, "builtin": 1, "utility": 2}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Source != entries[j].Source {
			return sourceRank[entries[i].Source] < sourceRank[entries[j].Source]
		}
		if entries[i].Class != entries[j].Class {
			return entries[i].Class < entries[j].Class
		}
		return entries[i].Method < entries[j].Method
	})
	return entries
}

func coalesceVoid(t string) string {
	if t == "" {
		return "void"
	}
	return t
}

// TestVarargCensusMatchesApiSurface keeps the audit fixture in step with the
// engine API. A drift means a vararg method was added, removed, or changed
// its declared return type, so the audit table and the generator guard need
// to be re-checked against the new surface.
func TestVarargCensusMatchesApiSurface(t *testing.T) {
	eapi := loadFixtureExtensionApi(t)
	census := loadVarargCensus(t)

	derived := deriveVarargCensusEntries(eapi)
	derivedCounts := varargCensusFixture{}
	for _, e := range derived {
		switch e.Source {
		case "class":
			derivedCounts.Counts.Class++
		case "builtin":
			derivedCounts.Counts.Builtin++
		case "utility":
			derivedCounts.Counts.Utility++
		}
	}

	assert.Equal(t, census.Counts, derivedCounts.Counts, "vararg census counts must match extension_api.json")
	assert.Equal(t, census.Entries, derived, "vararg census entries must match extension_api.json")

	// The guard requirement: every class-method vararg must carry a Go
	// conversion from the engine's Variant return buffer.
	require.NoError(t, checkVarargReturnTypes(eapi))
}

// TestGeneratedVarargBuffersAreVariantSized scans the committed generated
// file: every r_return pointer cast to GDExtensionUninitializedVariantPtr
// (the varcall shape) must address a Variant, never a smaller declared-type
// variable. A null pointer is the accepted void shape.
func TestGeneratedVarargBuffersAreVariantSized(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(projectRoot, "pkg", "gdclassimpl", "classes.gen.go"))
	require.NoError(t, err)

	const cast = "\tretPtr := (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(&"
	var lastName, lastType string
	found := 0
	for _, line := range strings.Split(string(body), "\n") {
		if rest, ok := strings.CutPrefix(line, "\tvar ret"); ok && !strings.HasPrefix(rest, "Ptr") {
			fields := strings.Fields(line)
			lastName, lastType = fields[1], fields[2]
			continue
		}
		if rest, ok := strings.CutPrefix(line, cast); ok {
			name, _, _ := strings.Cut(rest, ")")
			found++
			require.Equalf(t, name, lastName, "varcall buffer %q has no preceding declaration", name)
			require.Equalf(t, "Variant", lastType, "varcall buffer %q must be a Variant, got %s", name, lastType)
		}
	}
	// The Variant-returning varargs plus the three converted Error-returning
	// ones must all be present; a small count would mean the scan silently
	// stopped matching after a codegen shape change.
	assert.GreaterOrEqual(t, found, 10, "expected every Variant-buffered vararg call to be scanned")
}

// TestUpdateVarargCensusFixture rewrites vararg_census.json from
// extension_api.json, keeping the file's comment. It is skipped unless
// UPDATE_VARARG_CENSUS=1 is set, so a plain `go test` never mutates the
// fixture. Run it after a header refresh whenever the audit test reports
// drift, then re-check the design audit table and the generator guard.
func TestUpdateVarargCensusFixture(t *testing.T) {
	if os.Getenv("UPDATE_VARARG_CENSUS") == "" {
		t.Skip("set UPDATE_VARARG_CENSUS=1 to regenerate vararg_census.json from extension_api.json")
	}
	eapi := loadFixtureExtensionApi(t)
	census := loadVarargCensus(t)

	entries := deriveVarargCensusEntries(eapi)
	census.Entries = entries
	census.Counts.Class = 0
	census.Counts.Builtin = 0
	census.Counts.Utility = 0
	for _, e := range entries {
		switch e.Source {
		case "class":
			census.Counts.Class++
		case "builtin":
			census.Counts.Builtin++
		case "utility":
			census.Counts.Utility++
		}
	}

	body, err := json.MarshalIndent(census, "", " ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(varargCensusFixturePath, append(body, '\n'), 0o644))
	t.Logf("rewrote %s: %d vararg entries", varargCensusFixturePath, len(entries))
}
