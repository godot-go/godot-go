package core

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// These tests enforce the variant-copy-ownership rule: a value produced by
// builtin.VariantViewFromConstPtr is a non-owning view of engine-owned Variant
// storage, and destroying one frees memory the source is still using.
//
// The check scans source rather than running the engine. A Variant's reference
// count is not observable from Go -- Array and Dictionary do not expose the
// underlying count -- and deliberately destroying a borrowed view to demonstrate
// the hazard would produce a crash, not a test result. Source scanning catches
// the edit that matters, which is routing a borrowed view into an owning path.
//
// Matching is scoped to the block the view was created in, not the whole file.
// Names like `v` are reused freely across a function's switch branches, and a
// file-wide name match flags every unrelated Destroy in the file. The hazard being
// guarded is make-a-view-then-destroy-it, which lives in one scope.
//
// Known limit, accepted deliberately: a view returned from one function, rebound
// elsewhere, and destroyed there is outside the window this looks at. It catches
// the direct mistake, which is the one that was actually made.

var (
	// viewAssign captures the target of an assignment whose value is a view.
	// The target may be a plain identifier or an indexed element such as
	// args[i], because marshalling fills a slice element-wise.
	viewAssign = regexp.MustCompile(
		"(?m)^(?:[ \\t]*)(?:var[ \\t]+)?([A-Za-z_][\\w]*(?:\\[[^\\]]*\\])?)[ \\t]*(?::=|=)[ \\t]*(?:\\w+\\.)?VariantViewFromConstPtr\\(")

	// destroyCall captures the receiver of a Destroy() call.
	destroyCall = regexp.MustCompile(`([A-Za-z_][\w]*(?:\[[^\]]*\])?)\.Destroy\(\)`)

	// rangeOver captures the collection in a range statement.
	rangeOver = regexp.MustCompile(`\brange\s+([A-Za-z_][\w]*)\b`)

	// nextCaseClause marks the start of a sibling switch clause. Go's case labels
	// do not open a brace-delimited block, so a brace-only window would run past
	// the clause that created the view and pick up unrelated destroys from its
	// siblings.
	nextCaseClause = regexp.MustCompile(`(?m)^[ \t]*(?:case\b|default:)`)
)

// viewSourceFiles returns the non-test Go sources of this package.
func viewSourceFiles(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package directory: %v", err)
	}
	var files []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		files = append(files, name)
	}
	if len(files) == 0 {
		t.Fatal("scanned no source files; the check would pass vacuously")
	}
	return files
}

func readSource(t *testing.T, file string) string {
	t.Helper()

	src, err := os.ReadFile(filepath.Join(".", file))
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	return string(src)
}

// baseIdent strips an index from an assignment target: args[i] -> args.
func baseIdent(target string) string {
	if i := strings.IndexByte(target, '['); i >= 0 {
		return target[:i]
	}
	return target
}

// scopeAfter returns the text from `from` up to the point where the block
// enclosing that position closes. Brace depth is tracked from `from`; the window
// ends when depth would go negative, which is the closing brace of the scope the
// assignment lives in.
func scopeAfter(text string, from int) string {
	depth := 0
	for i := from; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			if depth == 0 {
				return text[from:i]
			}
			depth--
		}
	}
	return text[from:]
}

// viewScope returns the region a view assignment can affect: from the assignment
// to the end of its clause. The clause ends at the closing brace of the
// enclosing block, or at the next sibling switch clause, whichever comes first.
// Truncating early only narrows what is checked, so a misjudged boundary costs
// coverage rather than producing a false accusation.
func viewScope(text string, from int) string {
	scope := scopeAfter(text, from)
	if cut := nextCaseClause.FindStringIndex(scope); cut != nil {
		return scope[:cut[0]]
	}
	return scope
}

// TestVariantViewsAreNeverDestroyed is the guard the rename exists to make
// enforceable: nothing produced as a borrowed view may be destroyed.
func TestVariantViewsAreNeverDestroyed(t *testing.T) {
	produced := 0

	for _, file := range viewSourceFiles(t) {
		text := readSource(t, file)

		for _, m := range viewAssign.FindAllStringSubmatchIndex(text, -1) {
			produced++
			target := text[m[2]:m[3]]
			base := baseIdent(target)
			scope := viewScope(text, m[1])

			for _, d := range destroyCall.FindAllStringSubmatch(scope, -1) {
				receiver := d[1]
				if receiver == target {
					t.Errorf("%s: %s is a non-owning Variant view and must not be destroyed", file, receiver)
				} else if receiver == base {
					t.Errorf("%s: %s holds non-owning Variant views and must not be destroyed", file, receiver)
				}
			}

			// The indirect form: a range loop over the view-holding collection
			// that destroys what it ranges.
			for _, r := range rangeOver.FindAllStringSubmatch(scope, -1) {
				if r[1] != base {
					continue
				}
				body := blockBody(scope, r[0])
				if body != "" && destroyCall.MatchString(body) {
					t.Errorf("%s: range over %s destroys a non-owning Variant view", file, r[1])
				}
			}
		}
	}

	if produced == 0 {
		t.Fatal("no VariantViewFromConstPtr call sites found; the check would pass vacuously")
	}
}

// TestViewHelperIsNamedForWhatItIs guards the rename itself. The pre-rename name
// read as a constructor while returning a borrowed view; if it comes back, the
// confusion this change removed comes back with it.
func TestViewHelperIsNamedForWhatItIs(t *testing.T) {
	for _, file := range viewSourceFiles(t) {
		if strings.Contains(readSource(t, file), "NewVariantCopyWithGDExtensionConstVariantPtr") {
			t.Errorf("%s: the pre-rename helper name is back; it read as a constructor "+
				"while returning a non-owning view, which is the confusion this "+
				"change removed", file)
		}
	}
}

// blockBody returns the brace-delimited body that follows match, or "" when the
// match is not followed by a balanced block.
func blockBody(text, match string) string {
	start := strings.Index(text, match)
	if start < 0 {
		return ""
	}
	open := strings.IndexByte(text[start:], '{')
	if open < 0 {
		return ""
	}
	bodyStart := start + open
	depth := 0
	for i := bodyStart; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[bodyStart : i+1]
			}
		}
	}
	return ""
}
