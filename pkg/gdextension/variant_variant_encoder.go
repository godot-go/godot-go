package gdextension

// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
)

func createVariantEncoder() argumentEncoder[Variant, Variant] {
	decodeTypePtrArg := func(ptr GDExtensionConstTypePtr, pOut *Variant) {
		if unsafe.Pointer(ptr) == unsafe.Pointer(pOut) {
			// noop
			return
		}
		CallFunc_GDExtensionInterfaceVariantNewCopy((GDExtensionUninitializedVariantPtr)(unsafe.Pointer(pOut)), (GDExtensionConstVariantPtr)(ptr))
	}
	decodeTypePtr := func(ptr GDExtensionConstTypePtr) Variant {
		var out Variant
		decodeTypePtrArg(ptr, &out)
		return out
	}
	encodeTypePtrArg := func(in Variant, pOut GDExtensionUninitializedTypePtr) {
		CallFunc_GDExtensionInterfaceVariantNewCopy((GDExtensionUninitializedVariantPtr)(unsafe.Pointer(pOut)), (GDExtensionConstVariantPtr)(unsafe.Pointer(&in)))
	}
	encodeTypePtr := func(in Variant) GDExtensionTypePtr {
		var out Variant
		pOut := (GDExtensionTypePtr)(unsafe.Pointer(&out))
		encodeTypePtrArg(in, (GDExtensionUninitializedTypePtr)(pOut))
		return pOut
	}
	decodeVariantPtrArg := func(ptr GDExtensionConstVariantPtr, pOut *Variant) {
		decodeTypePtrArg((GDExtensionConstTypePtr)(ptr), pOut)
	}
	decodeVariantPtr := func(ptr GDExtensionConstVariantPtr) Variant {
		var out Variant
		decodeVariantPtrArg(ptr, &out)
		return out
	}
	encodeVariantPtrArg := func(in Variant, pOut GDExtensionUninitializedVariantPtr) {
		encodeTypePtrArg(in, (GDExtensionUninitializedTypePtr)(pOut))
	}
	encodeVariantPtr := func(in Variant) GDExtensionVariantPtr {
		var out Variant
		pOut := (GDExtensionVariantPtr)(unsafe.Pointer(&out))
		encodeVariantPtrArg(
			in,
			(GDExtensionUninitializedVariantPtr)(pOut),
		)
		return pOut
	}
	decodeReflectTypePtr := func(ptr GDExtensionConstTypePtr) reflect.Value {
		v := decodeTypePtr(ptr)
		return reflect.ValueOf(v)
	}
	encodeReflectTypePtrArg := func(rv reflect.Value, pOut GDExtensionUninitializedTypePtr) {
		v := rv.Interface().(Variant)
		encodeTypePtrArg(v, pOut)
	}
	encodeReflectTypePtr := func(rv reflect.Value) GDExtensionTypePtr {
		var out Variant
		pOut := (GDExtensionTypePtr)(unsafe.Pointer(&out))
		encodeReflectTypePtrArg(rv, (GDExtensionUninitializedTypePtr)(pOut))
		return pOut
	}
	decodeReflectVariantPtr := func(ptr GDExtensionConstVariantPtr) reflect.Value {
		v := decodeVariantPtr(ptr)
		return reflect.ValueOf(v)
	}
	encodeReflectVariantPtrArg := func(rv reflect.Value, pOut GDExtensionUninitializedVariantPtr) {
		v := rv.Interface().(Variant)
		encodeVariantPtrArg(v, pOut)
	}
	encodeReflectVariantPtr := func(rv reflect.Value) GDExtensionVariantPtr {
		var out Variant
		pOut := (GDExtensionVariantPtr)(unsafe.Pointer(&out))
		encodeReflectVariantPtrArg(rv, (GDExtensionUninitializedVariantPtr)(pOut))
		return pOut
	}
	return argumentEncoder[Variant, Variant]{
		decodeTypePtrArg:           decodeTypePtrArg,
		decodeTypePtr:              decodeTypePtr,
		encodeTypePtrArg:           encodeTypePtrArg,
		encodeTypePtr:              encodeTypePtr,
		decodeVariantPtrArg:        decodeVariantPtrArg,
		decodeVariantPtr:           decodeVariantPtr,
		encodeVariantPtrArg:        encodeVariantPtrArg,
		encodeVariantPtr:           encodeVariantPtr,
		decodeReflectTypePtr:       decodeReflectTypePtr,
		encodeReflectTypePtrArg:    encodeReflectTypePtrArg,
		encodeReflectTypePtr:       encodeReflectTypePtr,
		decodeReflectVariantPtr:    decodeReflectVariantPtr,
		encodeReflectVariantPtrArg: encodeReflectVariantPtrArg,
		encodeReflectVariantPtr:    encodeReflectVariantPtr,
	}
}
