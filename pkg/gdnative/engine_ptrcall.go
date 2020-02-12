package gdnative

// #include <godot/gdnative_interface.h>
// #include "gdnative_wrapper.gen.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"
)

// callNativeMBRetObj port from godot-cpp:
// template <class O, class... Args>
//
//	O *_call_native_mb_ret_obj(const GDNativeMethodBindPtr mb, void *instance, const Args &...args) {
//		GodotObject *ret = nullptr;
//		std::array<const GDNativeTypePtr, sizeof...(Args)> mb_args = { { (const GDNativeTypePtr)args... } };
//		internal::gdn_interface->object_method_bind_ptrcall(mb, instance, mb_args.data(), &ret);
//		return reinterpret_cast<O *>(internal::gdn_interface->object_get_instance_binding(ret, internal::token, &O::___binding_callbacks));
//	}
func callNativeMBRetObj[T any](mb GDNativeMethodBindPtr, instance unsafe.Pointer, args ...GDNativeTypePtr) unsafe.Pointer {
	m := (C.GDNativeMethodBindPtr)(mb)
	a := (*C.GDNativeTypePtr)(unsafe.Pointer(&args[0]))

	var ret *GodotObject

	cInterface := (*C.GDNativeInterface)(internal.gdnInterface)

	cInst := (C.GDNativeObjectPtr)(instance)

	retTypePtr := (C.GDNativeTypePtr)(&ret)

	C.cgo_GDNativeInterface_object_method_bind_ptrcall(cInterface, m, cInst, a, retTypePtr)

	var cBindingCallbacks *C.GDNativeInstanceBindingCallbacks

	refNativeObjectPtr := (C.GDNativeObjectPtr)(ret)

	wrappedRet := C.cgo_GDNativeInterface_object_get_instance_binding(cInterface, refNativeObjectPtr, internal.token, cBindingCallbacks)

	return (unsafe.Pointer)(wrappedRet)
}

// callNativMBRet port from godot-cpp:
// template <class R, class... Args>
//
//	R _call_native_mb_ret(const GDNativeMethodBindPtr mb, void *instance, const Args &...args) {
//		R ret;
//		std::array<const GDNativeTypePtr, sizeof...(Args)> mb_args = { { (const GDNativeTypePtr)args... } };
//		internal::gdn_interface->object_method_bind_ptrcall(mb, instance, mb_args.data(), &ret);
//		return ret;
//	}
func callNativMBRet[T any](mb GDNativeMethodBindPtr, instance unsafe.Pointer, args ...GDNativeTypePtr) T {
	m := (C.GDNativeMethodBindPtr)(mb)
	a := (*C.GDNativeTypePtr)(unsafe.Pointer(&args[0]))

	var ret T

	cInterface := (*C.GDNativeInterface)(internal.gdnInterface)

	cInst := (C.GDNativeObjectPtr)(instance)

	retTypePtr := (C.GDNativeTypePtr)(&ret)

	C.cgo_GDNativeInterface_object_method_bind_ptrcall(cInterface, m, cInst, a, retTypePtr)

	return ret
}

// callNativMBNoRet port from godot-cpp:
// template <class... Args>
//
//	void _call_native_mb_no_ret(const GDNativeMethodBindPtr mb, void *instance, const Args &...args) {
//		std::array<const GDNativeTypePtr, sizeof...(Args)> mb_args = { { (const GDNativeTypePtr)args... } };
//		internal::gdn_interface->object_method_bind_ptrcall(mb, instance, mb_args.data(), nullptr);
//	}
func callNativMBNoRet[T any](mb GDNativeMethodBindPtr, instance unsafe.Pointer, args ...GDNativeTypePtr) {
	m := (C.GDNativeMethodBindPtr)(mb)
	a := (*C.GDNativeTypePtr)(unsafe.Pointer(&args[0]))

	cInterface := (*C.GDNativeInterface)(internal.gdnInterface)

	cInst := (C.GDNativeObjectPtr)(instance)

	retTypePtr := (C.GDNativeTypePtr)(nullptr)

	C.cgo_GDNativeInterface_object_method_bind_ptrcall(cInterface, m, cInst, a, retTypePtr)
}

// callUtilityRet port from godot-cpp:
// template <class R, class... Args>
//
//	R _call_utility_ret(GDNativePtrUtilityFunction func, const Args &...args) {
//		R ret;
//		std::array<const GDNativeTypePtr, sizeof...(Args)> mb_args = { { (const GDNativeTypePtr)args... } };
//		func(&ret, mb_args.data(), mb_args.size());
//		return ret;
//	}
func callUtilityRet[T any](callback GDNativePtrUtilityFunction, args ...GDNativeTypePtr) T {
	a := (*C.GDNativeTypePtr)(unsafe.Pointer(&args[0]))
	sz := (C.int)(len(args))

	var ret T

	retTypePtr := (C.GDNativeTypePtr)(&ret)

	cCallback := (C.GDNativePtrUtilityFunction)(callback)

	C.cgo_GDNativePtrUtilityFunction(cCallback, retTypePtr, a, sz)

	return ret
}

// callUtilityRetObj port from godot-cpp:
// template <class... Args>
//
//	Object *_call_utility_ret_obj(const GDNativePtrUtilityFunction func, void *instance, const Args &...args) {
//		GodotObject *ret = nullptr;
//		std::array<const GDNativeTypePtr, sizeof...(Args)> mb_args = { { (const GDNativeTypePtr)args... } };
//		func(&ret, mb_args.data(), mb_args.size());
//		return (Object *)internal::gdn_interface->object_get_instance_binding(ret, internal::token, &Object::___binding_callbacks);
//	}
func callUtilityRetObj(callback GDNativePtrUtilityFunction, args ...GDNativeTypePtr) *Object {
	a := (*C.GDNativeTypePtr)(unsafe.Pointer(&args[0]))
	sz := (C.int)(len(args))

	var ret *GodotObject

	retTypePtr := (C.GDNativeTypePtr)(&ret)

	cCallback := (C.GDNativePtrUtilityFunction)(callback)

	C.cgo_GDNativePtrUtilityFunction(cCallback, retTypePtr, a, sz)

	refNativeObjectPtr := (C.GDNativeObjectPtr)(ret)

	cInterface := (*C.GDNativeInterface)(internal.gdnInterface)

	tn := (TypeName)(reflect.TypeOf(ret).Name())

	bindingCallbacks, ok := gdExtensionBindingGDNativeInstanceBindingCallbacks.Get(tn)

	if !ok {
		panic(fmt.Sprintf("could not find instance binding callbacks for %s", (string)(tn)))
	}

	cBindingCallbacks := (*C.GDNativeInstanceBindingCallbacks)(&bindingCallbacks)

	wrappedRet := C.cgo_GDNativeInterface_object_get_instance_binding(cInterface, refNativeObjectPtr, internal.token, cBindingCallbacks)

	return (*Object)(wrappedRet)
}

// template <class... Args>
//
//	void _call_utility_no_ret(const GDNativePtrUtilityFunction func, const Args &...args) {
//		std::array<const GDNativeTypePtr, sizeof...(Args)> mb_args = { { (const GDNativeTypePtr)args... } };
//		func(nullptr, mb_args.data(), mb_args.size());
//	}
func callUtilityNoRet(callback GDNativePtrUtilityFunction, args ...GDNativeTypePtr) {
	a := (*C.GDNativeTypePtr)(unsafe.Pointer(&args[0]))
	sz := (C.int)(len(args))

	retTypePtr := (C.GDNativeTypePtr)(nullptr)

	cCallback := (C.GDNativePtrUtilityFunction)(callback)

	C.cgo_GDNativePtrUtilityFunction(cCallback, retTypePtr, a, sz)
}
