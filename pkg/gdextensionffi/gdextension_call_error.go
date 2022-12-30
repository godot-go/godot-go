package gdextensionffi

// #include <godot/gdextension_interface.h>
import "C"
import "fmt"

func (e GDExtensionCallError) Error() string {
	return fmt.Sprintf("GDExtensionCallError(error=%d, argument=%v", e.error, e.argument)
}

// Ok returns true if error equals GDEXTENSION_CALL_OK.
func (e GDExtensionCallError) Ok() bool {
	return (GDExtensionCallErrorType)(e.error) == GDEXTENSION_CALL_OK
}

// InvalidMethod returns true if error equals GDEXTENSION_CALL_ERROR_INVALID_METHOD.
func (e GDExtensionCallError) InvalidMethod() bool {
	return (GDExtensionCallErrorType)(e.error) == GDEXTENSION_CALL_ERROR_INVALID_METHOD
}

// InvalidArgument returns true if error equals GDEXTENSION_CALL_ERROR_INVALID_ARGUMENT.
func (e GDExtensionCallError) InvalidArgument() bool {
	return (GDExtensionCallErrorType)(e.error) == GDEXTENSION_CALL_ERROR_INVALID_ARGUMENT
}

// TooManyArguments returns true if error equals GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS.
func (e GDExtensionCallError) TooManyArguments() bool {
	return (GDExtensionCallErrorType)(e.error) == GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS
}

// TooFewArguments returns true if error equals GDEXTENSION_CALL_ERROR_TOO_FEW_ARGUMENTS.
func (e GDExtensionCallError) TooFewArguments() bool {
	return (GDExtensionCallErrorType)(e.error) == GDEXTENSION_CALL_ERROR_TOO_FEW_ARGUMENTS
}

// InstanceIsNull returns true if error equals GDEXTENSION_CALL_ERROR_INSTANCE_IS_NULL.
func (e GDExtensionCallError) InstanceIsNull() bool {
	return (GDExtensionCallErrorType)(e.error) == GDEXTENSION_CALL_ERROR_INSTANCE_IS_NULL
}

func (e *GDExtensionCallError) SetErrorFields(
	errorType GDExtensionCallErrorType,
	argument int32,
	expected int32,
) {
	((*C.GDExtensionCallError)(e)).error = (C.GDExtensionCallErrorType)(errorType)
	((*C.GDExtensionCallError)(e)).argument = (C.int)(argument)
	((*C.GDExtensionCallError)(e)).expected = (C.int)(expected)
}
