package gdextension

/*
#include <godot/gdextension_interface.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdnative"

	"github.com/CannibalVox/cgoalloc"
)

// AllocCopy returns a duplicated data allocated in C memory.
func AllocCopy(src unsafe.Pointer, bytes int) unsafe.Pointer {
	m := GDExtensionInterface_mem_alloc(internal.gdnInterface, uint64(bytes))

	C.memcpy(m, src, C.size_t(bytes))

	return m
}

// AllocZeros returns zeroed out bytes allocated in C memory.
func AllocZeros(bytes int) unsafe.Pointer {
	m := GDExtensionInterface_mem_alloc(internal.gdnInterface, uint64(bytes))

	C.memset(m, 0, C.size_t(bytes))

	return m
}

// AllocArrayPtr
func AllocArrayPtr[T any](len int) *T {
	var t T

	bytes := int(unsafe.Sizeof(t)) * len

	return (*T)(AllocZeros(bytes))
}

// SliceHeaderDataPtr
func SliceHeaderDataPtr[A any, R any](args []*A) *R {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&args))

	if header == nil {
		return (*R)(nullptr)
	}

	return (*R)(unsafe.Pointer(header.Data))
}

// Alloc returns allocated memory in C memory.
func Alloc(bytes int) unsafe.Pointer {
	return GDExtensionInterface_mem_alloc(internal.gdnInterface, uint64(bytes))
}

// Realloc returns allocated memory in C memory.
func Realloc(ptr unsafe.Pointer, bytes int) unsafe.Pointer {
	return GDExtensionInterface_mem_realloc(internal.gdnInterface, ptr, uint64(bytes))
}

// Free frees allocated memory.
func Free(ptr unsafe.Pointer) {
	GDExtensionInterface_mem_free(internal.gdnInterface, ptr)
}

func allocCopyVariantPtrSliceToUnsafePointerArray(ptrs []*Variant) unsafe.Pointer {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&ptrs))

	return AllocCopy(unsafe.Pointer(header.Data), header.Len)
}

// AllocCopyVariantPtrSliceAsGDExtensionVariantPtrPtr
func AllocCopyVariantPtrSliceAsGDExtensionVariantPtrPtr(ptrs []*Variant) *GDExtensionConstVariantPtr {
	copiedPtrs := allocCopyVariantPtrSliceToUnsafePointerArray(ptrs)

	header := (*reflect.SliceHeader)(unsafe.Pointer(&copiedPtrs))

	return (*GDExtensionConstVariantPtr)(unsafe.Pointer(header.Data))
}

var _ cgoalloc.Allocator = &GodotAllocator{}

type GodotAllocator struct {}

func (a *GodotAllocator) Malloc(size int) unsafe.Pointer {
	return Alloc(size)
}

func (a *GodotAllocator) Free(pointer unsafe.Pointer) {
	Free(pointer)
}

func (a *GodotAllocator) Destroy() error {
	return nil
}
