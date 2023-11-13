package gdextensionparser

import (
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestGenerateGDExtensionInterfaceAST(t *testing.T) {
	projectPath := os.Getenv("VSCODE_WORKSPACE_FOLDER")
	require.NotEmpty(t, projectPath)
	f, err := GenerateGDExtensionInterfaceAST(projectPath, "")
	require.NoError(t, err)
	spew.Dump(f)
}
