package gdnative

// #include <godot/gdnative_interface.h>
// #include "gdnative_wrapper.gen.h"
// #include "gdnative_binding_wrapper.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import "fmt"

func (e GDNativeCallError) Error() string {
	return fmt.Sprintf("GDNativeCallError(error=%d, argument=%v", e.error, e.argument)
}

func (e GDNativeCallError) Ok() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_OK
}

func (e GDNativeCallError) InvalidMethod() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_ERROR_INVALID_METHOD
}

func (e GDNativeCallError) InvalidArgument() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_ERROR_INVALID_ARGUMENT
}

func (e GDNativeCallError) TooManyArguments() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_ERROR_TOO_MANY_ARGUMENTS
}

func (e GDNativeCallError) TooFewArguments() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_ERROR_TOO_FEW_ARGUMENTS
}

func (e GDNativeCallError) InstanceIsNull() bool {
	return (GDNativeCallErrorType)(e.error) == GDNATIVE_CALL_ERROR_INSTANCE_IS_NULL
}
