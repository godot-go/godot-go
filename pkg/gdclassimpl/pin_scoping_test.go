package gdclassimpl

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// These tests enforce the pin-lifetime rule from the native-call-pinning spec:
// per-call scratch is pinned on a body-scoped runtime.Pinner that is released by
// defer Unpin(), and the package-level pinner is used only for program-lifetime
// retention. They scan source rather than run the engine so the regression is
// caught by a plain `go test`.

type scannedFunc struct {
	name string
	body string
}

// splitTopLevelFuncs splits Go source into top-level function declarations by
// brace-matching from each line that begins with "func " in column 0.
func splitTopLevelFuncs(t *testing.T, path string) []scannedFunc {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	lines := strings.Split(string(data), "\n")

	var funcs []scannedFunc
	var current []string
	depth := 0
	inFunc := false
	name := ""

	funcSig := regexp.MustCompile(`^func (?:\([^)]*\) )?([A-Za-z0-9_]+)`)

	for _, line := range lines {
		if !inFunc {
			m := funcSig.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			inFunc = true
			name = m[1]
			current = []string{line}
			depth = strings.Count(line, "{") - strings.Count(line, "}")
			if depth <= 0 {
				funcs = append(funcs, scannedFunc{name, strings.Join(current, "\n")})
				inFunc = false
			}
			continue
		}

		current = append(current, line)
		depth += strings.Count(line, "{") - strings.Count(line, "}")
		if depth <= 0 {
			funcs = append(funcs, scannedFunc{name, strings.Join(current, "\n")})
			inFunc = false
		}
	}
	return funcs
}

// instanceCreationFunc matches the generated per-class constructors, which pin the
// wrapper for as long as the engine holds it. That is program-lifetime retention
// by design, not call scratch.
var instanceCreationFunc = regexp.MustCompile(`^(NewGDExtensionClassFrom.*Owner|New.*WithGodotOwnerObject)$`)

// generatedCallPaths are the generated files whose method, constructor, utility and
// variant-conversion bodies must scope their pins.
var generatedCallPaths = []string{
	"pkg/gdclassimpl/classes.gen.go",
	"pkg/gdutilfunc/utilityfunctions.gen.go",
	"pkg/builtin/builtinclasses.gen.go",
	"pkg/builtin/variant.gen.go",
}

// handWrittenCallPaths are hand-written per-call helpers that must scope their pins.
var handWrittenCallPaths = []string{
	"pkg/ffi/builtin_ptrcall.go",
	"pkg/core/method_bind_callback.go",
	"pkg/builtin/char_string.go",
	"pkg/builtin/variant.go",
	"pkg/builtin/container_copy_constructors.go",
	"pkg/builtin/wrapped.go",
	"pkg/core/classdb_callback.go",
	"pkg/core/method_bind.go",
}

// packagePinnerAllowlist names functions that legitimately pin on the package
// pinner inside an otherwise per-call file, with the reason they outlive the call.
//
//   - ObjectCastTo: the instance-binding callbacks pointer it hands the engine is
//     stored with the binding and read again after the call returns.
//   - GoCallback_ClassCreationInfoGetPropertyList: the engine keeps the returned
//     property-info slice until the paired free-property-list callback runs.
//   - GoMethodMetadata.Destroy / NewGoMethodMetadata /
//     NewGDExtensionClassMethodInfoFromMethodBind: registration and teardown of
//     metadata the engine holds pointers to.
var packagePinnerAllowlist = map[string]string{
	"ObjectCastTo": "binding callbacks retained by the engine with the binding",
	"GoCallback_ClassCreationInfoGetPropertyList": "property list retained until the paired free callback",
	"Destroy":             "metadata teardown",
	"NewGoMethodMetadata": "registration-time metadata",
	"NewGDExtensionClassMethodInfoFromMethodBind": "registration-time method info",

	// Registration stores these pointers inside Go structs (GDExtensionPropertyInfo
	// and friends) that are then handed to C. cgocheck rejects a cgo argument
	// whose pointee contains an unpinned Go pointer, so the targets must stay
	// pinned for at least as long as those structs can be crossed. The generated
	// per-call path uses the *Pinned variants instead.
	"AsGDExtensionConstStringNamePtr": "reachable through nested Go pointers in registration structs",
	"AsGDExtensionConstStringPtr":     "reachable through nested Go pointers in registration structs",
}

// pinsThroughCaller reports whether a function takes the pinner as a parameter,
// in which case the caller owns the defer and the function legitimately pins
// without releasing.
func pinsThroughCaller(fn scannedFunc) bool {
	return strings.Contains(fn.body, "*runtime.Pinner")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate module root containing go.mod")
		}
		dir = parent
	}
}

// TestGeneratedCallBodiesScopeTheirPins asserts that every generated call body
// that pins does so on a local pinner it releases, and that the package pinner
// appears only in the instance-creation constructors.
func TestGeneratedCallBodiesScopeTheirPins(t *testing.T) {
	root := repoRoot(t)

	for _, rel := range generatedCallPaths {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			for _, fn := range splitTopLevelFuncs(t, filepath.Join(root, rel)) {
				usesScoped := strings.Contains(fn.body, "pinner.Pin(")
				hasDefer := strings.Contains(fn.body, "defer pinner.Unpin()")

				if usesScoped && !hasDefer && !pinsThroughCaller(fn) {
					t.Errorf("%s: %s pins on a scoped pinner but never defers Unpin()", rel, fn.name)
				}
				// The reverse (defers but never pins) is deliberately not an
				// error: the template declares the pinner uniformly at the top of
				// every body, and bodies whose branch needs no pin leave it empty.
				// Unpin() on a pinner holding nothing is a no-op.

				if strings.Contains(fn.body, "pnr.Pin(") && !instanceCreationFunc.MatchString(fn.name) {
					t.Errorf(
						"%s: %s pins on the package pinner; generated call bodies must use a "+
							"body-scoped runtime.Pinner instead",
						rel, fn.name,
					)
				}
			}
		})
	}
}

// TestHandWrittenCallHelpersScopeTheirPins asserts the same for the hand-written
// per-call helpers, allowing only the documented engine-retention cases.
func TestHandWrittenCallHelpersScopeTheirPins(t *testing.T) {
	root := repoRoot(t)

	for _, rel := range handWrittenCallPaths {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			for _, fn := range splitTopLevelFuncs(t, filepath.Join(root, rel)) {
				usesScoped := strings.Contains(fn.body, "pinner.Pin(")
				hasDefer := strings.Contains(fn.body, "defer pinner.Unpin()")

				if usesScoped && !hasDefer && !pinsThroughCaller(fn) {
					t.Errorf("%s: %s pins on a scoped pinner but never defers Unpin()", rel, fn.name)
				}

				if !strings.Contains(fn.body, "pnr.Pin(") {
					continue
				}
				if _, allowed := packagePinnerAllowlist[fn.name]; allowed {
					continue
				}
				t.Errorf(
					"%s: %s pins on the package pinner; per-call helpers must use a "+
						"body-scoped runtime.Pinner, or be added to packagePinnerAllowlist "+
						"with the reason the value outlives the call",
					rel, fn.name,
				)
			}
		})
	}
}

// TestScopedPinnersOutnumberPackagePins guards against the fix being reverted en
// masse: the whole point of the change is that call scratch vastly outnumbers the
// small, bounded set of program-lifetime pins.
func TestScopedPinnersOutnumberPackagePins(t *testing.T) {
	root := repoRoot(t)

	scoped, packagePinned := 0, 0
	for _, rel := range append(append([]string{}, generatedCallPaths...), handWrittenCallPaths...) {
		src, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		scoped += strings.Count(string(src), "pinner.Pin(")
		packagePinned += strings.Count(string(src), "pnr.Pin(")
	}

	if scoped == 0 {
		t.Fatal("no scoped pins found anywhere; the per-call pinner pattern is missing")
	}
	if scoped <= packagePinned {
		t.Errorf(
			"expected scoped call pins to dominate program-lifetime pins, got %d scoped vs %d package",
			scoped, packagePinned,
		)
	}
	t.Logf("scoped call pins: %d, program-lifetime pins: %d", scoped, packagePinned)
}
