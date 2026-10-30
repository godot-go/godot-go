package pkg

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// In-engine tests for the ptrcall object-argument decoder (openspec change
// fix-ptrcall-object-arg-decode).
//
// These drive the decoder directly rather than through a GDScript call site, on
// purpose. A typed GDScript object call only reaches ptrcall once
// fix-object-argument-type-metadata lands, so depending on that would leave this
// decoder untestable until a sibling change ships. Instead each test takes its
// object through the varcall path, which works today, and then builds the
// ptrcall argument cell by hand.
//
// The cell shape is the one the generator produces on the outbound side: a
// GDExtensionObjectPtr holding the object pointer, whose address goes in the
// slice. Getting this wrong is exactly the second defect this change fixes, so
// building it explicitly here is part of what is under test.
//
// Return codes follow the existing convention: 1 on success, negative per failed
// assertion, -900 on panic.

// decodeOne runs the decoder for a single declared parameter against a cell
// holding objPtr, and returns the decoded value.
func decodeOne(recv GDClass, objPtr GDExtensionObjectPtr, decl reflect.Type) reflect.Value {
	cell := objPtr
	args := []GDExtensionConstTypePtr{(GDExtensionConstTypePtr)(unsafe.Pointer(&cell))}
	vals := DecodePtrcallArgs(recv, args, []reflect.Type{decl})
	if len(vals) != 2 {
		panic(fmt.Sprintf("decoder returned %d values, want 2 (receiver + 1 argument)", len(vals)))
	}
	return vals[1]
}

// capturePanic runs fn and returns the panic message, or "" if none occurred.
func capturePanic(fn func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg = fmt.Sprint(r)
		}
	}()
	fn()
	return ""
}

// TestPtrcallDecodePlainObject decodes a Node cell against the Node interface.
// Before the fix this panicked "unsupported interface type": the zero-value type
// switch could never match a concrete case against a nil interface.
func (e *Example) TestPtrcallDecodePlainObject(node Node) int32 {
	return objectArgGuard("ptrcall_plain_object", func() int32 {
		if isNullObject(node) {
			return -1
		}
		wantID := node.GetInstanceId()

		decoded, ok := decodeOne(e, ObjectArgPtr(node), reflect.TypeFor[Node]()).Interface().(Node)
		if !ok || isNullObject(decoded) {
			return -2
		}
		if got := decoded.GetInstanceId(); got != wantID {
			fmt.Printf("  ptrcall_plain_object: FAIL: decoded id %d, want %d\n", got, wantID)
			return -3
		}
		fmt.Printf("  ptrcall_plain_object: decoded instance %d\n", wantID)
		return 1
	})
}

// TestPtrcallDecodeSubclass decodes a CircleShape2D against the base Shape2D
// interface. Resolution keys on the argument's runtime class name, which is what
// makes Godot's polymorphism work through a base-typed parameter.
func (e *Example) TestPtrcallDecodeSubclass(shape Shape2D) int32 {
	return objectArgGuard("ptrcall_subclass", func() int32 {
		if isNullObject(shape) {
			return -1
		}
		wantID := shape.GetInstanceId()

		decoded, ok := decodeOne(e, ObjectArgPtr(shape), reflect.TypeFor[Shape2D]()).Interface().(Shape2D)
		if !ok || isNullObject(decoded) {
			return -2
		}
		if got := decoded.GetInstanceId(); got != wantID {
			fmt.Printf("  ptrcall_subclass: FAIL: decoded id %d, want %d\n", got, wantID)
			return -3
		}
		cls := decoded.GetClass()
		defer cls.Destroy()
		if cls.ToUtf8() != "CircleShape2D" {
			fmt.Printf("  ptrcall_subclass: FAIL: decoded class %s, want CircleShape2D\n", cls.ToUtf8())
			return -4
		}
		fmt.Printf("  ptrcall_subclass: Shape2D parameter received instance %d as %s\n", wantID, cls.ToUtf8())
		return 1
	})
}

// TestPtrcallDecodeNullObject feeds a null cell and requires a nil decode with no
// engine call against the null pointer. The branch this covers was written before
// null was ever considered, because it never ran.
func (e *Example) TestPtrcallDecodeNullObject() int32 {
	return objectArgGuard("ptrcall_null_object", func() int32 {
		v := decodeOne(e, nil, reflect.TypeFor[Node]())
		if !v.IsValid() {
			return -2
		}
		if decoded, ok := v.Interface().(Node); ok && !isNullObject(decoded) {
			fmt.Printf("  ptrcall_null_object: FAIL: null cell decoded to a live object\n")
			return -3
		}
		fmt.Printf("  ptrcall_null_object: null cell decoded to a nil Node without touching the engine\n")
		return 1
	})
}

// TestPtrcallDecodeRefRegression guards the branch that already worked. Godot's
// Ref<T> is a single pointer member, so a Ref cell is read the same way and
// ref_get_object takes the cell address directly.
func (e *Example) TestPtrcallDecodeRefRegression(shape Shape2D) int32 {
	return objectArgGuard("ptrcall_ref_regression", func() int32 {
		if isNullObject(shape) {
			return -1
		}
		wantID := shape.GetInstanceId()

		decoded, ok := decodeOne(e, ObjectArgPtr(shape), reflect.TypeFor[RefShape2D]()).Interface().(RefShape2D)
		if !ok || decoded == nil || !decoded.IsValid() {
			return -2
		}
		defer decoded.Unref() // release the acquire the seam took; bypasses the dispatch cleanup
		if got := decoded.ToObject().GetInstanceId(); got != wantID {
			fmt.Printf("  ptrcall_ref_regression: FAIL: ref points at %d, want %d\n", got, wantID)
			return -3
		}
		fmt.Printf("  ptrcall_ref_regression: RefShape2D still resolves instance %d\n", wantID)
		return 1
	})
}

// notAnObjectAtAll implements neither Object nor Ref, so the decoder must refuse
// it loudly rather than produce something meaningless.
type notAnObjectAtAll interface {
	SomethingUnrelated()
}

// TestPtrcallDecodeUndecodableInterface requires a loud failure naming the
// argument index and the offending type.
//
// The decoder refuses this the only way it can: log.Panic. The test catches the
// panic and, because a PANIC log line arriving in `make test` output would
// otherwise read as a failure, announces the expectation first so the noise is
// recognisable as this test doing its job.
func (e *Example) TestPtrcallDecodeUndecodableInterface(node Node) int32 {
	return objectArgGuard("ptrcall_undecodable", func() int32 {
		if isNullObject(node) {
			return -1
		}
		fmt.Printf("  ptrcall_undecodable: about to feed a non-object interface; one expected PANIC log follows from the decoder refusing it\n")
		var msg string
		func() {
			msg = capturePanic(func() {
				decodeOne(e, ObjectArgPtr(node), reflect.TypeFor[notAnObjectAtAll]())
			})
		}()
		if msg == "" {
			fmt.Printf("  ptrcall_undecodable: FAIL: undecodable interface decoded without error\n")
			return -2
		}
		if !strings.Contains(msg, "unsupported interface type") || !strings.Contains(msg, "notAnObjectAtAll") {
			fmt.Printf("  ptrcall_undecodable: FAIL: message lacks the type name: %s\n", msg)
			return -3
		}
		fmt.Printf("  ptrcall_undecodable: PASS - refused loudly as required: %s\n", firstLine(msg))
		return 1
	})
}

func firstLine(s string) string {
	return strings.SplitN(s, "\n", 2)[0]
}

// TestPtrcallDecodeRelease collects the wrappers built during decoding while the
// engine is still up, so nothing reaches the exit leak check.
func (e *Example) TestPtrcallDecodeRelease() int32 {
	return objectArgGuard("ptrcall_decode_release", func() int32 {
		runtime.GC()
		runtime.GC()
		return 1
	})
}

// TestObjectArgFallbackBoundary pins the fallback boundary of
// typed-method-dispatch (task 1.3) on live registration data. An Object- or
// Ref-typed parameter sits outside the typed unions — the testdata fixtures
// typedbindobjectarg and typedbindrefarg prove the typed API rejects such a
// signature at compile time — because decoding it is per-call instance-binding
// work a raw load cannot express. Two things follow, and both are asserted here:
//
//   - no bind carrying an object or reference parameter serves dispatch through
//     a compiled closure (were the unions ever widened, a raw-loaded interface
//     would reach the handler in place of the resolved instance), and
//   - such a bind still dispatches through the reflective decoder, resolving the
//     receiver and the argument instance exactly as before the typed path existed.
//
// The classification mirrors the decoder's own boundary in
// reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs: an interface
// declaration implementing Object or Ref takes the instance-binding branch.
func (e *Example) TestObjectArgFallbackBoundary(node Node) int32 {
	return objectArgGuard("object_arg_fallback_boundary", func() int32 {
		if isNullObject(node) {
			return -1
		}

		objectType := reflect.TypeFor[Object]()
		refType := reflect.TypeFor[Ref]()
		declaresObjectArgument := func(decls []reflect.Type) bool {
			for _, decl := range decls {
				if decl == nil || decl.Kind() != reflect.Interface {
					continue
				}
				if decl.Implements(objectType) || decl.Implements(refType) {
					return true
				}
			}
			return false
		}

		objectBinds := 0
		for _, ci := range Internal.GDRegisteredGDClasses.Values() {
			if err := ci.HasError(); err != nil {
				continue
			}
			for _, binds := range []map[string]*MethodBindAndClassMethodInfo{ci.MethodMap, ci.VirtualMethodMap} {
				for gdName, mcmi := range binds {
					md := mcmi.GoMethodMetadata
					if md == nil || !declaresObjectArgument(md.GoArgumentTypes) {
						continue
					}
					objectBinds++
					d, found := DispatchFor(ci.Name, gdName)
					if !found {
						fmt.Printf("  object_arg_boundary: FAIL: %s.%s unresolvable through the dispatch seam\n", ci.Name, gdName)
						return -2
					}
					if d.Compiled() {
						fmt.Printf("  object_arg_boundary: FAIL: %s.%s carries a compiled closure despite its object argument\n", ci.Name, gdName)
						return -3
					}
				}
			}
		}
		if objectBinds == 0 {
			// The demo registers object-argument binds (the neighbouring
			// decoder fixtures' own binds do); zero means the walk lost its
			// subject and the assertion above would pass vacuously.
			fmt.Println("  object_arg_boundary: FAIL: no object-argument bind found to check")
			return -4
		}

		// Drive one through the entry the engine uses. The handler is the
		// plain-object decoder fixture, which returns 1 only when its own
		// decode of the argument cell produced the same instance.
		d, found := DispatchFor("Example", "test_ptrcall_decode_plain_object")
		if !found {
			fmt.Println("  object_arg_boundary: FAIL: test_ptrcall_decode_plain_object unresolvable")
			return -5
		}
		cell := ObjectArgPtr(node)
		cells := []GDExtensionConstTypePtr{(GDExtensionConstTypePtr)(unsafe.Pointer(&cell))}
		var ret [8]byte
		d.Ptrcall(e, cells, GDExtensionUninitializedTypePtr(&ret[0]))
		if code := int32(binary.LittleEndian.Uint64(ret[:])); code != 1 {
			fmt.Printf("  object_arg_boundary: FAIL: fallback dispatch returned %d, want 1\n", code)
			return -6
		}
		fmt.Printf("  object_arg_boundary: %d object/Ref-argument binds, all reflective; fallback dispatch resolved instance %d\n",
			objectBinds, node.GetInstanceId())
		return 1
	})
}
