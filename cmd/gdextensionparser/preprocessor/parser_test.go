package preprocessor

import (
	"testing"

	_ "embed"

	"github.com/stretchr/testify/require"
)

func TestPreprocessorParseHeaderFile(t *testing.T) {
	content := `/*******
* test *
********/

#ifndef MYFILE_H
#define MYFILE_H
#include <test.h>

#ifndef __cplusplus
typedef uint32_t char32_t;
typedef uint16_t char16_t;
#endif

#ifdef __cplusplus
extern "C" {
#endif

typedef void *GDExtensionVariantPtr;

#ifdef __cplusplus
}
#endif
#endif // MYFILE_H
`

	ast, err := ParsePreprocessorString(content)

	require.NoError(t, err)
	require.Len(t, ast.Directives, 1)
	require.NotNil(t, ast.Directives[0].Ifndef)
	require.Equal(t, "MYFILE_H", ast.Directives[0].Ifndef.Name)
	require.Len(t, ast.Directives[0].Ifndef.Directives, 6)
	require.NotNil(t, ast.Directives[0].Ifndef.Directives[0].Define)
	require.Equal(t, "MYFILE_H", ast.Directives[0].Ifndef.Directives[0].Define.Name)
	require.NotNil(t, ast.Directives[0].Ifndef.Directives[1].Include)
	require.NotNil(t, ast.Directives[0].Ifndef.Directives[2].Ifndef)
	require.Equal(t, "__cplusplus", ast.Directives[0].Ifndef.Directives[2].Ifndef.Name)
	require.NotNil(t, ast.Directives[0].Ifndef.Directives[3].Ifdef)
	require.Equal(t, "__cplusplus", ast.Directives[0].Ifndef.Directives[3].Ifdef.Name)
	require.Len(t, ast.Directives[0].Ifndef.Directives[3].Ifdef.Directives, 1)
	require.Equal(t, "extern \"C\" {\n", ast.Directives[0].Ifndef.Directives[3].Ifdef.Directives[0].Source)
	require.Equal(t, "typedef void *GDExtensionVariantPtr;\n\n", ast.Directives[0].Ifndef.Directives[4].Source)
	require.NotNil(t, ast.Directives[0].Ifndef.Directives[5].Ifdef)
	require.Len(t, ast.Directives[0].Ifndef.Directives[5].Ifdef.Directives, 1)
	require.Equal(t, "__cplusplus", ast.Directives[0].Ifndef.Directives[5].Ifdef.Name)
	require.Equal(t, "}\n", ast.Directives[0].Ifndef.Directives[5].Ifdef.Directives[0].Source)
	require.Equal(t, "\n\ntypedef uint32_t char32_t;\ntypedef uint16_t char16_t;\n\n\n\ntypedef void *GDExtensionVariantPtr;\n\n\n\n\n", ast.Eval(false))
}

func TestParseCommentRegression(t *testing.T) {
	content := `/*******
	* test *
	********/

	#ifndef REGRESSION_H
	/* misc types */
	const i int = 5;
	const j int = 6;
	const k int = 7;

	/* typed arrays */
	const x int = 1;
	const y int = 2;

	#endif // REGRESSION_H
`
	ast, err := ParsePreprocessorString(content)

	require.NoError(t, err)

	require.Len(t, ast.Directives, 1)
	require.NotNil(t, ast.Directives[0].Ifndef)
	require.Equal(t, "REGRESSION_H", ast.Directives[0].Ifndef.Name)
	require.Len(t, ast.Directives[0].Ifndef.Directives, 1)
	require.Equal(t, "const i int = 5;\n\tconst j int = 6;\n\tconst k int = 7;\n\n\t/* typed arrays */\n\tconst x int = 1;\n\tconst y int = 2;\n\n\t", ast.Directives[0].Ifndef.Directives[0].Source)
	require.Equal(t, "const i int = 5;\n\tconst j int = 6;\n\tconst k int = 7;\n\n\t/* typed arrays */\n\tconst x int = 1;\n\tconst y int = 2;\n\n\t\n\n", ast.Eval(false))
}
