package core

// #include <godot/gdextension_interface.h>
// #include <stdlib.h>
import "C"
import (
	"sync/atomic"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// PropertyDescriptor is the class-registration input for a property exposed
// through the class get_property_list callback (openspec
// fix-class-property-list-lifecycle). Callers describe properties with plain
// Go strings; the engine string objects and the stable C cells backing them
// are built once per class registration and torn down with it.
type PropertyDescriptor struct {
	ClassName  string
	Type       GDExtensionVariantType
	Name       string
	Hint       uint32
	HintString string
	Usage      uint32
}

// classPropertySpec is the class-owned backing store for one exposed property.
//
// The GDExtension ABI treats the string fields of GDExtensionPropertyInfo as
// cells: pointers to memory holding an opaque engine pointer. Keeping those
// cells in C memory makes the addresses stable for the whole registration
// lifetime with no Go pinning — the GC never moves them — and lets every
// per-query array contain nothing but C data (design D1/D2).
type classPropertySpec struct {
	variantType GDExtensionVariantType
	hint        uint32
	usage       uint32
	// Each cell is a calloc'd 8-byte slot written by the matching engine
	// constructor; the engine object it holds is destroyed exactly once in
	// destroy(), together with the cell.
	nameCell  unsafe.Pointer
	classCell unsafe.Pointer
	hintCell  unsafe.Pointer
}

const cellSize = C.size_t(unsafe.Sizeof(uintptr(0)))

// newClassPropertySpec materializes the engine strings and stable cells for a
// descriptor. On allocation failure it returns ok=false (the caller reports
// an empty property list rather than corrupting registration).
func newClassPropertySpec(d PropertyDescriptor) (s classPropertySpec, ok bool) {
	s = classPropertySpec{
		variantType: d.Type,
		hint:        d.Hint,
		usage:       d.Usage,
		nameCell:    C.calloc(1, cellSize),
		classCell:   C.calloc(1, cellSize),
		hintCell:    C.calloc(1, cellSize),
	}
	if s.nameCell == nil || s.classCell == nil || s.hintCell == nil {
		s.destroy()
		return classPropertySpec{}, false
	}
	CallFunc_GDExtensionInterfaceStringNameNewWithUtf8Chars(
		(GDExtensionUninitializedStringNamePtr)(s.nameCell), d.Name)
	CallFunc_GDExtensionInterfaceStringNameNewWithUtf8Chars(
		(GDExtensionUninitializedStringNamePtr)(s.classCell), d.ClassName)
	CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars(
		(GDExtensionUninitializedStringPtr)(s.hintCell), d.HintString)
	return s, true
}

// buildClassPropertySpecs converts registration descriptors, panicking with
// the offending property on allocation failure (registration cannot proceed
// with a partial list).
func buildClassPropertySpecs(descriptors []PropertyDescriptor) []classPropertySpec {
	if len(descriptors) == 0 {
		return nil
	}
	specs := make([]classPropertySpec, 0, len(descriptors))
	for _, d := range descriptors {
		s, ok := newClassPropertySpec(d)
		if !ok {
			// Release what was built so far before the panic unwinds.
			for i := range specs {
				specs[i].destroy()
			}
			log.Panic("failed to allocate class property backing store",
				zap.String("property", d.Name))
		}
		specs = append(specs, s)
	}
	return specs
}

// destroy frees the engine string objects behind the cells and the cells
// themselves. Safe to call on a zero-value spec and idempotent.
func (s *classPropertySpec) destroy() {
	stringNameDestructor := (GDExtensionPtrDestructor)(CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING_NAME))
	stringDestructor := (GDExtensionPtrDestructor)(CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING))
	if s.nameCell != nil {
		if stringNameDestructor != nil {
			CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(s.nameCell))
		}
		C.free(s.nameCell)
		s.nameCell = nil
	}
	if s.classCell != nil {
		if stringNameDestructor != nil {
			CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(s.classCell))
		}
		C.free(s.classCell)
		s.classCell = nil
	}
	if s.hintCell != nil {
		if stringDestructor != nil {
			CallFunc_GDExtensionPtrDestructor(stringDestructor, (GDExtensionTypePtr)(s.hintCell))
		}
		C.free(s.hintCell)
		s.hintCell = nil
	}
}

// fillCProperty renders the spec into one element of a query-owned C array.
// The element string fields receive the cell addresses; the engine deep-copies
// through them during the query (Object::get_property_list), so the cells must
// only remain valid until the paired free callback — class lifetime covers
// that trivially.
func (s *classPropertySpec) fillCProperty(dst *C.GDExtensionPropertyInfo) {
	dst._type = C.GDExtensionVariantType(s.variantType)
	dst.name = C.GDExtensionStringNamePtr(s.nameCell)
	dst.class_name = C.GDExtensionStringNamePtr(s.classCell)
	dst.hint = C.uint32_t(s.hint)
	dst.hint_string = C.GDExtensionStringPtr(s.hintCell)
	dst.usage = C.uint32_t(s.usage)
}

// propertyArrayAllocs / propertyArrayFrees count query-owned C property
// arrays. The arrays live in C memory, so runtime.ReadMemStats cannot see a
// missing release; the release-accounting gate (openspec
// fix-class-property-list-lifecycle) reads these counters instead, mirroring
// gdclassinit.BindingCreateCount/BindingFreeCount.
var (
	propertyArrayAllocs atomic.Int64
	propertyArrayFrees  atomic.Int64
)

// PropertyArrayAllocCount reports how many query-owned property arrays have
// been allocated since process start.
func PropertyArrayAllocCount() int64 { return propertyArrayAllocs.Load() }

// PropertyArrayFreeCount reports how many query-owned property arrays the
// paired free_property_list2 callback has released since process start.
func PropertyArrayFreeCount() int64 { return propertyArrayFrees.Load() }

// allocPropertyArray returns a query-owned C array of len(specs) elements
// plus a slice over it, or (nil, nil) for an empty list / allocation failure.
func allocPropertyArray(specs []classPropertySpec) (*C.GDExtensionPropertyInfo, []C.GDExtensionPropertyInfo) {
	if len(specs) == 0 {
		return nil, nil
	}
	size := C.size_t(unsafe.Sizeof(C.GDExtensionPropertyInfo{}))
	ptr := C.calloc(C.size_t(len(specs)), size)
	if ptr == nil {
		return nil, nil
	}
	arr := unsafe.Slice((*C.GDExtensionPropertyInfo)(ptr), len(specs))
	for i := range specs {
		specs[i].fillCProperty(&arr[i])
	}
	propertyArrayAllocs.Add(1)
	return (*C.GDExtensionPropertyInfo)(ptr), arr
}

// freePropertyArray releases one query-owned array from allocPropertyArray.
// The class-level cells and strings are untouched: they belong to ClassInfo
// and are destroyed with the registration.
func freePropertyArray(pList *C.GDExtensionPropertyInfo) {
	if pList == nil {
		return
	}
	C.free(unsafe.Pointer(pList))
	propertyArrayFrees.Add(1)
}

// RunPropertyCacheTeardownCheck drives the property-cache lifecycle — build
// descriptors, zero / one / many query alloc-free cycles, then class
// teardown — and reports whether it completed cleanly (openspec
// fix-class-property-list-lifecycle task 2.2). A double-free or a release of
// stale cells aborts the process inside the allocator, so a clean true return
// is the assertion; the counter balance additionally proves the query arrays
// from the cycles were each released exactly once. Real registration
// teardown (ClassDBUnregisterClass → ClassInfo.Destroy) then runs at
// extension shutdown for every class, after the run's many queries.
func RunPropertyCacheTeardownCheck() bool {
	descs := []PropertyDescriptor{
		{ClassName: "TeardownProbe", Type: GDEXTENSION_VARIANT_TYPE_FLOAT, Name: "probe_alpha", Hint: uint32(PROPERTY_HINT_NONE), HintString: "", Usage: uint32(PROPERTY_USAGE_DEFAULT)},
		{ClassName: "TeardownProbe", Type: GDEXTENSION_VARIANT_TYPE_VECTOR2, Name: "probe_beta", Hint: uint32(PROPERTY_HINT_NONE), HintString: "probe", Usage: uint32(PROPERTY_USAGE_DEFAULT)},
	}

	// Zero queries, then teardown.
	specs := buildClassPropertySpecs(descs)
	if len(specs) != len(descs) {
		return false
	}
	for i := range specs {
		specs[i].destroy()
	}

	// One query, released, then teardown.
	specs = buildClassPropertySpecs(descs)
	if arr, _ := allocPropertyArray(specs); arr == nil {
		return false
	} else {
		freePropertyArray(arr)
	}
	for i := range specs {
		specs[i].destroy()
	}

	// Many queries, then teardown, then an idempotent second teardown.
	specs = buildClassPropertySpecs(descs)
	balance := PropertyArrayAllocCount() - PropertyArrayFreeCount()
	for i := 0; i < 1000; i++ {
		arr, _ := allocPropertyArray(specs)
		if arr == nil {
			return false
		}
		freePropertyArray(arr)
	}
	if PropertyArrayAllocCount()-PropertyArrayFreeCount() != balance {
		return false
	}
	for i := range specs {
		specs[i].destroy()
	}
	for i := range specs {
		specs[i].destroy()
	}
	return true
}
