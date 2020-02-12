package gdnative

// #include <godot/gdnative_interface.h>
// #include "gdnative_wrapper.gen.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import "unsafe"

const (
	MAX_ARG_COUNT = 255
)

func callBuiltinConstructor(constructor GDNativePtrConstructor, base GDNativeTypePtr, args ...GDNativeTypePtr) {
	c := (C.GDNativePtrConstructor)(constructor)
	b := (C.GDNativeTypePtr)(base)

	var a *C.GDNativeTypePtr
	if len(args) > 0 {
		a = (*C.GDNativeTypePtr)(unsafe.Pointer(&args[0]))
	}

	C.cgo_GDNativePtrConstructor(c, b, a)
}

func callBuiltinMethodPtrRet[T any](method GDNativePtrBuiltInMethod, base GDNativeTypePtr, args *[MAX_ARG_COUNT]GDNativeTypePtr) T {
	m := (C.GDNativePtrBuiltInMethod)(method)
	b := (C.GDNativeTypePtr)(base)
	var a *C.GDNativeTypePtr
	if len(args) > 0 {
		a = (*C.GDNativeTypePtr)(unsafe.Pointer(&args[0]))
	}
	ca := (C.int)(len(args))

	var ret T

	ptr := (C.GDNativeTypePtr)(unsafe.Pointer(&ret))

	C.cgo_GDNativePtrBuiltInMethod(m, b, a, ptr, ca)

	return ret
}

func callBuiltinMethodPtrNoRet(method GDNativePtrBuiltInMethod, base GDNativeTypePtr, args *[MAX_ARG_COUNT]GDNativeTypePtr) {
	m := (C.GDNativePtrBuiltInMethod)(method)
	b := (C.GDNativeTypePtr)(base)
	var a *C.GDNativeTypePtr
	if len(args) > 0 {
		a = (*C.GDNativeTypePtr)(unsafe.Pointer(&args[0]))
	}
	ca := (C.int)(len(args))

	C.cgo_GDNativePtrBuiltInMethod(m, b, a, nil, ca)
}

func callBuiltinOperatorPtr[T any](operator GDNativePtrOperatorEvaluator, left GDNativeTypePtr, right GDNativeTypePtr) T {
	op := (C.GDNativePtrOperatorEvaluator)(operator)
	l := (C.GDNativeTypePtr)(left)
	r := (C.GDNativeTypePtr)(right)

	var ret T

	ptr := (C.GDNativeTypePtr)(unsafe.Pointer(&ret))

	C.cgo_GDNativePtrOperatorEvaluator(op, l, r, ptr)

	return ret
}

func callBuiltinPtrGetter[T any](getter GDNativePtrGetter, base GDNativeTypePtr) T {
	g := (C.GDNativePtrGetter)(getter)
	b := (C.GDNativeTypePtr)(base)

	var ret T

	ptr := (C.GDNativeTypePtr)(unsafe.Pointer(&ret))

	C.cgo_GDNativePtrGetter(g, b, ptr)

	return ret
}

func callBuiltinPtrSetter[T any](setter GDNativePtrSetter, base GDNativeTypePtr) T {
	g := (C.GDNativePtrSetter)(setter)
	b := (C.GDNativeTypePtr)(base)

	var ret T

	ptr := (C.GDNativeTypePtr)(unsafe.Pointer(&ret))

	C.cgo_GDNativePtrSetter(g, b, ptr)

	return ret
}
