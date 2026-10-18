package gdclassimpl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func varargApi(methods ...extensionapiparser.ClassMethod) extensionapiparser.ExtensionApi {
	return extensionapiparser.ExtensionApi{
		Classes: []extensionapiparser.Class{
			{Name: "Widget", Methods: methods},
		},
	}
}

func TestCheckVarargReturnTypesRejectsUnsupported(t *testing.T) {
	eapi := varargApi(extensionapiparser.ClassMethod{
		Name:        "bounce_back",
		IsVararg:    true,
		ReturnValue: extensionapiparser.ClassMethodReturnValue{Type: "Vector2"},
	})

	err := checkVarargReturnTypes(eapi)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Widget.bounce_back")
	assert.Contains(t, err.Error(), "Vector2")
}

func TestCheckVarargReturnTypesAcceptsSupportedSurface(t *testing.T) {
	eapi := varargApi(
		extensionapiparser.ClassMethod{Name: "emit", IsVararg: true,
			ReturnValue: extensionapiparser.ClassMethodReturnValue{Type: "enum::Error"}},
		extensionapiparser.ClassMethod{Name: "call_it", IsVararg: true,
			ReturnValue: extensionapiparser.ClassMethodReturnValue{Type: "Variant"}},
		extensionapiparser.ClassMethod{Name: "say", IsVararg: true,
			ReturnValue: extensionapiparser.ClassMethodReturnValue{Type: "void"}},
		// Non-vararg methods keep the ptrcall path with a declared-type
		// buffer and are not constrained by the vararg conversion table.
		extensionapiparser.ClassMethod{Name: "get_size",
			ReturnValue: extensionapiparser.ClassMethodReturnValue{Type: "Vector2"}},
	)

	assert.NoError(t, checkVarargReturnTypes(eapi))
}

func TestVarargReturnExpr(t *testing.T) {
	expr, ok := varargReturnConversion("Error")
	require.True(t, ok)
	assert.Equal(t, "Error(retVariant.ToInt64())", expr)

	expr, ok = varargReturnConversion("Variant")
	require.True(t, ok)
	assert.Equal(t, "", expr, "Variant returns hand the buffer back untouched")

	_, ok = varargReturnConversion("Vector2")
	assert.False(t, ok)

	assert.PanicsWithValue(t, "unsupported vararg return type Vector2", func() {
		varargReturnExpr("Vector2")
	})
}

// TestGenerateClassesVarargBuffer renders a minimal class through the real
// generator and checks the vararg return shapes: Error converts from a
// Variant buffer, Variant is returned untouched, void passes no buffer.
func TestGenerateClassesVarargBuffer(t *testing.T) {
	eapi := varargApi(
		extensionapiparser.ClassMethod{Name: "emit_signal", IsVararg: true,
			ReturnValue: extensionapiparser.ClassMethodReturnValue{Type: "enum::Error"}},
		extensionapiparser.ClassMethod{Name: "call_it", IsVararg: true,
			ReturnValue: extensionapiparser.ClassMethodReturnValue{Type: "Variant"}},
		extensionapiparser.ClassMethod{Name: "say", IsVararg: true,
			ReturnValue: extensionapiparser.ClassMethodReturnValue{Type: "void"}},
	)

	out := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(out, "pkg", "gdclassimpl"), 0o755))
	require.NoError(t, GenerateClasses(out, eapi))

	body, err := os.ReadFile(filepath.Join(out, "pkg", "gdclassimpl", "classes.gen.go"))
	require.NoError(t, err)
	code := string(body)

	// funcBody returns the generated method text from its signature up to
	// the next top-level func declaration.
	funcBody := func(sig string) string {
		start := strings.Index(code, sig)
		require.NotEqualf(t, -1, start, "generated code missing %s", sig)
		if end := strings.Index(code[start:], "\nfunc "); end != -1 {
			return code[start : start+end]
		}
		return code[start:]
	}

	emit := funcBody("func (cx *WidgetImpl) EmitSignal(")
	assert.Contains(t, emit, "var retVariant Variant", "Error-returning vararg must buffer a Variant")
	assert.NotContains(t, emit, "var ret Error", "Error-returning vararg must not buffer an Error")
	assert.Contains(t, emit, "defer retVariant.Destroy()")
	assert.Contains(t, emit, "return Error(retVariant.ToInt64())")

	call := funcBody("func (cx *WidgetImpl) CallIt(")
	assert.Contains(t, call, "var ret Variant", "Variant-returning vararg keeps the Variant buffer named ret")
	assert.Contains(t, call, "return ret")

	say := funcBody("func (cx *WidgetImpl) Say(")
	assert.Contains(t, say, "retPtr := (GDExtensionUninitializedVariantPtr)(nullptr)", "void vararg passes no buffer")
}

// TestGenerateClassesFailsOnUnsupportedVararg checks the guard stops
// generation before the template renders an undersized buffer.
func TestGenerateClassesFailsOnUnsupportedVararg(t *testing.T) {
	eapi := varargApi(extensionapiparser.ClassMethod{
		Name:        "bounce_back",
		IsVararg:    true,
		ReturnValue: extensionapiparser.ClassMethodReturnValue{Type: "Vector2"},
	})

	out := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(out, "pkg", "gdclassimpl"), 0o755))

	err := GenerateClasses(out, eapi)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Widget.bounce_back")
}
