package builtin

import (
	"runtime"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
)

// Each copy constructor pins the caller's source only for the duration of the
// synchronous engine call that reads it; the engine copies the container rather
// than retaining the pointer, so the pin is released when the call returns.

func ArrayCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalArrayMethodBindings.constructor_1, out, src)
}

func PackedByteArrayCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalPackedByteArrayMethodBindings.constructor_1, out, src)
}

func PackedInt32ArrayCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalPackedInt32ArrayMethodBindings.constructor_1, out, src)
}

func PackedInt64ArrayCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalPackedInt64ArrayMethodBindings.constructor_1, out, src)
}

func PackedFloat32ArrayCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalPackedFloat32ArrayMethodBindings.constructor_1, out, src)
}

func PackedFloat64ArrayCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalPackedFloat64ArrayMethodBindings.constructor_1, out, src)
}

func PackedStringArrayCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalPackedStringArrayMethodBindings.constructor_1, out, src)
}

func PackedVector2ArrayCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalPackedVector2ArrayMethodBindings.constructor_1, out, src)
}

func PackedVector3ArrayCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalPackedVector3ArrayMethodBindings.constructor_1, out, src)
}

func PackedColorArrayCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalPackedColorArrayMethodBindings.constructor_1, out, src)
}
