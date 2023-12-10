package ffi

import (
	"os"
	"testing"

	"github.com/godot-go/godot-go/cmd/gdextensionparser/clang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	projectPath := os.Getenv("VSCODE_WORKSPACE_FOLDER")
	require.NotEmpty(t, projectPath)
	ast := clang.CHeaderFileAST{
		Expr: []clang.Expr{},
	}
	var panicFunc assert.PanicTestFunc = func() {
		Generate(projectPath, ast)
	}
	require.NotPanics(t, panicFunc)
}
