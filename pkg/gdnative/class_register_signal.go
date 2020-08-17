package gdnative

/*
#include <nativescript.wrappergen.h>
#include <cgo_gateway_register_class.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"
)

type RegisterSignalArg struct {
	Name string
	Type VariantType
}

func (d ClassRegisteredEvent) RegisterSignal(signalName string, varargs ...RegisterSignalArg) {
	cClassName := C.CString(d.ClassName)
	defer C.free(unsafe.Pointer(cClassName))

	gsSignalName := NewStringFromGoString(signalName)
	defer gsSignalName.Destroy()

	size := len(varargs)

	signal := C.godot_signal{}
	signal.name = *(*C.godot_string)(unsafe.Pointer(&gsSignalName))
	signal.num_args = (C.int)(size)
	signal.num_default_args = 0

	if size > 0 {
		signal.args = (*C.godot_signal_argument)(unsafe.Pointer(AllocZeros(int32(unsafe.Sizeof(SignalArgument{})) * int32(size))))
		defer Free(unsafe.Pointer(signal.args))
	}

	argPtr := (*C.godot_signal_argument)(unsafe.Pointer(signal.args))

	for i, a := range varargs {
		str := NewStringFromGoString(a.Name)
		defer str.Destroy()

		curArgPtr := (*C.godot_signal_argument)(unsafe.Pointer(uintptr(unsafe.Pointer(argPtr)) + uintptr(i)*uintptr(C.sizeof_godot_signal_argument)))
		curArgPtr.name = *(*C.godot_string)(unsafe.Pointer(&str))
		curArgPtr._type = (C.godot_int)(a.Type)
	}

	C.go_godot_nativescript_register_signal(NativescriptApi, RegisterState.NativescriptHandle, cClassName, &signal)
}
