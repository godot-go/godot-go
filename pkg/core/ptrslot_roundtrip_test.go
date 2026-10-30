package core

import (
	"encoding/binary"
	"math"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// typed-method-dispatch task 1.2 RED tests.
//
// Unit tier (engine-free): the typed ptrcall arg decoder for native-slot
// builtins is a plain raw load, `*(*A)(ptr)`, exactly matching the loads the
// reflective ptrcall decoder performs for every member of the native family
// (method_bind_reflect.go lines 602-707 scalars, 973-1039 containers/strings,
// 1197-1199 Color, 1200-1202 Variant). These tests plant bytes into an
// engine-medium cell and assert the generic loader returns the identical
// native value; the reflective concrete equivalents that hand through engine
// constructors (Vector2 family) are byte-identical to those constructors'
// output for POD values, asserted in the engine tier alongside them.
//
// Excluded on purpose (documented by their omission from the family):
//   - float32: reflective decodes double-load + truncate; a raw *float32 load
//     breaks that parity, so float32 stays on the reflective path.
//   - Go string, slices, Object/Ref-typed parameters: UTF-8 conversion, deep
//     copy and binding resolution are per-call work that cannot be a uniform
//     raw load; they follow the name-string bind API's reflective path.

// typedSlotLoad[A nativeSlot] and storeTypedRet[R retStore] are the generic
// decoder/encoder the typed closure family instantiates per concrete type
// (defined in method_bind_typed.go); their absence is the RED gate.

func TestTypedSlotLoadByteParity(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		// Engine scalar slots are int64s; the reflective decoder loads the
		// low byte as bool (nativeSlotCell Cell built the same way).
		var cell [8]byte
		binary.LittleEndian.PutUint64(cell[:], 1)
		got := loadTypedSlot[bool](GDExtensionConstTypePtr(&cell[0]))
		if got != true {
			t.Fatalf("got %v, want true", got)
		}
	})
	t.Run("int", func(t *testing.T) {
		var cell [8]byte
		binary.LittleEndian.PutUint64(cell[:], math.MaxUint64)
		got := loadTypedSlot[int](GDExtensionConstTypePtr(&cell[0]))
		if got != -1 {
			t.Fatalf("got %d, want -1", got)
		}
	})
	t.Run("int64", func(t *testing.T) {
		var cell [8]byte
		binary.LittleEndian.PutUint64(cell[:], 77)
		got := loadTypedSlot[int64](GDExtensionConstTypePtr(&cell[0]))
		if got != 77 {
			t.Fatalf("got %d, want 77", got)
		}
	})
	t.Run("uint64", func(t *testing.T) {
		var cell [8]byte
		binary.LittleEndian.PutUint64(cell[:], math.MaxUint64)
		got := loadTypedSlot[uint64](GDExtensionConstTypePtr(&cell[0]))
		if got != math.MaxUint64 {
			t.Fatalf("got %d, want MaxUint64", got)
		}
	})
	t.Run("float64", func(t *testing.T) {
		var cell [8]byte
		binary.LittleEndian.PutUint64(cell[:], math.Float64bits(1.5))
		got := loadTypedSlot[float64](GDExtensionConstTypePtr(&cell[0]))
		if got != 1.5 {
			t.Fatalf("got %v, want 1.5", got)
		}
	})
	t.Run("Vector2", func(t *testing.T) {
		// Vector2 is [8]uint8: two little-endian float32s (engine real_t =
		// float32 in non-double-precision builds). The reflective fixture
		// round enters through the engine copy constructor, but the resulting
		// native bytes are the copied cell; assert the raw load reproduces
		// the bytes exactly.
		var cell [8]byte
		binary.LittleEndian.PutUint32(cell[0:4], math.Float32bits(1.25))
		binary.LittleEndian.PutUint32(cell[4:8], math.Float32bits(3.5))
		got := loadTypedSlot[Vector2](GDExtensionConstTypePtr(&cell[0]))
		if string(got[:]) != string(cell[:]) {
			t.Fatalf("Vector2 bytes %v, want %v", got, cell)
		}
	})
	t.Run("Color", func(t *testing.T) {
		// Color's reflective case is a raw load (1197-1199); byte-parity with
		// it is exactly the generic loader's behavior.
		var cell [16]byte
		binary.LittleEndian.PutUint32(cell[0:4], math.Float32bits(0.1))
		binary.LittleEndian.PutUint32(cell[4:8], math.Float32bits(0.2))
		binary.LittleEndian.PutUint32(cell[8:12], math.Float32bits(0.3))
		binary.LittleEndian.PutUint32(cell[12:16], math.Float32bits(0.4))
		got := loadTypedSlot[Color](GDExtensionConstTypePtr(&cell[0]))
		if string(got[:]) != string(cell[:]) {
			t.Fatalf("Color bytes %v, want %v", got, cell)
		}
	})
	t.Run("Array", func(t *testing.T) {
		// Array's reflective case is a raw load (976-978), borrow semantics,
		// nothing acquired, so the loader needs no constructor and no pin.
		var cell [8]byte
		binary.LittleEndian.PutUint64(cell[:], 0xDEADBEEFCAFE)
		got := loadTypedSlot[Array](GDExtensionConstTypePtr(&cell[0]))
		if string(got[:]) != string(cell[:]) {
			t.Fatalf("Array bytes %v, want %v", got, cell)
		}
	})
	t.Run("String", func(t *testing.T) {
		// String's reflective case is a raw load (1013-1015): the cell's
		// native struct is copied with no engine reference acquired.
		var cell [8]byte
		binary.LittleEndian.PutUint64(cell[:], 0x0102030405060708)
		got := loadTypedSlot[String](GDExtensionConstTypePtr(&cell[0]))
		if string(got[:]) != string(cell[:]) {
			t.Fatalf("String bytes %v, want %v", got, cell)
		}
	})
	t.Run("StringName", func(t *testing.T) {
		var cell [8]byte
		binary.LittleEndian.PutUint64(cell[:], 9)
		got := loadTypedSlot[StringName](GDExtensionConstTypePtr(&cell[0]))
		if string(got[:]) != string(cell[:]) {
			t.Fatalf("StringName bytes %v, want %v", got, cell)
		}
	})
	t.Run("RID", func(t *testing.T) {
		var cell [8]byte
		binary.LittleEndian.PutUint64(cell[:], 42)
		got := loadTypedSlot[RID](GDExtensionConstTypePtr(&cell[0]))
		if string(got[:]) != string(cell[:]) {
			t.Fatalf("RID bytes %v, want %v", got, cell)
		}
	})
	t.Run("Callable", func(t *testing.T) {
		// Callable/Signal are 16-byte value handles stolen as a copy per
		// the reflective case comments ("the copy is a borrow").
		var cell [16]byte
		binary.LittleEndian.PutUint64(cell[0:8], 1)
		binary.LittleEndian.PutUint64(cell[8:16], 2)
		got := loadTypedSlot[Callable](GDExtensionConstTypePtr(&cell[0]))
		if string(got[:]) != string(cell[:]) {
			t.Fatalf("Callable bytes %v, want %v", got, cell)
		}
	})
}

func TestTypedRetStoreByteParity(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		var dst [8]byte
		storeTypedRet[bool](GDExtensionTypePtr(&dst[0]), true)
		if binary.LittleEndian.Uint64(dst[:]) != 1 {
			t.Fatalf("dst %v, want low byte 1", dst)
		}
	})
	t.Run("int64", func(t *testing.T) {
		var dst [8]byte
		storeTypedRet[int64](GDExtensionTypePtr(&dst[0]), -5)
		if binary.LittleEndian.Uint64(dst[:]) != binary.LittleEndian.Uint64([]byte{0xfb, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
			t.Fatalf("dst %v, want -5 little-endian", dst)
		}
	})
	t.Run("float64", func(t *testing.T) {
		var dst [8]byte
		storeTypedRet[float64](GDExtensionTypePtr(&dst[0]), 1.5)
		if binary.LittleEndian.Uint64(dst[:]) != math.Float64bits(1.5) {
			t.Fatalf("dst %v, want 1.5 bits", dst)
		}
	})
	t.Run("Vector2", func(t *testing.T) {
		var dst [8]byte
		var src Vector2
		binary.LittleEndian.PutUint32(src[0:4], math.Float32bits(1.25))
		binary.LittleEndian.PutUint32(src[4:8], math.Float32bits(3.5))
		storeTypedRet[Vector2](GDExtensionTypePtr(&dst[0]), src)
		if string(dst[:]) != string(src[:]) {
			t.Fatalf("dst %v, want %v", dst, src)
		}
	})
	t.Run("String", func(t *testing.T) {
		// String returns raw-store per the audited matrix
		// (StringEncoder.EncodeTypePtrArg), so the ret family contains it.
		var dst [8]byte
		var src String
		binary.LittleEndian.PutUint64(src[:], 0x0102030405060708)
		storeTypedRet[String](GDExtensionTypePtr(&dst[0]), src)
		if string(dst[:]) != string(src[:]) {
			t.Fatalf("dst %v, want %v", dst, src)
		}
	})
}
