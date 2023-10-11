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

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"

	"github.com/CannibalVox/cgoalloc"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

const (
	MaxAllocBytes = 2 << 28
)

// AllocCopy returns a duplicated data allocated in C memory.
func AllocCopy(src unsafe.Pointer, bytes int) unsafe.Pointer {
	switch {
	case bytes < 0:
		log.Panic("invalid memory",
			zap.Int("bytes", bytes),
		)
	case bytes >= MaxAllocBytes:
		log.Panic("memory too large",
			zap.Int("bytes", bytes),
		)
	}
	m := CallFunc_GDExtensionInterfaceMemAlloc(uint64(bytes))
	if m == nullptr {
		log.Panic("memory allocation failure",
			zap.Int("bytes", bytes),
		)
	}
	C.memcpy(m, src, C.size_t(bytes))
	return m
}

// func AllocTypedCopyDest[T any](dest *T, src *T) {
// 	var t T
// 	bytes := int(unsafe.Sizeof(t))
// 	AllocCopyDest(unsafe.Pointer(dest), unsafe.Pointer(src), bytes)
// }

// func AllocCopyDest(dest unsafe.Pointer, src unsafe.Pointer, bytes int) {
// 	switch {
// 	case bytes < 0:
// 		log.Panic("invalid memory",
// 			zap.Int("bytes", bytes),
// 		)
// 	case bytes >= MaxAllocBytes:
// 		log.Panic("memory too large",
// 			zap.Int("bytes", bytes),
// 		)
// 	}
// 	if dest == nullptr {
// 		log.Panic("destination cannot be nil",
// 			zap.Int("bytes", bytes),
// 		)
// 	}
// 	C.memcpy(dest, src, C.size_t(bytes))
// }

// AllocZeros returns zeroed out bytes allocated in C memory.
func AllocZeros(bytes int) unsafe.Pointer {
	switch {
	case bytes < 0:
		log.Panic("invalid memory",
			zap.Int("bytes", bytes),
		)
	case bytes >= MaxAllocBytes:
		log.Panic("memory too large",
			zap.Int("bytes", bytes),
		)
	}
	m := CallFunc_GDExtensionInterfaceMemAlloc(uint64(bytes))
	if m == nullptr {
		log.Panic("memory allocation failure",
			zap.Int("bytes", bytes),
		)
	}
	C.memset(m, 0, C.size_t(bytes))
	return m
}

// AllocArrayPtr
func AllocArrayPtr[T any](len int) *T {
	var t T

	bytes := int(unsafe.Sizeof(t)) * len

	return (*T)(AllocZeros(bytes))
}

// Alloc returns allocated memory in C memory.
func Alloc(bytes int) unsafe.Pointer {
	return CallFunc_GDExtensionInterfaceMemAlloc(uint64(bytes))
}

// Realloc returns allocated memory in C memory.
func Realloc(ptr unsafe.Pointer, bytes int) unsafe.Pointer {
	return CallFunc_GDExtensionInterfaceMemRealloc(ptr, uint64(bytes))
}

// Free frees allocated memory.
func Free(ptr unsafe.Pointer) {
	CallFunc_GDExtensionInterfaceMemFree(ptr)
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

type GodotAllocator struct{}

func (a *GodotAllocator) Malloc(size int) unsafe.Pointer {
	return Alloc(size)
}

func (a *GodotAllocator) Free(pointer unsafe.Pointer) {
	Free(pointer)
}

func (a *GodotAllocator) Destroy() error {
	return nil
}
