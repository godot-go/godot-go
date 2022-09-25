package gdnative

// #include <godot/gdnative_interface.h>
import "C"
import "fmt"

func (e GDNativeCallError) Error() string {
	return fmt.Sprintf("GDNativeCallError(error=%d, argument=%v", e.error, e.argument)
}

// Ok returns true if error equals GDNATIVE_CALL_OK.
func (e GDNativeCallError) Ok() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_OK
}

// InvalidMethod returns true if error equals GDNATIVE_CALL_ERROR_INVALID_METHOD.
func (e GDNativeCallError) InvalidMethod() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_ERROR_INVALID_METHOD
}

// InvalidArgument returns true if error equals GDNATIVE_CALL_ERROR_INVALID_ARGUMENT.
func (e GDNativeCallError) InvalidArgument() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_ERROR_INVALID_ARGUMENT
}

// TooManyArguments returns true if error equals GDNATIVE_CALL_ERROR_TOO_MANY_ARGUMENTS.
func (e GDNativeCallError) TooManyArguments() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_ERROR_TOO_MANY_ARGUMENTS
}

// TooFewArguments returns true if error equals GDNATIVE_CALL_ERROR_TOO_FEW_ARGUMENTS.
func (e GDNativeCallError) TooFewArguments() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_ERROR_TOO_FEW_ARGUMENTS
}

// InstanceIsNull returns true if error equals GDNATIVE_CALL_ERROR_INSTANCE_IS_NULL.
func (e GDNativeCallError) InstanceIsNull() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_ERROR_INSTANCE_IS_NULL
}

func (e *GDNativeCallError) SetErrorFields(
	errorType GDNativeCallErrorType,
	argument int32,
	expected int32,
) {
	((*C.GDNativeCallError)(e)).error = (C.GDNativeCallErrorType)(errorType)
	((*C.GDNativeCallError)(e)).argument = (C.int)(argument)
	((*C.GDNativeCallError)(e)).expected = (C.int)(expected)
}
