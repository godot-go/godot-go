package ffi

import (
	"runtime"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
)

// Each helper pins only the scratch it hands to a single synchronous engine call.
// The engine reads the argument pointer array and the return slot during the call
// and retains neither, so the pins are released when the helper returns.

func CallBuiltinConstructor(constructor GDExtensionPtrConstructor, base GDExtensionUninitializedTypePtr, args ...GDExtensionConstTypePtr) {
	if constructor == nil {
		log.Panic("constructor is null")
	}
	var pinner runtime.Pinner
	defer pinner.Unpin()
	argsPtr := (*GDExtensionConstTypePtr)(unsafe.SliceData(args))
	pinner.Pin(argsPtr)
	CallFunc_GDExtensionPtrConstructor(constructor, base, argsPtr)
}

func CallBuiltinMethodPtrRet[T any](method GDExtensionPtrBuiltInMethod, base GDExtensionTypePtr, args ...GDExtensionTypePtr) T {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	m := (GDExtensionPtrBuiltInMethod)(method)
	b := (GDExtensionTypePtr)(base)
	a := (*GDExtensionConstTypePtr)(unsafe.SliceData(args))
	ca := (int32)(len(args))
	var ret T
	ptr := (GDExtensionTypePtr)(unsafe.Pointer(&ret))
	pinner.Pin(a)
	pinner.Pin(ptr)
	CallFunc_GDExtensionPtrBuiltInMethod(m, b, a, ptr, ca)
	return ret
}

func CallBuiltinMethodPtrNoRet(method GDExtensionPtrBuiltInMethod, base GDExtensionTypePtr, args ...GDExtensionTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	m := (GDExtensionPtrBuiltInMethod)(method)
	b := (GDExtensionTypePtr)(base)
	a := (*GDExtensionConstTypePtr)(unsafe.SliceData(args))
	ca := (int32)(len(args))
	pinner.Pin(a)
	CallFunc_GDExtensionPtrBuiltInMethod(m, b, a, nil, ca)
}

func CallBuiltinOperatorPtr[T any](operator GDExtensionPtrOperatorEvaluator, left GDExtensionConstTypePtr, right GDExtensionConstTypePtr) T {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	op := (GDExtensionPtrOperatorEvaluator)(operator)
	l := (GDExtensionConstTypePtr)(left)
	r := (GDExtensionConstTypePtr)(right)
	var ret T
	ptr := (GDExtensionTypePtr)(unsafe.Pointer(&ret))
	pinner.Pin(l)
	pinner.Pin(r)
	pinner.Pin(ptr)
	CallFunc_GDExtensionPtrOperatorEvaluator(op, l, r, ptr)
	return ret
}

func CallBuiltinPtrGetter[T any](getter GDExtensionPtrGetter, base GDExtensionConstTypePtr) T {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	g := (GDExtensionPtrGetter)(getter)
	b := (GDExtensionConstTypePtr)(base)
	var ret T
	ptr := (GDExtensionTypePtr)(unsafe.Pointer(&ret))
	pinner.Pin(ptr)
	CallFunc_GDExtensionPtrGetter(g, b, ptr)
	return ret
}

func CallBuiltinPtrSetter[T any](setter GDExtensionPtrSetter, base GDExtensionTypePtr) T {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	g := (GDExtensionPtrSetter)(setter)
	b := (GDExtensionTypePtr)(base)
	var ret T
	ptr := (GDExtensionConstTypePtr)(unsafe.Pointer(&ret))
	pinner.Pin(ptr)
	CallFunc_GDExtensionPtrSetter(g, b, ptr)
	return ret
}
