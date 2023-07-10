package gdextension

import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"

	"github.com/godot-go/godot-go/pkg/log"
)

func callBuiltinConstructor(constructor GDExtensionPtrConstructor, base GDExtensionUninitializedTypePtr, args ...GDExtensionConstTypePtr) {
	c := (GDExtensionPtrConstructor)(constructor)
	b := (GDExtensionUninitializedTypePtr)(base)

	if c == nil {
		log.Panic("constructor is null")
	}

	var a *GDExtensionConstTypePtr
	if len(args) > 0 {
		a = (*GDExtensionConstTypePtr)(unsafe.Pointer(&args[0]))
	} else {
		a = (*GDExtensionConstTypePtr)(nullptr)
	}

	CallFunc_GDExtensionPtrConstructor(c, b, a)
}

func callBuiltinMethodPtrRet[T any](method GDExtensionPtrBuiltInMethod, base GDExtensionTypePtr, args ...GDExtensionTypePtr) T {
	m := (GDExtensionPtrBuiltInMethod)(method)
	b := (GDExtensionTypePtr)(base)
	var a *GDExtensionConstTypePtr
	if len(args) > 0 {
		a = (*GDExtensionConstTypePtr)(unsafe.Pointer(&args[0]))
	} else {
		a = (*GDExtensionConstTypePtr)(nullptr)
	}
	ca := (int32)(len(args))

	var ret T

	ptr := (GDExtensionTypePtr)(unsafe.Pointer(&ret))

	CallFunc_GDExtensionPtrBuiltInMethod(m, b, a, ptr, ca)

	return ret
}

func callBuiltinMethodPtrNoRet(method GDExtensionPtrBuiltInMethod, base GDExtensionTypePtr, args ...GDExtensionTypePtr) {
	m := (GDExtensionPtrBuiltInMethod)(method)
	b := (GDExtensionTypePtr)(base)
	var a *GDExtensionConstTypePtr
	if len(args) > 0 {
		a = (*GDExtensionConstTypePtr)(unsafe.Pointer(&args[0]))
	} else {
		a = (*GDExtensionConstTypePtr)(nullptr)
	}
	ca := (int32)(len(args))

	CallFunc_GDExtensionPtrBuiltInMethod(m, b, a, nil, ca)
}

func callBuiltinOperatorPtr[T any](operator GDExtensionPtrOperatorEvaluator, left GDExtensionConstTypePtr, right GDExtensionConstTypePtr) T {
	op := (GDExtensionPtrOperatorEvaluator)(operator)
	l := (GDExtensionConstTypePtr)(left)
	r := (GDExtensionConstTypePtr)(right)

	var ret T

	ptr := (GDExtensionTypePtr)(unsafe.Pointer(&ret))

	CallFunc_GDExtensionPtrOperatorEvaluator(op, l, r, ptr)

	return ret
}

func callBuiltinPtrGetter[T any](getter GDExtensionPtrGetter, base GDExtensionConstTypePtr) T {
	g := (GDExtensionPtrGetter)(getter)
	b := (GDExtensionConstTypePtr)(base)

	var ret T

	ptr := (GDExtensionTypePtr)(unsafe.Pointer(&ret))

	CallFunc_GDExtensionPtrGetter(g, b, ptr)

	return ret
}

func callBuiltinPtrSetter[T any](setter GDExtensionPtrSetter, base GDExtensionTypePtr) T {
	g := (GDExtensionPtrSetter)(setter)
	b := (GDExtensionTypePtr)(base)

	var ret T

	ptr := (GDExtensionConstTypePtr)(unsafe.Pointer(&ret))

	CallFunc_GDExtensionPtrSetter(g, b, ptr)

	return ret
}
