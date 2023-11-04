package ffi

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	projectPath := os.Getenv("VSCODE_WORKSPACE_FOLDER")

	require.NotEmpty(t, projectPath)

	Generate(projectPath, "")
}
