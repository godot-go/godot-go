package gdextensionparser

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

// projectRoot is the repository root relative to this package directory.
const projectRoot = "../.."

func TestGenerateGDExtensionInterfaceAST(t *testing.T) {
	ast, err := GenerateGDExtensionInterfaceAST(projectRoot, "")
	require.NoError(t, err)
	spew.Dump(ast)
}
