package gdclassimpl

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// projectRoot is the repository root relative to this package directory.
const projectRoot = "../../.."

const censusFixturePath = "virtual_census.json"

type censusFixture struct {
	Comment   string `json:"comment"`
	Criterion struct {
		A string `json:"A"`
		B string `json:"B"`
		C string `json:"C"`
	} `json:"criterion"`
	Counts struct {
		TotalVirtuals int `json:"total_virtuals"`
		CategoryA     int `json:"category_a"`
		CategoryB     int `json:"category_b"`
	} `json:"counts"`
	CategoryA []censusEntry `json:"category_a"`
}

type censusEntry struct {
	Class      string   `json:"class"`
	Method     string   `json:"method"`
	Signature  []string `json:"signature"`
	ReturnType string   `json:"return_type"`
}

func loadFixtureExtensionApi(t *testing.T) extensionapiparser.ExtensionApi {
	t.Helper()
	eapi, err := extensionapiparser.ParseExtensionApiJson(projectRoot)
	require.NoError(t, err)
	return eapi
}

func loadCensus(t *testing.T) censusFixture {
	t.Helper()
	body, err := os.ReadFile(censusFixturePath)
	require.NoError(t, err)
	var census censusFixture
	require.NoError(t, json.Unmarshal(body, &census))
	return census
}

// censusCriterionCNames extracts the Category-C names listed after the colon
// in the census criterion field.
func censusCriterionCNames(t *testing.T, census censusFixture) []string {
	t.Helper()
	_, names, found := strings.Cut(census.Criterion.C, ":")
	require.True(t, found, "census criterion C must list its names after a colon")
	var out []string
	for _, n := range strings.Split(names, ",") {
		n = strings.TrimSpace(n)
		if n != "" {
			out = append(out, n)
		}
	}
	return out
}

func virtualKey(class, method string) string {
	return fmt.Sprintf("%s::%s", class, method)
}

func TestVirtualCensusCountsMatchApiSurface(t *testing.T) {
	eapi := loadFixtureExtensionApi(t)
	census := loadCensus(t)

	totalVirtuals := 0
	for _, c := range eapi.Classes {
		for _, m := range c.Methods {
			if m.IsVirtual {
				totalVirtuals++
			}
		}
	}
	assert.Equal(t, census.Counts.TotalVirtuals, totalVirtuals, "census total_virtuals must equal the is_virtual count in extension_api.json")
	assert.Equal(t, len(census.CategoryA), census.Counts.CategoryA, "census category_a list must match its own count")
	assert.Equal(t, census.Counts.CategoryA+census.Counts.CategoryB, census.Counts.TotalVirtuals, "categories A and B must partition the total")

	// Category C names are creation-info routed and never appear in the API
	// surface as virtuals.
	cNames := map[string]bool{}
	for _, n := range censusCriterionCNames(t, census) {
		cNames[n] = true
	}
	require.NotEmpty(t, cNames)
	for _, c := range eapi.Classes {
		for _, m := range c.Methods {
			if m.IsVirtual {
				assert.Falsef(t, cNames[m.Name], "Category-C name %q must not appear as a virtual on class %s", m.Name, c.Name)
			}
		}
	}
}

func TestDerivedCategoryAMatchesCensus(t *testing.T) {
	eapi := loadFixtureExtensionApi(t)
	census := loadCensus(t)

	derived := map[string]bool{}
	for class, ms := range DeriveCategoryA(eapi) {
		for _, m := range ms {
			derived[virtualKey(class, m.Name)] = true
		}
	}

	censusKeys := map[string]bool{}
	for _, e := range census.CategoryA {
		require.Falsef(t, censusKeys[virtualKey(e.Class, e.Method)], "duplicate census entry %s", virtualKey(e.Class, e.Method))
		censusKeys[virtualKey(e.Class, e.Method)] = true
	}

	assert.Equal(t, len(censusKeys), len(derived), "criterion drift: derived Category-A count differs from the census")
	for k := range censusKeys {
		assert.Truef(t, derived[k], "census entry %s is not produced by the derivation criterion", k)
	}
	for k := range derived {
		assert.Truef(t, censusKeys[k], "derived entry %s is missing from the census", k)
	}
}

func TestGeneratedVirtualsMatchCensus(t *testing.T) {
	eapi := loadFixtureExtensionApi(t)
	census := loadCensus(t)

	rendered, err := renderVirtuals(eapi)
	require.NoError(t, err)
	output := string(rendered)

	// Declaration-only: the generated surface must never carry function bodies.
	assert.NotContains(t, output, "func ", "generated virtual-surface output must be declaration-only")

	// Parse the emitted interfaces back into class -> ordered method lines.
	emitted := map[string][]string{}
	var current string
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if rest, ok := strings.CutPrefix(line, "type "); ok {
			name, ok := strings.CutSuffix(rest, "Virtuals interface {")
			require.Truef(t, ok, "unexpected type declaration: %s", line)
			current = name
			require.NotContainsf(t, emitted, current, "duplicate interface for class %s", current)
			emitted[current] = nil
			continue
		}
		if current != "" {
			require.NotContainsf(t, line, "interface {", "nested interface in %s", current)
			if line == "}" {
				current = ""
				continue
			}
			require.Truef(t, strings.HasPrefix(line, "\t"), "method line outside interface: %q", line)
			emitted[current] = append(emitted[current], strings.TrimPrefix(line, "\t"))
		}
	}

	// Every emitted declaration must match, line for line, the declaration
	// derived from the API for that class (qualified name + exact signature).
	view := virtualsView{ExtensionApi: eapi, VirtualClasses: buildVirtualClassesView(eapi)}
	require.Len(t, view.VirtualClasses, len(emitted), "one interface per class with Category-A virtuals")
	for _, info := range view.VirtualClasses {
		var expected []string
		for _, m := range info.Methods {
			expected = append(expected, virtualMethodDecl(view, m))
		}
		assert.Equal(t, expected, emitted[info.StructName], "emitted interface for class %s must match the derived Category-A surface", info.StructName)
	}

	// Census <-> API: every census entry names a real virtual whose resolved
	// signature matches, and every census entry has an emitted counterpart
	// under the qualified convention (and vice versa, via the derived set).
	methodIndex := map[string]extensionapiparser.ClassMethod{}
	for _, c := range eapi.Classes {
		for _, m := range c.Methods {
			methodIndex[virtualKey(c.Name, m.Name)] = m
		}
	}
	emittedNames := map[string]bool{}
	for _, decls := range emitted {
		for _, d := range decls {
			emittedNames[strings.Fields(strings.SplitN(d, "(", 2)[0])[0]] = true
		}
	}
	for _, e := range census.CategoryA {
		m, ok := methodIndex[virtualKey(e.Class, e.Method)]
		require.Truef(t, ok, "census entry %s::%s not found in extension_api.json", e.Class, e.Method)
		assert.Truef(t, m.IsVirtual, "census entry %s::%s is not virtual", e.Class, e.Method)

		var argTypes []string
		for _, a := range m.Arguments {
			argTypes = append(argTypes, typeOrMeta(a.Meta, a.Type))
		}
		if argTypes == nil {
			argTypes = []string{}
		}
		sig := e.Signature
		if sig == nil {
			sig = []string{}
		}
		assert.Equal(t, sig, argTypes, "census signature for %s::%s must match the API", e.Class, e.Method)
		assert.Equal(t, e.ReturnType, typeOrMeta(m.ReturnValue.Meta, m.ReturnValue.Type), "census return type for %s::%s must match the API", e.Class, e.Method)

		qualified := fmt.Sprintf("V_%s_%s", e.Class, strcase.ToCamel(strings.TrimPrefix(e.Method, "_")))
		assert.Truef(t, emittedNames[qualified], "census entry %s::%s has no emitted declaration %s", e.Class, e.Method, qualified)
	}
	require.Len(t, emittedNames, len(census.CategoryA), "emitted declaration count must equal the census Category-A count")
}

// TestUpdateCensusFixture rewrites virtual_census.json from
// extension_api.json, keeping the file's comment and criterion text. It is
// skipped unless UPDATE_CENSUS=1 is set, so a plain `go test` never mutates
// the fixture. Run it after a header refresh whenever the completeness test
// reports drift.
func TestUpdateCensusFixture(t *testing.T) {
	if os.Getenv("UPDATE_CENSUS") == "" {
		t.Skip("set UPDATE_CENSUS=1 to regenerate virtual_census.json from extension_api.json")
	}
	eapi := loadFixtureExtensionApi(t)
	census := loadCensus(t)

	totalVirtuals := 0
	derived := DeriveCategoryA(eapi)
	for _, c := range eapi.Classes {
		for _, m := range c.Methods {
			if m.IsVirtual {
				totalVirtuals++
			}
		}
	}

	entries := []censusEntry{}
	for _, c := range eapi.Classes {
		for _, m := range derived[c.Name] {
			sig := []string{}
			for _, a := range m.Arguments {
				sig = append(sig, typeOrMeta(a.Meta, a.Type))
			}
			entries = append(entries, censusEntry{
				Class:      c.Name,
				Method:     m.Name,
				Signature:  sig,
				ReturnType: typeOrMeta(m.ReturnValue.Meta, m.ReturnValue.Type),
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Class != entries[j].Class {
			return entries[i].Class < entries[j].Class
		}
		return entries[i].Method < entries[j].Method
	})

	census.CategoryA = entries
	census.Counts.TotalVirtuals = totalVirtuals
	census.Counts.CategoryA = len(entries)
	census.Counts.CategoryB = totalVirtuals - len(entries)

	body, err := json.MarshalIndent(census, "", " ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(censusFixturePath, append(body, '\n'), 0o644))
	t.Logf("rewrote %s: %d Category-A entries, %d total virtuals", censusFixturePath, len(entries), totalVirtuals)
}
