package gdextension

import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdnative"
)

func callBuiltinConstructor(constructor GDNativePtrConstructor, base GDNativeTypePtr, args ...GDNativeTypePtr) {
	c := (GDNativePtrConstructor)(constructor)
	b := (GDNativeTypePtr)(base)

	var a *GDNativeTypePtr
	if len(args) > 0 {
		a = (*GDNativeTypePtr)(unsafe.Pointer(&args[0]))
	}

	CallFunc_GDNativePtrConstructor(c, b, a)
}

func callBuiltinMethodPtrRet[T any](method GDNativePtrBuiltInMethod, base GDNativeTypePtr, args *[MAX_ARG_COUNT]GDNativeTypePtr) T {
	m := (GDNativePtrBuiltInMethod)(method)
	b := (GDNativeTypePtr)(base)
	var a *GDNativeTypePtr
	if len(args) > 0 {
		a = (*GDNativeTypePtr)(unsafe.Pointer(&args[0]))
	}
	ca := (int32)(len(args))

	var ret T

	ptr := (GDNativeTypePtr)(unsafe.Pointer(&ret))

	CallFunc_GDNativePtrBuiltInMethod(m, b, a, ptr, ca)

	return ret
}

func callBuiltinMethodPtrNoRet(method GDNativePtrBuiltInMethod, base GDNativeTypePtr, args *[MAX_ARG_COUNT]GDNativeTypePtr) {
	m := (GDNativePtrBuiltInMethod)(method)
	b := (GDNativeTypePtr)(base)
	var a *GDNativeTypePtr
	if len(args) > 0 {
		a = (*GDNativeTypePtr)(unsafe.Pointer(&args[0]))
	}
	ca := (int32)(len(args))

	CallFunc_GDNativePtrBuiltInMethod(m, b, a, nil, ca)
}

func callBuiltinOperatorPtr[T any](operator GDNativePtrOperatorEvaluator, left GDNativeTypePtr, right GDNativeTypePtr) T {
	op := (GDNativePtrOperatorEvaluator)(operator)
	l := (GDNativeTypePtr)(left)
	r := (GDNativeTypePtr)(right)

	var ret T

	ptr := (GDNativeTypePtr)(unsafe.Pointer(&ret))

	CallFunc_GDNativePtrOperatorEvaluator(op, l, r, ptr)

	return ret
}

func callBuiltinPtrGetter[T any](getter GDNativePtrGetter, base GDNativeTypePtr) T {
	g := (GDNativePtrGetter)(getter)
	b := (GDNativeTypePtr)(base)

	var ret T

	ptr := (GDNativeTypePtr)(unsafe.Pointer(&ret))

	CallFunc_GDNativePtrGetter(g, b, ptr)

	return ret
}

func callBuiltinPtrSetter[T any](setter GDNativePtrSetter, base GDNativeTypePtr) T {
	g := (GDNativePtrSetter)(setter)
	b := (GDNativeTypePtr)(base)

	var ret T

	ptr := (GDNativeTypePtr)(unsafe.Pointer(&ret))

	CallFunc_GDNativePtrSetter(g, b, ptr)

	return ret
}
