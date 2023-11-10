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
		GDEXTENSION_CALL_OK,
		GDEXTENSION_CALL_ERROR_INVALID_METHOD,
		GDEXTENSION_CALL_ERROR_INVALID_ARGUMENT, /* expected is variant type */
		GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS, /* expected is number of arguments */
		GDEXTENSION_CALL_ERROR_TOO_FEW_ARGUMENTS, /*  expected is number of arguments */
		GDEXTENSION_CALL_ERROR_INSTANCE_IS_NULL,
	} GDExtensionCallErrorType;`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Enum)

	require.Equal(t, "GDExtensionCallErrorType", *f.Expr[0].Enum.Name)

	require.EqualValues(t, 6, len(*&f.Expr[0].Enum.Values))

	values := f.Expr[0].Enum.Values

	require.Equal(t, "GDEXTENSION_CALL_OK", values[0].Name)

	require.Equal(t, "GDEXTENSION_CALL_ERROR_INVALID_METHOD", values[1].Name)

	require.Equal(t, "GDEXTENSION_CALL_ERROR_INVALID_ARGUMENT", values[2].Name)

	require.Equal(t, "GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS", values[3].Name)

	require.Equal(t, "GDEXTENSION_CALL_ERROR_TOO_FEW_ARGUMENTS", values[4].Name)

	require.Equal(t, "GDEXTENSION_CALL_ERROR_INSTANCE_IS_NULL", values[5].Name)
}

func TestParseTypedefEnum3(t *testing.T) {
	content := `
	typedef enum {
		GDEXTENSION_EXTENSION_METHOD_FLAG_NORMAL = 1,
		GDEXTENSION_EXTENSION_METHOD_FLAG_EDITOR = 2,
		GDEXTENSION_EXTENSION_METHOD_FLAG_NOSCRIPT = 4,
		GDEXTENSION_EXTENSION_METHOD_FLAG_CONST = 8,
		GDEXTENSION_EXTENSION_METHOD_FLAG_REVERSE = 16, /* used for events */
		GDEXTENSION_EXTENSION_METHOD_FLAG_VIRTUAL = 32,
		GDEXTENSION_EXTENSION_METHOD_FLAG_FROM_SCRIPT = 64,
		GDEXTENSION_EXTENSION_METHOD_FLAG_VARARG = 128,
		GDEXTENSION_EXTENSION_METHOD_FLAG_STATIC = 256,
		GDEXTENSION_EXTENSION_METHOD_FLAGS_DEFAULT = GDEXTENSION_EXTENSION_METHOD_FLAG_NORMAL,
	} GDExtensionClassMethodFlags;`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Enum)

	require.Equal(t, "GDExtensionClassMethodFlags", *f.Expr[0].Enum.Name)

	require.EqualValues(t, 10, len(f.Expr[0].Enum.Values))

	require.Equal(t, "GDEXTENSION_EXTENSION_METHOD_FLAG_NORMAL", f.Expr[0].Enum.Values[0].Name)
	require.EqualValues(t, 1, *f.Expr[0].Enum.Values[0].IntValue)

	require.Equal(t, "GDEXTENSION_EXTENSION_METHOD_FLAG_EDITOR", f.Expr[0].Enum.Values[1].Name)
	require.EqualValues(t, 2, *f.Expr[0].Enum.Values[1].IntValue)

	require.Equal(t, "GDEXTENSION_EXTENSION_METHOD_FLAG_NOSCRIPT", f.Expr[0].Enum.Values[2].Name)
	require.EqualValues(t, 4, *f.Expr[0].Enum.Values[2].IntValue)

	require.Equal(t, "GDEXTENSION_EXTENSION_METHOD_FLAG_CONST", f.Expr[0].Enum.Values[3].Name)
	require.EqualValues(t, 8, *f.Expr[0].Enum.Values[3].IntValue)

	require.Equal(t, "GDEXTENSION_EXTENSION_METHOD_FLAG_REVERSE", f.Expr[0].Enum.Values[4].Name)
	require.EqualValues(t, 16, *f.Expr[0].Enum.Values[4].IntValue)

	require.Equal(t, "GDEXTENSION_EXTENSION_METHOD_FLAG_VIRTUAL", f.Expr[0].Enum.Values[5].Name)
	require.EqualValues(t, 32, *f.Expr[0].Enum.Values[5].IntValue)

	require.Equal(t, "GDEXTENSION_EXTENSION_METHOD_FLAG_FROM_SCRIPT", f.Expr[0].Enum.Values[6].Name)
	require.EqualValues(t, 64, *f.Expr[0].Enum.Values[6].IntValue)

	require.Equal(t, "GDEXTENSION_EXTENSION_METHOD_FLAG_VARARG", f.Expr[0].Enum.Values[7].Name)
	require.EqualValues(t, 128, *f.Expr[0].Enum.Values[7].IntValue)

	require.Equal(t, "GDEXTENSION_EXTENSION_METHOD_FLAG_STATIC", f.Expr[0].Enum.Values[8].Name)
	require.EqualValues(t, 256, *f.Expr[0].Enum.Values[8].IntValue)

	require.Equal(t, "GDEXTENSION_EXTENSION_METHOD_FLAGS_DEFAULT", f.Expr[0].Enum.Values[9].Name)
	require.Equal(t, "GDEXTENSION_EXTENSION_METHOD_FLAG_NORMAL", *f.Expr[0].Enum.Values[9].ConstRefValue)
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
		GDExtensionCallErrorType error;
		int32_t argument;
		int32_t expected;

		const char *some_other_pointer_string;

		/*  extra utilities */
		void (*string_new_with_latin1_chars)(GDExtensionStringPtr r_dest, const char *p_contents);

		int64_t (*worker_thread_pool_add_native_group_task)(GDExtensionObjectPtr p_instance, void (*p_func)(void *, uint32_t), void *p_userdata, int p_elements, int p_tasks, GDExtensionBool p_high_priority, GDExtensionConstStringPtr p_description);
	} GDExtensionCallError;`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Struct)

	require.Equal(t, "GDExtensionCallError", f.Expr[0].Struct.Name)

	require.Len(t, f.Expr[0].Struct.Fields, 6)

	require.False(t, *&f.Expr[0].Struct.Fields[2].Variable.Type.IsConst)
	require.Equal(t, "GDExtensionCallErrorType", *&f.Expr[0].Struct.Fields[0].Variable.Type.Name)
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

	require.Equal(t, "void", *&f.Expr[0].Struct.Fields[4].Function.ReturnType.Name)
	require.False(t, *&f.Expr[0].Struct.Fields[4].Function.ReturnType.IsPointer)
	require.False(t, *&f.Expr[0].Struct.Fields[4].Function.ReturnType.IsConst)
	require.Equal(t, "string_new_with_latin1_chars", *&f.Expr[0].Struct.Fields[4].Function.Name)
	require.Len(t, *&f.Expr[0].Struct.Fields[4].Function.Arguments, 2)
	require.Equal(t, "GDExtensionStringPtr", *&f.Expr[0].Struct.Fields[4].Function.Arguments[0].Type.Primative.Name)
	require.Equal(t, "r_dest", *&f.Expr[0].Struct.Fields[4].Function.Arguments[0].Name)
	require.Equal(t, "char", *&f.Expr[0].Struct.Fields[4].Function.Arguments[1].Type.Primative.Name)
	require.True(t, *&f.Expr[0].Struct.Fields[4].Function.Arguments[1].Type.Primative.IsConst)
	require.True(t, *&f.Expr[0].Struct.Fields[4].Function.Arguments[1].Type.Primative.IsPointer)
	require.Equal(t, "p_contents", *&f.Expr[0].Struct.Fields[4].Function.Arguments[1].Name)

	require.Equal(t, "int64_t", *&f.Expr[0].Struct.Fields[5].Function.ReturnType.Name)
	require.False(t, *&f.Expr[0].Struct.Fields[5].Function.ReturnType.IsConst)
	require.False(t, *&f.Expr[0].Struct.Fields[5].Function.ReturnType.IsPointer)
	require.Equal(t, "worker_thread_pool_add_native_group_task", *&f.Expr[0].Struct.Fields[5].Function.Name)
	require.Len(t, *&f.Expr[0].Struct.Fields[5].Function.Arguments, 7)
	require.Equal(t, "GDExtensionObjectPtr", *&f.Expr[0].Struct.Fields[5].Function.Arguments[0].Type.Primative.Name)
	require.Equal(t, "p_instance", *&f.Expr[0].Struct.Fields[5].Function.Arguments[0].Name)

	require.Equal(t, "p_func", *&f.Expr[0].Struct.Fields[5].Function.Arguments[1].Type.Function.Name)
	require.Equal(t, "void", *&f.Expr[0].Struct.Fields[5].Function.Arguments[1].Type.Function.ReturnType.Name)
	require.False(t, *&f.Expr[0].Struct.Fields[5].Function.Arguments[1].Type.Function.ReturnType.IsPointer)
	require.Len(t, *&f.Expr[0].Struct.Fields[5].Function.Arguments[1].Type.Function.Arguments, 2)
	require.Equal(t, "void", *&f.Expr[0].Struct.Fields[5].Function.Arguments[1].Type.Function.Arguments[0].Type.Primative.Name)
	require.True(t, *&f.Expr[0].Struct.Fields[5].Function.Arguments[1].Type.Function.Arguments[0].Type.Primative.IsPointer)
	require.Equal(t, "uint32_t", *&f.Expr[0].Struct.Fields[5].Function.Arguments[1].Type.Function.Arguments[1].Type.Primative.Name)
	require.False(t, *&f.Expr[0].Struct.Fields[5].Function.Arguments[1].Type.Function.Arguments[1].Type.Primative.IsPointer)

	require.Equal(t, "void", *&f.Expr[0].Struct.Fields[5].Function.Arguments[2].Type.Primative.Name)
	require.True(t, *&f.Expr[0].Struct.Fields[5].Function.Arguments[2].Type.Primative.IsPointer)
	require.Equal(t, "p_userdata", *&f.Expr[0].Struct.Fields[5].Function.Arguments[2].Name)

	require.Equal(t, "int", *&f.Expr[0].Struct.Fields[5].Function.Arguments[3].Type.Primative.Name)
	require.False(t, *&f.Expr[0].Struct.Fields[5].Function.Arguments[3].Type.Primative.IsPointer)
	require.Equal(t, "p_elements", *&f.Expr[0].Struct.Fields[5].Function.Arguments[3].Name)

	require.Equal(t, "int", *&f.Expr[0].Struct.Fields[5].Function.Arguments[4].Type.Primative.Name)
	require.False(t, *&f.Expr[0].Struct.Fields[5].Function.Arguments[4].Type.Primative.IsPointer)
	require.Equal(t, "p_tasks", *&f.Expr[0].Struct.Fields[5].Function.Arguments[4].Name)

	require.Equal(t, "GDExtensionBool", *&f.Expr[0].Struct.Fields[5].Function.Arguments[5].Type.Primative.Name)
	require.False(t, *&f.Expr[0].Struct.Fields[5].Function.Arguments[5].Type.Primative.IsPointer)
	require.Equal(t, "p_high_priority", *&f.Expr[0].Struct.Fields[5].Function.Arguments[5].Name)

	require.Equal(t, "GDExtensionConstStringPtr", *&f.Expr[0].Struct.Fields[5].Function.Arguments[6].Type.Primative.Name)
	require.False(t, *&f.Expr[0].Struct.Fields[5].Function.Arguments[6].Type.Primative.IsPointer)
	require.Equal(t, "p_description", *&f.Expr[0].Struct.Fields[5].Function.Arguments[6].Name)
}

func TestParseTypedefFunction(t *testing.T) {
	content := `typedef void (*GDExtensionPtrOperatorEvaluator)(const GDExtensionTypePtr p_left, const GDExtensionTypePtr p_right, GDExtensionTypePtr r_result);`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Function)

	require.Equal(t, "GDExtensionPtrOperatorEvaluator", f.Expr[0].Function.Name)

	require.Len(t, f.Expr[0].Function.Arguments, 3)
}

func TestParseTypedefFunction2(t *testing.T) {
	content := `typedef const GDExtensionPropertyInfo *(*GDExtensionClassGetPropertyList)(GDExtensionClassInstancePtr p_instance, uint32_t *r_count);`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Function)

	require.Equal(t, "GDExtensionClassGetPropertyList", f.Expr[0].Function.Name)

	require.Len(t, f.Expr[0].Function.Arguments, 2)
}

func TestParseTypedefFuncPointer(t *testing.T) {
	content := `
	typedef GDExtensionInterfaceFunctionPtr (*GDExtensionInterfaceGetProcAddress)(const char *p_function_name);
	`

	f, err := ParseCString(content)
	require.NoError(t, err)
	require.Equal(t, len(f.Expr), 1)
	require.NotNil(t, f.Expr[0].Function)
	require.Equal(t, f.Expr[0].Function.Name, "GDExtensionInterfaceGetProcAddress")
	require.Equal(t, f.Expr[0].Function.ReturnType.Name, "GDExtensionInterfaceFunctionPtr")
	require.False(t, f.Expr[0].Function.ReturnType.IsPointer)
	require.False(t, f.Expr[0].Function.ReturnType.IsConst)
	require.Len(t, f.Expr[0].Function.Arguments, 1)
	require.Equal(t, "p_function_name", f.Expr[0].Function.Arguments[0].Name)
	require.NotNil(t, f.Expr[0].Function.Arguments[0].Type.Primative)
	require.Equal(t, "char", f.Expr[0].Function.Arguments[0].Type.Primative.Name)
	require.True(t, f.Expr[0].Function.Arguments[0].Type.Primative.IsPointer)
	require.True(t, f.Expr[0].Function.Arguments[0].Type.Primative.IsConst)

	funcs := f.CollectFunctions()

	require.Len(t, funcs, 1)
}

func TestParseTypedefFunctionWithFunctionArgument(t *testing.T) {
	content := `typedef int64_t (*GDExtensionInterfaceWorkerThreadPoolAddNativeGroupTask)(GDExtensionObjectPtr p_instance, void (*p_func)(void *, uint32_t), void *p_userdata, int p_elements, int p_tasks, GDExtensionBool p_high_priority, GDExtensionConstStringPtr p_description);`

	f, err := ParseCString(content)

	require.NoError(t, err)
	require.Equal(t, len(f.Expr), 1)
	require.NotNil(t, f.Expr[0].Function)
	require.Equal(t, "GDExtensionInterfaceWorkerThreadPoolAddNativeGroupTask", f.Expr[0].Function.Name)
	require.Len(t, f.Expr[0].Function.Arguments, 7)
	require.Equal(t, "p_instance", f.Expr[0].Function.Arguments[0].Name)
	require.Equal(t, "GDExtensionObjectPtr", f.Expr[0].Function.Arguments[0].Type.Primative.Name)
	require.Equal(t, "p_func", f.Expr[0].Function.Arguments[1].Type.Function.Name)
	require.Len(t, f.Expr[0].Function.Arguments[1].Type.Function.Arguments, 2)
	require.True(t, f.Expr[0].Function.Arguments[1].Type.Function.Arguments[0].Type.Primative.IsPointer)
	require.Equal(t, "void", f.Expr[0].Function.Arguments[1].Type.Function.Arguments[0].Type.Primative.Name)
	require.False(t, f.Expr[0].Function.Arguments[1].Type.Function.Arguments[1].Type.Primative.IsPointer)
	require.Equal(t, "uint32_t", f.Expr[0].Function.Arguments[1].Type.Function.Arguments[1].Type.Primative.Name)
}

func TestParseTypedefFunctionNoArgumentNames(t *testing.T) {
	content := `typedef void (*GDExtensionVariantFromTypeConstructorFunc)(GDExtensionVariantPtr, GDExtensionTypePtr);`

	f, err := ParseCString(content)

	require.NoError(t, err)

	require.Equal(t, len(f.Expr), 1)

	require.NotNil(t, f.Expr[0].Function)

	require.Equal(t, "GDExtensionVariantFromTypeConstructorFunc", f.Expr[0].Function.Name)

	require.Len(t, f.Expr[0].Function.Arguments, 2)
}
