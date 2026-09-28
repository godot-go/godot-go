package builtin

// The instance-binding slot is the engine's per-object, per-extension-library
// scratch pointer. godot_headers/godot/gdextension_interface.h documents it as an
// extension-defined opaque void*: the engine only stores it and hands it back,
// keyed by the library token, and calls the extension's create callback if no
// binding exists yet. The shape is therefore entirely this project's choice.
//
// This file owns that shape. See openspec change
// fix-user-defined-class-object-arg-decode for why two shapes in the same slot
// produced a SIGSEGV.

import (
	"fmt"
	"runtime"
	"runtime/cgo"

	. "github.com/godot-go/godot-go/pkg/ffi"
)

// ObjectBindingError reports that an instance-binding slot could not be resolved
// to a Go wrapper for the engine object.
//
// It exists so an unresolvable lookup is a value the caller can report rather
// than a panic. The varcall decode path turns it into engine-visible error
// output instead of aborting the process.
type ObjectBindingError struct {
	// Class is the Godot class name when it was known, empty otherwise.
	Class string

	// Reason explains what could not be resolved.
	Reason string
}

func (e *ObjectBindingError) Error() string {
	if e.Class == "" {
		return "instance binding unresolved: " + e.Reason
	}
	return "instance binding unresolved for class " + e.Class + ": " + e.Reason
}

// annotateClass fills in the class name on a binding error that lacks one, so the
// message a developer sees names the class involved.
func annotateClass(err error, class string) error {
	if be, ok := err.(*ObjectBindingError); ok && be.Class == "" {
		be.Class = class
	}
	return err
}

// objectFromBindingPtr decodes the instance-binding slot into the Go wrapper
// that stands for the engine object.
//
// The slot is a cgo.Handle value, which is what the generated setter asserts:
// cmd/generate/ffi/templatefunctions.go types any void* parameter named
// p_binding as cgo.Handle, and the registration sites store cgo.NewHandle(...)
// through it. Two value types arrive behind that handle: Wrapped (from
// SetConstructInfo) and *WrappedClassInstance (from WrappedPostInitialize).
// Both normalise to Object.
//
// The parameter is a uintptr rather than an unsafe.Pointer because the thing
// being decoded is a pointer-sized integer that must never be dereferenced, only
// reinterpreted as a handle. Taking uintptr keeps an unsafe conversion off the
// API surface, and lets the decode be unit-tested without cgo (this project's
// build does not allow cgo in _test.go files).
//
// A null slot, a handle that is no longer live, or a handle carrying something
// other than a recognised wrapper is a typed error. cgo.Handle.Value() panics on
// a handle that was never created or already deleted, so the recover below is
// what keeps a bad slot from taking the process down.
func objectFromBindingPtr(binding uintptr) (Object, error) {
	if binding == 0 {
		return nil, &ObjectBindingError{Reason: "binding slot is null"}
	}

	// Anything else ever written here (the address of a Go interface, say)
	// reads back as a huge, meaningless handle number, and Value() says so
	// loudly.
	var value any
	func() {
		defer func() {
			if recover() != nil {
				value = nil
			}
		}()
		value = cgo.Handle(binding).Value()
	}()
	if value == nil {
		return nil, &ObjectBindingError{
			Reason: fmt.Sprintf("binding handle %d is not live", binding),
		}
	}

	switch v := value.(type) {
	case *WrappedClassInstance:
		if v == nil || v.Instance == nil {
			return nil, &ObjectBindingError{Reason: "wrapped class instance holds no object"}
		}
		return v.Instance, nil
	case Object:
		return v, nil
	default:
		return nil, &ObjectBindingError{
			Reason: fmt.Sprintf("binding value is not a recognised wrapper: %T", value),
		}
	}
}

// ObjectFromInstanceBinding resolves an engine object to its Go wrapper by
// reading this library's instance-binding slot.
//
// It first asks for an existing binding without supplying callbacks, so the
// engine will not create one for an object this library has no callbacks for. If
// nothing is bound yet it looks up the callbacks registered for the object's
// class and asks again, which is what lets the engine create the binding.
//
// Every failure is an *ObjectBindingError naming the class where the class is
// known.
func ObjectFromInstanceBinding(engineObject *GodotObject) (Object, error) {
	if engineObject == nil {
		return nil, &ObjectBindingError{Reason: "engine object is null"}
	}

	if raw := CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		(GDExtensionObjectPtr)(engineObject),
		FFI.Token,
		nil,
	); raw != nil {
		return objectFromBindingPtr(uintptr(raw))
	}

	className, err := engineObjectClassName(engineObject)
	if err != nil {
		return nil, err
	}

	cbs, ok := GDExtensionBindingGDExtensionInstanceBindingCallbacks.Get(className)
	if !ok {
		return nil, &ObjectBindingError{
			Class:  className,
			Reason: "no instance-binding callbacks are registered for this class",
		}
	}

	// The callbacks struct is copied here, so the copy must stay put for the
	// duration of the call that reads through it.
	var pinner runtime.Pinner
	defer pinner.Unpin()
	cbsPtr := &cbs
	pinner.Pin(cbsPtr)
	pinner.Pin(FFI.Token)

	raw := CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		(GDExtensionObjectPtr)(engineObject),
		FFI.Token,
		cbsPtr,
	)
	runtime.KeepAlive(engineObject)
	runtime.KeepAlive(cbsPtr)
	runtime.KeepAlive(FFI.Token)

	if raw == nil {
		return nil, &ObjectBindingError{
			Class:  className,
			Reason: "the engine created no binding",
		}
	}
	obj, err := objectFromBindingPtr(uintptr(raw))
	if err != nil {
		return nil, annotateClass(err, className)
	}
	return obj, nil
}

// engineObjectClassName asks the engine for an object's runtime class name.
func engineObjectClassName(engineObject *GodotObject) (string, error) {
	var pinner runtime.Pinner
	defer pinner.Unpin()

	sn := StringName{}
	snPtr := sn.NativePtr()
	pinner.Pin(snPtr)
	defer sn.Destroy()

	ok := CallFunc_GDExtensionInterfaceObjectGetClassName(
		(GDExtensionConstObjectPtr)(engineObject),
		FFI.Library,
		(GDExtensionUninitializedStringNamePtr)(snPtr),
	)
	if ok == 0 {
		return "", &ObjectBindingError{Reason: "the engine refused to report the class name"}
	}
	return sn.ToUtf8(), nil
}

// godotObjectPtrFromVariant extracts the engine object pointer an OBJECT variant
// carries, without resolving its instance binding. Splitting this out is what
// lets a caller reach ObjectFromInstanceBinding, which needs the engine pointer
// rather than the wrapper that resolving the binding would produce.
func godotObjectPtrFromVariant(v *Variant) *GodotObject {
	var pinner runtime.Pinner
	defer pinner.Unpin()

	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	var engineObject *GodotObject
	engineObjectPtr := &engineObject
	pinner.Pin(engineObjectPtr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(engineObjectPtr),
		v.NativePtr(),
	)
	return engineObject
}

// ObjectFromVariant resolves an OBJECT variant to the Go wrapper standing for
// the engine object it holds.
//
// This is the error-returning form of Variant.ToObject, intended for decode
// paths that must report a failure to the engine rather than panic or crash.
func ObjectFromVariant(v *Variant) (Object, error) {
	if v == nil || v.IsNil() {
		return nil, &ObjectBindingError{Reason: "variant is null"}
	}
	return ObjectFromInstanceBinding(godotObjectPtrFromVariant(v))
}
