package ffi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/godot-go/godot-go/cmd/gdextensionparser/clang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	// Generate writes real gen files under projectPath, so run it against a
	// throwaway tree; pointing it at the repository would overwrite the
	// committed generated files with empty-AST output.
	projectPath := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(projectPath, "pkg", "ffi"), 0o755))

	ast := clang.CHeaderFileAST{
		Expr: []clang.Expr{},
	}
	var panicFunc assert.PanicTestFunc = func() {
		Generate(projectPath, ast)
	}
	require.NotPanics(t, panicFunc)

	for _, name := range []string{
		"ffi_wrapper.gen.h",
		"ffi_wrapper.gen.c",
		"ffi_wrapper.gen.go",
		"ffi.gen.go",
	} {
		require.FileExists(t, filepath.Join(projectPath, "pkg", "ffi", name))
	}
}
