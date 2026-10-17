package core

import (
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// Go-native slice marshalling against engine Packed*K containers
// (support-native-slice-marshalling). Design D1: every crossing is a deep
// copy — decoded slices are Go-owned memory released by the GC with no engine
// release call, and the engine never holds a pointer to Go memory. The single
// element→packed-type table lives in packedVariantTypeForSliceElement
// (variant_reflect_type.go); these helpers only ever see elements it accepts.

// sliceBulkCopy copies n elements starting at base into a fresh Go slice.
// A non-positive length or a null base (the engine may hand back null for an
// empty buffer) yields a nil slice, so the returned pointer is never touched
// for zero-length arrays.
func sliceBulkCopy[T any](base *T, n int64) []T {
	if n <= 0 || base == nil {
		return nil
	}
	src := unsafe.Slice(base, n)
	out := make([]T, n)
	copy(out, src)
	return out
}

// sliceFromPackedCell copies the elements of a Packed*K cell into a fresh Go
// slice whose element type is elem, reading through the engine const
// operator_index accessors. The cell is borrowed: it is fully consumed before
// returning, so an owned temporary may be released as soon as this returns.
// elem must be a key accepted by packedVariantTypeForSliceElement.
func sliceFromPackedCell(cell GDExtensionConstTypePtr, elem reflect.Type) reflect.Value {
	switch reflect.Zero(elem).Interface().(type) {
	case byte:
		base := CallFunc_GDExtensionInterfacePackedByteArrayOperatorIndexConst(cell, 0)
		return reflect.ValueOf(sliceBulkCopy(base, (*PackedByteArray)(unsafe.Pointer(cell)).Size()))
	case int32:
		base := CallFunc_GDExtensionInterfacePackedInt32ArrayOperatorIndexConst(cell, 0)
		return reflect.ValueOf(sliceBulkCopy(base, (*PackedInt32Array)(unsafe.Pointer(cell)).Size()))
	case int:
		// Go int follows the repo's int→int64 scalar rule; the engine element
		// is int64 on every supported platform.
		base := CallFunc_GDExtensionInterfacePackedInt64ArrayOperatorIndexConst(cell, 0)
		return reflect.ValueOf(sliceBulkCopy((*int)(unsafe.Pointer(base)), (*PackedInt64Array)(unsafe.Pointer(cell)).Size()))
	case int64:
		base := CallFunc_GDExtensionInterfacePackedInt64ArrayOperatorIndexConst(cell, 0)
		return reflect.ValueOf(sliceBulkCopy(base, (*PackedInt64Array)(unsafe.Pointer(cell)).Size()))
	case float32:
		base := CallFunc_GDExtensionInterfacePackedFloat32ArrayOperatorIndexConst(cell, 0)
		return reflect.ValueOf(sliceBulkCopy(base, (*PackedFloat32Array)(unsafe.Pointer(cell)).Size()))
	case float64:
		base := CallFunc_GDExtensionInterfacePackedFloat64ArrayOperatorIndexConst(cell, 0)
		return reflect.ValueOf(sliceBulkCopy(base, (*PackedFloat64Array)(unsafe.Pointer(cell)).Size()))
	case string:
		return reflect.ValueOf(stringsFromPackedCell(cell))
	case Vector2:
		base := CallFunc_GDExtensionInterfacePackedVector2ArrayOperatorIndexConst(cell, 0)
		return reflect.ValueOf(sliceBulkCopy((*Vector2)(unsafe.Pointer(base)), (*PackedVector2Array)(unsafe.Pointer(cell)).Size()))
	case Vector3:
		base := CallFunc_GDExtensionInterfacePackedVector3ArrayOperatorIndexConst(cell, 0)
		return reflect.ValueOf(sliceBulkCopy((*Vector3)(unsafe.Pointer(base)), (*PackedVector3Array)(unsafe.Pointer(cell)).Size()))
	case Vector4:
		base := CallFunc_GDExtensionInterfacePackedVector4ArrayOperatorIndexConst(cell, 0)
		return reflect.ValueOf(sliceBulkCopy((*Vector4)(unsafe.Pointer(base)), (*PackedVector4Array)(unsafe.Pointer(cell)).Size()))
	case Color:
		base := CallFunc_GDExtensionInterfacePackedColorArrayOperatorIndexConst(cell, 0)
		return reflect.ValueOf(sliceBulkCopy((*Color)(unsafe.Pointer(base)), (*PackedColorArray)(unsafe.Pointer(cell)).Size()))
	}
	// Unreachable for bound signatures: registration rejects elements outside
	// the D2 table before any call reaches the decoders.
	panic("slice element type has no packed array decoder: " + elem.String())
}

// stringsFromPackedCell decodes a PackedStringArray cell element-by-element:
// const operator_index borrows the native String at each slot (design D3),
// ToUtf8 produces the retained Go copy, and nothing needs releasing.
func stringsFromPackedCell(cell GDExtensionConstTypePtr) []string {
	n := (*PackedStringArray)(unsafe.Pointer(cell)).Size()
	if n <= 0 {
		return nil
	}
	out := make([]string, n)
	for i := int64(0); i < n; i++ {
		slot := CallFunc_GDExtensionInterfacePackedStringArrayOperatorIndexConst(cell, GDExtensionInt(i))
		if slot == nil {
			continue
		}
		borrowed := *(*String)(unsafe.Pointer(slot))
		out[i] = borrowed.ToUtf8()
	}
	return out
}

// sliceDecodeVarcallArg converts a varcall argument to a fresh Go slice. The
// engine's strict-conversion gate has already accepted the variant against the
// declared PACKED_* registration type; the ToPacked*K conversion hands back an
// owned temporary that is released here before returning, including on every
// early path (the varcall-argument-decode release discipline).
func sliceDecodeVarcallArg(arg Variant, elem reflect.Type) reflect.Value {
	switch reflect.Zero(elem).Interface().(type) {
	case byte:
		tmp := arg.ToPackedByteArray()
		defer tmp.Destroy()
		return sliceFromPackedCell(tmp.NativeConstPtr(), elem)
	case int32:
		tmp := arg.ToPackedInt32Array()
		defer tmp.Destroy()
		return sliceFromPackedCell(tmp.NativeConstPtr(), elem)
	case int, int64:
		tmp := arg.ToPackedInt64Array()
		defer tmp.Destroy()
		return sliceFromPackedCell(tmp.NativeConstPtr(), elem)
	case float32:
		tmp := arg.ToPackedFloat32Array()
		defer tmp.Destroy()
		return sliceFromPackedCell(tmp.NativeConstPtr(), elem)
	case float64:
		tmp := arg.ToPackedFloat64Array()
		defer tmp.Destroy()
		return sliceFromPackedCell(tmp.NativeConstPtr(), elem)
	case string:
		tmp := arg.ToPackedStringArray()
		defer tmp.Destroy()
		return sliceFromPackedCell(tmp.NativeConstPtr(), elem)
	case Vector2:
		tmp := arg.ToPackedVector2Array()
		defer tmp.Destroy()
		return sliceFromPackedCell(tmp.NativeConstPtr(), elem)
	case Vector3:
		tmp := arg.ToPackedVector3Array()
		defer tmp.Destroy()
		return sliceFromPackedCell(tmp.NativeConstPtr(), elem)
	case Vector4:
		tmp := arg.ToPackedVector4Array()
		defer tmp.Destroy()
		return sliceFromPackedCell(tmp.NativeConstPtr(), elem)
	case Color:
		tmp := arg.ToPackedColorArray()
		defer tmp.Destroy()
		return sliceFromPackedCell(tmp.NativeConstPtr(), elem)
	}
	panic("slice element type has no packed array decoder: " + elem.String())
}

// sliceBulkWrite copies src into engine memory starting at base. The cell must
// already be resized to len(src): the engine may hand back null for slots
// outside the buffer, so the empty case never dereferences base.
func sliceBulkWrite[T any](base *T, src []T) {
	if len(src) == 0 {
		return
	}
	if base == nil {
		panic("packed array operator index returned null for a resized buffer")
	}
	copy(unsafe.Slice(base, int64(len(src))), src)
}

// sliceEncodeReturn copies a Go slice into a Packed*K cell and installs the
// value at rOut. The nine cells with generated copy constructors follow the
// ratified idiom (copy-ctor into rOut, then release the temporary);
// PackedVector4Array follows the predating move idiom of the cell return
// path (raw cell transfer into rOut, no destroy). A nil or empty slice
// encodes as an empty container, never null (design D6).
func sliceEncodeReturn(rOut GDExtensionUninitializedTypePtr, v reflect.Value) {
	switch reflect.Zero(v.Type().Elem()).Interface().(type) {
	case byte:
		src := v.Interface().([]byte)
		tmp := NewPackedByteArray()
		sliceResizeCell(tmp.Resize(int64(len(src))), "PackedByteArray", len(src))
		base := CallFunc_GDExtensionInterfacePackedByteArrayOperatorIndex(tmp.NativePtr(), 0)
		sliceBulkWrite(base, src)
		PackedByteArrayCopyConstructor(rOut, tmp.NativeConstPtr())
		tmp.Destroy()
	case int32:
		src := v.Interface().([]int32)
		tmp := NewPackedInt32Array()
		sliceResizeCell(tmp.Resize(int64(len(src))), "PackedInt32Array", len(src))
		base := CallFunc_GDExtensionInterfacePackedInt32ArrayOperatorIndex(tmp.NativePtr(), 0)
		sliceBulkWrite(base, src)
		PackedInt32ArrayCopyConstructor(rOut, tmp.NativeConstPtr())
		tmp.Destroy()
	case int:
		src := v.Interface().([]int)
		tmp := NewPackedInt64Array()
		sliceResizeCell(tmp.Resize(int64(len(src))), "PackedInt64Array", len(src))
		base := CallFunc_GDExtensionInterfacePackedInt64ArrayOperatorIndex(tmp.NativePtr(), 0)
		sliceBulkWrite((*int)(unsafe.Pointer(base)), src)
		PackedInt64ArrayCopyConstructor(rOut, tmp.NativeConstPtr())
		tmp.Destroy()
	case int64:
		src := v.Interface().([]int64)
		tmp := NewPackedInt64Array()
		sliceResizeCell(tmp.Resize(int64(len(src))), "PackedInt64Array", len(src))
		base := CallFunc_GDExtensionInterfacePackedInt64ArrayOperatorIndex(tmp.NativePtr(), 0)
		sliceBulkWrite(base, src)
		PackedInt64ArrayCopyConstructor(rOut, tmp.NativeConstPtr())
		tmp.Destroy()
	case float32:
		src := v.Interface().([]float32)
		tmp := NewPackedFloat32Array()
		sliceResizeCell(tmp.Resize(int64(len(src))), "PackedFloat32Array", len(src))
		base := CallFunc_GDExtensionInterfacePackedFloat32ArrayOperatorIndex(tmp.NativePtr(), 0)
		sliceBulkWrite(base, src)
		PackedFloat32ArrayCopyConstructor(rOut, tmp.NativeConstPtr())
		tmp.Destroy()
	case float64:
		src := v.Interface().([]float64)
		tmp := NewPackedFloat64Array()
		sliceResizeCell(tmp.Resize(int64(len(src))), "PackedFloat64Array", len(src))
		base := CallFunc_GDExtensionInterfacePackedFloat64ArrayOperatorIndex(tmp.NativePtr(), 0)
		sliceBulkWrite(base, src)
		PackedFloat64ArrayCopyConstructor(rOut, tmp.NativeConstPtr())
		tmp.Destroy()
	case string:
		src := v.Interface().([]string)
		tmp := NewPackedStringArray()
		// Append(String) is the only engine-supported way to place a String
		// element: each element is constructed as a native String, appended
		// (the array takes its own reference), and released immediately.
		for _, el := range src {
			s := NewStringWithUtf8Chars(el)
			tmp.Append(s)
			s.Destroy()
		}
		PackedStringArrayCopyConstructor(rOut, tmp.NativeConstPtr())
		tmp.Destroy()
	case Vector2:
		src := v.Interface().([]Vector2)
		tmp := NewPackedVector2Array()
		sliceResizeCell(tmp.Resize(int64(len(src))), "PackedVector2Array", len(src))
		base := CallFunc_GDExtensionInterfacePackedVector2ArrayOperatorIndex(tmp.NativePtr(), 0)
		sliceBulkWrite((*Vector2)(unsafe.Pointer(base)), src)
		PackedVector2ArrayCopyConstructor(rOut, tmp.NativeConstPtr())
		tmp.Destroy()
	case Vector3:
		src := v.Interface().([]Vector3)
		tmp := NewPackedVector3Array()
		sliceResizeCell(tmp.Resize(int64(len(src))), "PackedVector3Array", len(src))
		base := CallFunc_GDExtensionInterfacePackedVector3ArrayOperatorIndex(tmp.NativePtr(), 0)
		sliceBulkWrite((*Vector3)(unsafe.Pointer(base)), src)
		PackedVector3ArrayCopyConstructor(rOut, tmp.NativeConstPtr())
		tmp.Destroy()
	case Vector4:
		src := v.Interface().([]Vector4)
		tmp := NewPackedVector4Array()
		sliceResizeCell(tmp.Resize(int64(len(src))), "PackedVector4Array", len(src))
		base := CallFunc_GDExtensionInterfacePackedVector4ArrayOperatorIndex(tmp.NativePtr(), 0)
		sliceBulkWrite((*Vector4)(unsafe.Pointer(base)), src)
		// Move idiom: the encoder writes the cell bytes into rOut without a
		// copy constructor, so ownership transfers and the temporary must
		// not be destroyed (mirrors the PackedVector4Array cell return).
		PackedVector4ArrayEncoder.EncodeTypePtrArg(tmp, rOut)
	case Color:
		src := v.Interface().([]Color)
		tmp := NewPackedColorArray()
		sliceResizeCell(tmp.Resize(int64(len(src))), "PackedColorArray", len(src))
		base := CallFunc_GDExtensionInterfacePackedColorArrayOperatorIndex(tmp.NativePtr(), 0)
		sliceBulkWrite((*Color)(unsafe.Pointer(base)), src)
		PackedColorArrayCopyConstructor(rOut, tmp.NativeConstPtr())
		tmp.Destroy()
	default:
		panic("slice element type has no packed array encoder: " + v.Type().Elem().String())
	}
}

// sliceResizeCell checks the engine Error returned by a packed array resize;
// a failed resize means the buffer was never allocated and must not be
// written through operator_index.
func sliceResizeCell(err int64, cell string, size int) {
	if err != 0 {
		log.Panic("packed array resize failed",
			zap.String("cell", cell),
			zap.Int("size", size),
			zap.Int64("error", err),
		)
	}
}

// sliceEncodeReturnVariant is the varcall-return counterpart of
// sliceEncodeReturn: the cell is built through the shared return builder and
// then moved into the Variant by the generated per-type variant encoder,
// exactly like the packed cell cases of
// GDExtensionVariantPtrFromReflectValue — the engine takes ownership of the
// cell, so no further release happens here.
func sliceEncodeReturnVariant(rOut GDExtensionUninitializedVariantPtr, v reflect.Value) {
	switch reflect.Zero(v.Type().Elem()).Interface().(type) {
	case byte:
		var cell PackedByteArray
		sliceEncodeReturn((GDExtensionUninitializedTypePtr)(cell.NativePtr()), v)
		PackedByteArrayEncoder.EncodeVariantPtrArg(cell, rOut)
	case int32:
		var cell PackedInt32Array
		sliceEncodeReturn((GDExtensionUninitializedTypePtr)(cell.NativePtr()), v)
		PackedInt32ArrayEncoder.EncodeVariantPtrArg(cell, rOut)
	case int, int64:
		var cell PackedInt64Array
		sliceEncodeReturn((GDExtensionUninitializedTypePtr)(cell.NativePtr()), v)
		PackedInt64ArrayEncoder.EncodeVariantPtrArg(cell, rOut)
	case float32:
		var cell PackedFloat32Array
		sliceEncodeReturn((GDExtensionUninitializedTypePtr)(cell.NativePtr()), v)
		PackedFloat32ArrayEncoder.EncodeVariantPtrArg(cell, rOut)
	case float64:
		var cell PackedFloat64Array
		sliceEncodeReturn((GDExtensionUninitializedTypePtr)(cell.NativePtr()), v)
		PackedFloat64ArrayEncoder.EncodeVariantPtrArg(cell, rOut)
	case string:
		var cell PackedStringArray
		sliceEncodeReturn((GDExtensionUninitializedTypePtr)(cell.NativePtr()), v)
		PackedStringArrayEncoder.EncodeVariantPtrArg(cell, rOut)
	case Vector2:
		var cell PackedVector2Array
		sliceEncodeReturn((GDExtensionUninitializedTypePtr)(cell.NativePtr()), v)
		PackedVector2ArrayEncoder.EncodeVariantPtrArg(cell, rOut)
	case Vector3:
		var cell PackedVector3Array
		sliceEncodeReturn((GDExtensionUninitializedTypePtr)(cell.NativePtr()), v)
		PackedVector3ArrayEncoder.EncodeVariantPtrArg(cell, rOut)
	case Vector4:
		var cell PackedVector4Array
		sliceEncodeReturn((GDExtensionUninitializedTypePtr)(cell.NativePtr()), v)
		PackedVector4ArrayEncoder.EncodeVariantPtrArg(cell, rOut)
	case Color:
		var cell PackedColorArray
		sliceEncodeReturn((GDExtensionUninitializedTypePtr)(cell.NativePtr()), v)
		PackedColorArrayEncoder.EncodeVariantPtrArg(cell, rOut)
	default:
		panic("slice element type has no packed array encoder: " + v.Type().Elem().String())
	}
}
