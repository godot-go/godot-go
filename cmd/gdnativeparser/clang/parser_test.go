package clang

import (
	"testing"

	_ "embed"

	"github.com/stretchr/testify/require"
)

func TestParseMultiLineComment(t *testing.T) {
	content := `/* mycomment */`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.Equal(t, "/* mycomment */", f.Expr[0].Comment)

	require.NotEmpty(t, f.Expr[0].Comment)
}

func TestParseSingleLineComment(t *testing.T) {
	content := `// mycomment`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Len(t, f.Expr, 1)

	require.NotNil(t, f.Expr[0].Comment)

	require.Equal(t, "// mycomment", f.Expr[0].Comment)

	require.NotEmpty(t, f.Expr[0].Comment)
}

func TestParseTypedefEmptyEnum(t *testing.T) {
	content := `typedef enum { } MyFlags;`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Enum)

	require.Equal(t, "MyFlags", *f.Expr[0].Enum.Name)

	require.Empty(t, len(f.Expr[0].Enum.Values))
}

func TestParseTypedefEnum(t *testing.T) {
	content := `
	typedef enum {
		VALUE1,
		VALUE2,
		VALUE3,
	} MyFlags;`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Enum)

	require.Equal(t, "MyFlags", *f.Expr[0].Enum.Name)

	require.EqualValues(t, 3, len(*&f.Expr[0].Enum.Values))

	values := f.Expr[0].Enum.Values

	require.Equal(t, "VALUE1", values[0].Name)

	require.Equal(t, "VALUE2", values[1].Name)

	require.Equal(t, "VALUE3", values[2].Name)
}

func TestParseTypedefEnum2(t *testing.T) {
	content := `
	typedef enum {
		GDNATIVE_CALL_OK,
		GDNATIVE_CALL_ERROR_INVALID_METHOD,
		GDNATIVE_CALL_ERROR_INVALID_ARGUMENT, /* expected is variant type */
		GDNATIVE_CALL_ERROR_TOO_MANY_ARGUMENTS, /* expected is number of arguments */
		GDNATIVE_CALL_ERROR_TOO_FEW_ARGUMENTS, /*  expected is number of arguments */
		GDNATIVE_CALL_ERROR_INSTANCE_IS_NULL,
	} GDNativeCallErrorType;`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Enum)

	require.Equal(t, "GDNativeCallErrorType", *f.Expr[0].Enum.Name)

	require.EqualValues(t, 6, len(*&f.Expr[0].Enum.Values))

	values := f.Expr[0].Enum.Values

	require.Equal(t, "GDNATIVE_CALL_OK", values[0].Name)

	require.Equal(t, "GDNATIVE_CALL_ERROR_INVALID_METHOD", values[1].Name)

	require.Equal(t, "GDNATIVE_CALL_ERROR_INVALID_ARGUMENT", values[2].Name)

	require.Equal(t, "GDNATIVE_CALL_ERROR_TOO_MANY_ARGUMENTS", values[3].Name)

	require.Equal(t, "GDNATIVE_CALL_ERROR_TOO_FEW_ARGUMENTS", values[4].Name)

	require.Equal(t, "GDNATIVE_CALL_ERROR_INSTANCE_IS_NULL", values[5].Name)
}

func TestParseTypedefEnum3(t *testing.T) {
	content := `
	typedef enum {
		GDNATIVE_EXTENSION_METHOD_FLAG_NORMAL = 1,
		GDNATIVE_EXTENSION_METHOD_FLAG_EDITOR = 2,
		GDNATIVE_EXTENSION_METHOD_FLAG_NOSCRIPT = 4,
		GDNATIVE_EXTENSION_METHOD_FLAG_CONST = 8,
		GDNATIVE_EXTENSION_METHOD_FLAG_REVERSE = 16, /* used for events */
		GDNATIVE_EXTENSION_METHOD_FLAG_VIRTUAL = 32,
		GDNATIVE_EXTENSION_METHOD_FLAG_FROM_SCRIPT = 64,
		GDNATIVE_EXTENSION_METHOD_FLAG_VARARG = 128,
		GDNATIVE_EXTENSION_METHOD_FLAG_STATIC = 256,
		GDNATIVE_EXTENSION_METHOD_FLAGS_DEFAULT = GDNATIVE_EXTENSION_METHOD_FLAG_NORMAL,
	} GDNativeExtensionClassMethodFlags;`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Enum)

	require.Equal(t, "GDNativeExtensionClassMethodFlags", *f.Expr[0].Enum.Name)

	require.EqualValues(t, 10, len(f.Expr[0].Enum.Values))

	require.Equal(t, "GDNATIVE_EXTENSION_METHOD_FLAG_NORMAL", f.Expr[0].Enum.Values[0].Name)
	require.EqualValues(t, 1, *f.Expr[0].Enum.Values[0].IntValue)

	require.Equal(t, "GDNATIVE_EXTENSION_METHOD_FLAG_EDITOR", f.Expr[0].Enum.Values[1].Name)
	require.EqualValues(t, 2, *f.Expr[0].Enum.Values[1].IntValue)

	require.Equal(t, "GDNATIVE_EXTENSION_METHOD_FLAG_NOSCRIPT", f.Expr[0].Enum.Values[2].Name)
	require.EqualValues(t, 4, *f.Expr[0].Enum.Values[2].IntValue)

	require.Equal(t, "GDNATIVE_EXTENSION_METHOD_FLAG_CONST", f.Expr[0].Enum.Values[3].Name)
	require.EqualValues(t, 8, *f.Expr[0].Enum.Values[3].IntValue)

	require.Equal(t, "GDNATIVE_EXTENSION_METHOD_FLAG_REVERSE", f.Expr[0].Enum.Values[4].Name)
	require.EqualValues(t, 16, *f.Expr[0].Enum.Values[4].IntValue)

	require.Equal(t, "GDNATIVE_EXTENSION_METHOD_FLAG_VIRTUAL", f.Expr[0].Enum.Values[5].Name)
	require.EqualValues(t, 32, *f.Expr[0].Enum.Values[5].IntValue)

	require.Equal(t, "GDNATIVE_EXTENSION_METHOD_FLAG_FROM_SCRIPT", f.Expr[0].Enum.Values[6].Name)
	require.EqualValues(t, 64, *f.Expr[0].Enum.Values[6].IntValue)

	require.Equal(t, "GDNATIVE_EXTENSION_METHOD_FLAG_VARARG", f.Expr[0].Enum.Values[7].Name)
	require.EqualValues(t, 128, *f.Expr[0].Enum.Values[7].IntValue)

	require.Equal(t, "GDNATIVE_EXTENSION_METHOD_FLAG_STATIC", f.Expr[0].Enum.Values[8].Name)
	require.EqualValues(t, 256, *f.Expr[0].Enum.Values[8].IntValue)

	require.Equal(t, "GDNATIVE_EXTENSION_METHOD_FLAGS_DEFAULT", f.Expr[0].Enum.Values[9].Name)
	require.Equal(t, "GDNATIVE_EXTENSION_METHOD_FLAG_NORMAL", *f.Expr[0].Enum.Values[9].ConstRefValue)
}

func TestParseTypedefEmptyStruct(t *testing.T) {
	content := `typedef struct { } MyStruct;`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Struct)

	require.Equal(t, "MyStruct", f.Expr[0].Struct.Name)

	require.Empty(t, len(f.Expr[0].Struct.Fields))
}

func TestParseTypedefStruct(t *testing.T) {
	content := `typedef struct {
		GDNativeCallErrorType error;
		int32_t argument;
		int32_t expected;

		const char *some_other_pointer_string;
	} GDNativeCallError;`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Struct)

	require.Equal(t, "GDNativeCallError", f.Expr[0].Struct.Name)

	require.Len(t, f.Expr[0].Struct.Fields, 4)

	require.False(t, *&f.Expr[0].Struct.Fields[2].Variable.Type.IsConst)
	require.Equal(t, "GDNativeCallErrorType", *&f.Expr[0].Struct.Fields[0].Variable.Type.Name)
	require.False(t, *&f.Expr[0].Struct.Fields[0].Variable.Type.IsPointer)
	require.Equal(t, "error", *&f.Expr[0].Struct.Fields[0].Variable.Name)

	require.False(t, *&f.Expr[0].Struct.Fields[2].Variable.Type.IsConst)
	require.Equal(t, "int32_t", *&f.Expr[0].Struct.Fields[1].Variable.Type.Name)
	require.False(t, *&f.Expr[0].Struct.Fields[1].Variable.Type.IsPointer)
	require.Equal(t, "argument", *&f.Expr[0].Struct.Fields[1].Variable.Name)

	require.False(t, *&f.Expr[0].Struct.Fields[2].Variable.Type.IsConst)
	require.Equal(t, "int32_t", *&f.Expr[0].Struct.Fields[2].Variable.Type.Name)
	require.False(t, *&f.Expr[0].Struct.Fields[2].Variable.Type.IsPointer)
	require.Equal(t, "expected", *&f.Expr[0].Struct.Fields[2].Variable.Name)

	require.True(t, *&f.Expr[0].Struct.Fields[3].Variable.Type.IsConst)
	require.Equal(t, "char", *&f.Expr[0].Struct.Fields[3].Variable.Type.Name)
	require.True(t, *&f.Expr[0].Struct.Fields[3].Variable.Type.IsPointer)
	require.Equal(t, "some_other_pointer_string", *&f.Expr[0].Struct.Fields[3].Variable.Name)
}

func TestParseTypedefFunction(t *testing.T) {
	content := `typedef void (*GDNativePtrOperatorEvaluator)(const GDNativeTypePtr p_left, const GDNativeTypePtr p_right, GDNativeTypePtr r_result);`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Function)

	require.Equal(t, "GDNativePtrOperatorEvaluator", f.Expr[0].Function.Name)

	require.Len(t, f.Expr[0].Function.Arguments, 3)
}

func TestParseTypedefFunction2(t *testing.T) {
	content := `typedef const GDNativePropertyInfo *(*GDNativeExtensionClassGetPropertyList)(GDExtensionClassInstancePtr p_instance, uint32_t *r_count);`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Function)

	require.Equal(t, "GDNativeExtensionClassGetPropertyList", f.Expr[0].Function.Name)

	require.Len(t, f.Expr[0].Function.Arguments, 2)
}

func TestParseTypedefFunctionNoArgumentNames(t *testing.T) {
	content := `typedef void (*GDNativeVariantFromTypeConstructorFunc)(GDNativeVariantPtr, GDNativeTypePtr);`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Function)

	require.Equal(t, "GDNativeVariantFromTypeConstructorFunc", f.Expr[0].Function.Name)

	require.Len(t, f.Expr[0].Function.Arguments, 2)
}
