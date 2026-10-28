package core

import (
	"strings"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// The typed bind overloads (typed_binds.gen.go) share one registration core,
// classDBBuildTypedMetadata, with the name-string binds. These tests pin the
// validation parity the task requires: every gate the name-string path runs,
// in the same order, with the same diagnostics — all engine-free, because
// every gate precedes the engine-dependent metadata build (whose StringName
// constructors need a live engine; the engine tier covers the built-metadata
// paths under GODOT).
//
// The generated overloads' plumbing (type inference from (t, fn), wrapper
// instantiation, delegation) is smoke-tested at the "Class doesn't exist."
// and V_ gates, which fire before any engine call.

// TypedBindStub models a real bound user class: the registered GD name equals
// the Go type name, which the strict receiver gate enforces for non-virtual
// instance binds (the name-string path compares GetClassName() against the
// receiver type's Name()).
type TypedBindStub struct {
	owner *GodotObject
}

func (s *TypedBindStub) Self() *TypedBindStub { return s }

func (s *TypedBindStub) Destroy()                                     {}
func (s *TypedBindStub) GetGodotObjectOwner() *GodotObject            { return s.owner }
func (s *TypedBindStub) SetGodotObjectOwner(owner *GodotObject)       { s.owner = owner }
func (s *TypedBindStub) GetClassName() string                         { return "TypedBindStub" }
func (s *TypedBindStub) GetParentClassName() string                   { return "Object" }
func (s *TypedBindStub) AsGDExtensionObjectPtr() GDExtensionObjectPtr { return nil }
func (s *TypedBindStub) AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr {
	return nil
}
func (s *TypedBindStub) AsGDExtensionTypePtr() GDExtensionTypePtr { return nil }
func (s *TypedBindStub) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return nil
}

// Bound-by-expression methods. Count returns a retValueSlot member and Tag a
// RetCopyEncodable member so the generated overloads' constraint unions infer
// their return parameters; none of the bodies run in these tests (every case
// stops at a gate before dispatch is ever built or called).
func (s *TypedBindStub) Count() int64          { return 0 }
func (s *TypedBindStub) Tag() StringName       { return StringName{} }
func (s *TypedBindStub) Echo(n int64) float64  { return float64(n) }
func (s *TypedBindStub) Ready()                {}
func (s *TypedBindStub) Pair() (int64, string) { return 0, "" }

// V_Bad is flat (no underscore after the V_ prefix): the qualified-form
// rejection must fire for virtual binds and the prefix ban for instance binds.
func (s *TypedBindStub) V_Bad() {}

// V_Other_Ready is qualified but names a level the embedding chain does not
// resolve: the chain-parity check must reject it for gdMethodName "_ready"
// (whose candidate is V_TypedBindStub_Ready).
func (s *TypedBindStub) V_Other_Ready() {}

// V_TypedBindStub_Ready is the chain resolution's exact candidate for "_ready"
// at this class's level.
func (s *TypedBindStub) V_TypedBindStub_Ready() {}

// OtherBindStub is a second class-shaped stub for the receiver/class mismatch
// gate: its method expressions belong to a different class than TypedBindStub.
type OtherBindStub struct {
	owner *GodotObject
}

func (s *OtherBindStub) Destroy()                                     {}
func (s *OtherBindStub) GetGodotObjectOwner() *GodotObject            { return s.owner }
func (s *OtherBindStub) SetGodotObjectOwner(owner *GodotObject)       { s.owner = owner }
func (s *OtherBindStub) GetClassName() string                         { return "OtherBindStub" }
func (s *OtherBindStub) GetParentClassName() string                   { return "Object" }
func (s *OtherBindStub) AsGDExtensionObjectPtr() GDExtensionObjectPtr { return nil }
func (s *OtherBindStub) AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr {
	return nil
}
func (s *OtherBindStub) AsGDExtensionTypePtr() GDExtensionTypePtr { return nil }
func (s *OtherBindStub) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return nil
}
func (s *OtherBindStub) Ping() {}

func typedBindProbe() *ClassInfo {
	return &ClassInfo{
		Name:             "TypedBindStub",
		MethodMap:        map[string]*MethodBindAndClassMethodInfo{},
		VirtualMethodMap: map[string]*MethodBindAndClassMethodInfo{},
	}
}

// typedBindProbeTaken carries pre-bound names for duplicate rejection.
func typedBindProbeTaken() *ClassInfo {
	return &ClassInfo{
		Name: "TypedBindStub",
		MethodMap: map[string]*MethodBindAndClassMethodInfo{
			"taken_method": {},
		},
		VirtualMethodMap: map[string]*MethodBindAndClassMethodInfo{
			"taken_virtual": {},
		},
	}
}

func TestTypedBindRejectsNonFunction(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{"TypedBindStub": typedBindProbe()}, func() {
		stub := &TypedBindStub{}
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(stub, 42, "echo", METHOD_FLAGS_DEFAULT, nil, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a non-function bind value must fail the bind")
		}
		if !strings.Contains(msg, "typed bind requires a method expression or function") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

func TestTypedBindRejectsBareFuncAsInstanceBind(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{"TypedBindStub": typedBindProbe()}, func() {
		stub := &TypedBindStub{}
		// Instance shape expects the receiver as In(0); a bare function is
		// only valid under the static flag — with or without parameters.
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(stub, func(n int64) int64 { return n }, "echo", METHOD_FLAGS_DEFAULT, []string{"n"}, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a bare function under instance flags must fail the bind")
		}
		if !strings.Contains(msg, "typed instance bind requires a method expression with a receiver") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
		msg, fired = recoverFatal(func() {
			classDBBuildTypedMetadata(stub, func() {}, "noop", METHOD_FLAGS_DEFAULT, nil, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a parameterless bare function under instance flags must fail the bind")
		}
		if !strings.Contains(msg, "typed instance bind requires a method expression with a receiver") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

func TestTypedBindRejectsInstanceVPrefix(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{"TypedBindStub": typedBindProbe()}, func() {
		stub := &TypedBindStub{}
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(stub, (*TypedBindStub).V_Bad, "bad", METHOD_FLAGS_DEFAULT, nil, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a V_-prefixed instance bind must fail")
		}
		if !strings.Contains(msg, `method name cannot have a prefix of "V_"`) {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

func TestTypedBindRejectsVirtualWithoutPrefix(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{"TypedBindStub": typedBindProbe()}, func() {
		stub := &TypedBindStub{}
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(stub, (*TypedBindStub).Ready, "_ready", METHOD_FLAG_VIRTUAL, nil, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a virtual bind without the V_ prefix must fail")
		}
		if !strings.Contains(msg, `virtual method name must have a prefix of "V_"`) {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

func TestTypedBindRejectsFlatVirtualName(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{"TypedBindStub": typedBindProbe()}, func() {
		stub := &TypedBindStub{}
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(stub, (*TypedBindStub).V_Bad, "_bad", METHOD_FLAG_VIRTUAL, nil, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a flat V_<Method> virtual bind must fail")
		}
		if !strings.Contains(msg, "flat virtual method name is no longer supported") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

func TestTypedBindRejectsVirtualChainMismatch(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{"TypedBindStub": typedBindProbe()}, func() {
		stub := &TypedBindStub{}
		// V_Other_Ready is qualified but the chain resolution for "_ready"
		// finds V_TypedBindStub_Ready at this class's level, so the supplied
		// expression is not what the name-string path would bind.
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(stub, (*TypedBindStub).V_Other_Ready, "_ready", METHOD_FLAG_VIRTUAL, nil, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a virtual expression off the chain resolution must fail")
		}
		if !strings.Contains(msg, "typed virtual bind expression does not match the embedding-chain resolution") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

func TestTypedBindRejectsReceiverClassMismatch(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{"TypedBindStub": typedBindProbe()}, func() {
		// The expression belongs to OtherBindStub while t registers TypedBindStub.
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(&TypedBindStub{}, (*OtherBindStub).Ping, "ping", METHOD_FLAGS_DEFAULT, nil, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a receiver type from another class must fail the bind")
		}
		if !strings.Contains(msg, "class name did not match reciever type") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

func TestTypedBindRejectsDuplicateNames(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{"TypedBindStub": typedBindProbeTaken()}, func() {
		stub := &TypedBindStub{}
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(stub, (*TypedBindStub).Echo, "taken_method", METHOD_FLAGS_DEFAULT, []string{"n"}, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a duplicate method name must fail the typed bind")
		}
		if !strings.Contains(msg, "Binding duplicate method.") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
		msg, fired = recoverFatal(func() {
			classDBBuildTypedMetadata(stub, (*TypedBindStub).Echo, "taken_virtual", METHOD_FLAGS_DEFAULT, []string{"n"}, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a name bound as virtual must fail the typed bind")
		}
		if !strings.Contains(msg, "Method already bound as virtual.") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

func TestTypedBindRejectsUnregisteredClass(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{}, func() {
		stub := &TypedBindStub{}
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(stub, (*TypedBindStub).Echo, "echo", METHOD_FLAGS_DEFAULT, []string{"n"}, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a bind for an unregistered class must fail")
		}
		if !strings.Contains(msg, "Class doesn't exist.") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

func TestTypedBindRejectsSignatureOverflow(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{"TypedBindStub": typedBindProbe()}, func() {
		stub := &TypedBindStub{}
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(stub, (*TypedBindStub).Echo, "echo", METHOD_FLAGS_DEFAULT,
				[]string{"a", "b"}, nil, nil, nil)
		})
		if !fired {
			t.Fatal("more argument names than parameters must fail the typed bind")
		}
		if !strings.Contains(msg, "Method definition has more arguments than the actual method.") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
		msg, fired = recoverFatal(func() {
			classDBBuildTypedMetadata(stub, (*TypedBindStub).Ready, "ready", METHOD_FLAGS_DEFAULT,
				nil, []Variant{{}}, nil, nil)
		})
		if !fired {
			t.Fatal("more default arguments than parameters must fail the typed bind")
		}
		if !strings.Contains(msg, "Method definition has more default arguments than the actual method.") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

func TestTypedBindRejectsVariadicMisalignment(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{"TypedBindStub": typedBindProbe()}, func() {
		stub := &TypedBindStub{}
		// A variadic signature without the VARARG flag must fail, same as the
		// name-string path's validateBindSignature gate. The static shape is
		// the receiver-free form, so the variadic gate is what fires.
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(stub, func(nums ...int64) int64 { return 0 }, "sum",
				METHOD_FLAG_STATIC, nil, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a variadic signature without the VARARG flag must fail")
		}
		if !strings.Contains(msg, "go method and method flags are not variadic aligned") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

func TestTypedBindRejectsSecondReturnNonBool(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{"TypedBindStub": typedBindProbe()}, func() {
		stub := &TypedBindStub{}
		// Fires inside newGoMethodMetadataFromFunc before any engine-backed
		// StringName construction, so it stays engine-free.
		msg, fired := recoverFatal(func() {
			classDBBuildTypedMetadata(stub, (*TypedBindStub).Pair, "pair", METHOD_FLAGS_DEFAULT, nil, nil, nil, nil)
		})
		if !fired {
			t.Fatal("a non-bool second return value must fail the typed bind")
		}
		if !strings.Contains(msg, "method 2nd return value must be of type bool") {
			t.Fatalf("unexpected diagnostic: %s", msg)
		}
	})
}

// TestTypedBindOverloadDelegation smokes the generated overloads' plumbing:
// every type parameter infers from (t, fn) and the delegation reaches the
// shared core's gates — which fire before any engine call, so the full
// overload path runs engine-free.
func TestTypedBindOverloadDelegation(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{}, func() {
		stub := &TypedBindStub{}
		msg, fired := recoverFatal(func() {
			ClassDBBindValueMethod0(stub, (*TypedBindStub).Count, "count", nil, nil)
		})
		if !fired || !strings.Contains(msg, "Class doesn't exist.") {
			t.Fatalf("generated instance overload did not reach the class gate: fired=%v msg=%q", fired, msg)
		}
		msg, fired = recoverFatal(func() {
			ClassDBBindStaticValueMethod0(stub, func() int64 { return 0 }, "zero", nil, nil)
		})
		if !fired || !strings.Contains(msg, "Class doesn't exist.") {
			t.Fatalf("generated static overload did not reach the class gate: fired=%v msg=%q", fired, msg)
		}
		msg, fired = recoverFatal(func() {
			ClassDBBindVirtualValueMethod0(stub, (*TypedBindStub).Count, "_count", nil, nil)
		})
		if !fired || !strings.Contains(msg, `virtual method name must have a prefix of "V_"`) {
			t.Fatalf("generated virtual overload did not reach the V_ gate: fired=%v msg=%q", fired, msg)
		}
		msg, fired = recoverFatal(func() {
			ClassDBBindCopyRetMethod0(stub, (*TypedBindStub).Tag, "tag", nil, nil)
		})
		if !fired || !strings.Contains(msg, "Class doesn't exist.") {
			t.Fatalf("generated copy-ret overload did not reach the class gate: fired=%v msg=%q", fired, msg)
		}
	})
}
