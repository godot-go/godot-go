package core

// #include <godot/gdextension_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassinit"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

var (
	gdClassType          = reflect.TypeOf((*GDClass)(nil)).Elem()
	gdExtensionClassType = reflect.TypeOf((*GDExtensionClass)(nil)).Elem()
	gdObjectType         = reflect.TypeOf((*Object)(nil)).Elem()
	gdArrayType          = reflect.TypeOf((*Array)(nil)).Elem()
	gdVariantType        = reflect.TypeOf((*Variant)(nil)).Elem()
	errorType            = reflect.TypeOf((*error)(nil)).Elem()
	refType              = reflect.TypeOf((*Ref)(nil)).Elem()
)

// reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs is called for each
// function call argument that needs to be translated when GDScript calls into Go.
// hasReceiver is false only for static binds, whose function has no receiver:
// the slice then carries exactly the decoded parameters and the decode
// bookkeeping ranges cover every entry.
//
// On a caller-induced decode failure it returns a *VarcallArgDecodeError naming
// the argument index and declared type, after releasing every owned argument it
// already decoded in this call. Ownership rule: the decoder owns the decoded
// prefix on failure, and GoMethodMetadata.Call owns the full set on success, so
// the two cleanup paths can never both fire.
func reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs(reciever GDClass, suppliedArgs []Variant, expectedArgTypes []reflect.Type, hasReceiver bool) ([]reflect.Value, error) {
	argsCount := len(expectedArgTypes)
	base := 0
	if hasReceiver {
		base = 1
	}
	args := make([]reflect.Value, argsCount+base)
	if hasReceiver {
		// add receiver instance as the first argument
		args[0] = reflect.ValueOf(reciever)
	}
	for i := 0; i < argsCount; i++ {
		v, err := convertVariantToGoTypeReflectValue(suppliedArgs[i], expectedArgTypes[i])
		if err != nil {
			// Entries args[base]..args[i+base-1] decoded successfully; the
			// remaining entries are still the zero reflect.Value, which
			// panics on Interface(), so only the decoded prefix is handed
			// to the releaser.
			destroyOwnedContainerArgs(args[base : i+base])
			releaseDecodedRefArgs(args[base : i+base])
			return nil, &VarcallArgDecodeError{
				Index:     i,
				ParamType: expectedArgTypes[i].String(),
				Err:       err,
			}
		}
		args[i+base] = v
	}
	log.Debug("argument converted",
		zap.String("args", spew.Sdump(args)),
	)
	return args, nil
}

// destroyOwnedContainerArgs releases container call arguments that
// convertVariantToGoTypeReflectValue decodes as owned copies. It must run only
// after any return value that may alias an argument has been encoded, so the
// encoded return already holds its own reference.
func destroyOwnedContainerArgs(args []reflect.Value) {
	for _, arg := range args {
		switch v := arg.Interface().(type) {
		case String:
			// convertVariantToGoTypeReflectValue decodes String through the
			// variant constructor, which acquires; the release was missing
			// here and leaked the constructed String per call. Handlers that
			// need the value past the call take their own Copy(), the same
			// discipline the StringName/NodePath borrows impose.
			v.Destroy()
		case Array:
			v.Destroy()
		case Dictionary:
			v.Destroy()
		case Callable:
			v.Destroy()
		case Signal:
			v.Destroy()
		case PackedByteArray:
			v.Destroy()
		case PackedInt32Array:
			v.Destroy()
		case PackedInt64Array:
			v.Destroy()
		case PackedFloat32Array:
			v.Destroy()
		case PackedFloat64Array:
			v.Destroy()
		case PackedStringArray:
			v.Destroy()
		case PackedVector2Array:
			v.Destroy()
		case PackedVector3Array:
			v.Destroy()
		case PackedVector4Array:
			v.Destroy()
		case PackedColorArray:
			v.Destroy()
		}
	}
}

// releaseDecodedRefArgs releases the reference every decoded Ref argument
// acquired during the call. Ref parameters are decoded as acquired (+1) so the
// object stays alive for the duration of the call, mirroring godot-cpp's
// PtrToArg<Ref<T>>::convert, whose temporary Ref is destroyed at call end by a
// C++ destructor. Go has no destructor and a finalizer runs too lazily to bound
// the reference (the acquired refs pile up and then release in a cascade at the
// first explicit GC), so both dispatch paths release deterministically here,
// after the return value has been encoded -- an echoed Ref return already took
// its own reference through the return slot.
//
// A Ref argument the callee needs past the call must be copied (NewRefAcquire
// of the pointee) rather than aliased: this release drops the decode's +1 and
// clears its finalizer, so an alias would dangle.
func releaseDecodedRefArgs(args []reflect.Value) {
	for _, arg := range args {
		switch v := arg.Interface().(type) {
		case Ref:
			v.Unref()
		}
	}
}

// refArgumentClass is the pointee class a declared Ref-typed parameter keys
// off. The generated RefXxx interfaces name their pointee by codegen
// convention; bare builtin.Ref is the protocol interface itself and stands
// for Ref<RefCounted>, accepting any reference-counted object.
func refArgumentClass(t reflect.Type) string {
	if t == refType {
		return "RefCounted"
	}
	return t.Name()[3:]
}

// decodeRefArgument decodes a Ref-typed method argument from a raw engine
// object pointer, identically on both dispatch paths (design D2/D3/D4).
//
// The declared Ref type keys the pointee class; an argument that is not an
// instance of it or a subclass -- including any non-reference-counted object
// -- is a caller error returned as a value, mirroring the binder's strict
// conversion gate rather than godot-cpp's silent null. The wrapper is built
// over the declared class so a subclass argument is sliced to the pointee the
// method asked for, and it acquires its own engine reference (released by
// finalizer), because the engine transfers nothing into an argument: godot-cpp
// does the same through PtrToArg<Ref<T>>::convert.
func decodeRefArgument(objPtr GDExtensionObjectPtr, t reflect.Type) (Ref, error) {
	className := refArgumentClass(t)
	if !ObjectInheritsClass(objPtr, className) {
		actual, nameErr := ObjectClassName((*GodotObject)(objPtr))
		if nameErr != nil {
			actual = "unknown"
		}
		return nil, fmt.Errorf("object of class %q is not a %s", actual, className)
	}
	refConstructor, ok := GDClassRefConstructors.Get(className)
	if !ok {
		return nil, fmt.Errorf("no ref constructor is registered for class %q", className)
	}
	owner := (*GodotObject)(objPtr)
	// Binding first, mirroring resolveObjectArgument: a user-defined extension
	// class instance decodes to the live Go wrapper the binding created for
	// it, so a Go method can reach the object's Go-side state through the
	// decoded Ref (the task-6.2 ExampleRef fixture pins this). The engine cast
	// above already proved the object inherits the declared class, so a
	// wrapper the constructor's own type assert rejects means the binding's
	// class model disagrees with the engine's; fall back rather than panic.
	if obj, err := ObjectFromInstanceBinding(owner); err == nil {
		if ref, usable := acquireRefOver(refConstructor, obj); usable {
			return ref, nil
		}
	}
	nativeConstructor, ok := GDNativeConstructors.Get(className)
	if !ok {
		return nil, fmt.Errorf("no wrapper constructor is registered for class %q", className)
	}
	return refConstructor(nativeConstructor(owner).(RefCounted)), nil
}

// acquireRefOver hands obj to the declared class's acquired Ref constructor.
// Generated constructors type-assert their argument to the declared pointee;
// a failed assert (or a panic from it) reports the wrapper as unusable so the
// caller can build one by declared class name instead.
func acquireRefOver(refConstructor RefCountedConstructor, obj Object) (ref Ref, ok bool) {
	defer func() {
		if recover() != nil {
			ref, ok = nil, false
		}
	}()
	rc, isRefcounted := obj.(RefCounted)
	if !isRefcounted {
		return nil, false
	}
	return refConstructor(rc), true
}

// resolveObjectArgument resolves a raw engine object pointer to the Go
// wrapper for an object-typed parameter: through the instance binding first,
// so a user-defined extension-class instance receives the wrapper registered
// for that exact object, and otherwise through the engine-class constructor
// keyed by the argument's runtime class name. Both dispatch paths decode
// objects through this one helper so they cannot drift apart.
func resolveObjectArgument(objPtr GDExtensionObjectPtr, declared reflect.Type) (Object, error) {
	engineObject := (*GodotObject)(objPtr)
	obj, err := ObjectFromInstanceBinding(engineObject)
	if err != nil {
		return nil, err
	}
	if reflect.TypeOf(obj).Implements(declared) {
		return obj, nil
	}
	// The binding resolved to a wrapper that does not satisfy the declared
	// parameter type, so fall back to building one by class name.
	gdsClass := obj.GetClass()
	defer gdsClass.Destroy()
	className := gdsClass.ToUtf8()
	constructor, ok := GDNativeConstructors.Get(className)
	if !ok {
		return nil, fmt.Errorf(
			"unsupported object class %q for parameter type %s", className, declared.String())
	}
	inst := constructor(engineObject).(Object)
	// A registered class name is not the same as a wrapper of the declared
	// type. Label is registered, and a LabelImpl still cannot be handed to a
	// method that asked for Sprite2D. Without this check the mismatch surfaces
	// one frame later as a reflect panic inside Func.Call -- outside the
	// decoder, outside the returned-error contract, and still an abort of the
	// host. The check belongs here rather than downstream.
	if !reflect.TypeOf(inst).Implements(declared) {
		return nil, fmt.Errorf(
			"unsupported object class %q for parameter type %s", className, declared.String())
	}
	return inst, nil
}

func convertVariantToGoTypeReflectValue(arg Variant, t reflect.Type) (reflect.Value, error) {
	var out reflect.Value
	if t == nil {
		return out, fmt.Errorf("type cannot be nil")
	}
	switch t.Kind() {
	case reflect.Bool:
		typedValue := arg.ToBool()
		log.Debug("varcall arg parsed",
			zap.Bool("value", typedValue),
			zap.String("type", "bool"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Int:
		typedValue := int(arg.ToInt64())
		log.Debug("varcall arg parsed",
			zap.Int("value", typedValue),
			zap.String("type", "int"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Int8:
		typedValue := int8(arg.ToInt64())
		log.Debug("varcall arg parsed",
			zap.Int8("value", typedValue),
			zap.String("type", "int8"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Int16:
		typedValue := int16(arg.ToInt64())
		log.Debug("varcall arg parsed",
			zap.Int16("value", typedValue),
			zap.String("type", "int16"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Int32:
		typedValue := int32(arg.ToInt64())
		log.Debug("varcall arg parsed",
			zap.Int32("value", typedValue),
			zap.String("type", "int32"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Int64:
		typedValue := arg.ToInt64()
		log.Debug("varcall arg parsed",
			zap.Int64("value", typedValue),
			zap.String("type", "int64"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Uint:
		typedValue := arg.ToUint()
		log.Debug("varcall arg parsed",
			zap.Uint("value", typedValue),
			zap.String("type", "uint"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Uint8:
		typedValue := arg.ToUint8()
		log.Debug("varcall arg parsed",
			zap.Uint8("value", typedValue),
			zap.String("type", "uint8"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Uint16:
		typedValue := arg.ToUint16()
		log.Debug("varcall arg parsed",
			zap.Uint16("value", typedValue),
			zap.String("type", "uint16"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Uint32:
		typedValue := arg.ToUint32()
		log.Debug("varcall arg parsed",
			zap.Uint32("value", typedValue),
			zap.String("type", "uint32"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Uint64:
		typedValue := arg.ToUint64()
		log.Debug("varcall arg parsed",
			zap.Uint64("value", typedValue),
			zap.String("type", "uint64"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Float32:
		typedValue := float32(arg.ToFloat64())
		log.Debug("varcall arg parsed",
			zap.Float32("value", typedValue),
			zap.String("type", "float32"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Float64:
		typedValue := arg.ToFloat64()
		log.Debug("varcall arg parsed",
			zap.Float64("value", typedValue),
			zap.String("type", "float64"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.String:
		// Ratified by the string-boundary-marshalling spec: a native Go
		// string parameter receives the converted text of any variant the
		// engine converts to STRING (String, StringName, NodePath arrive as
		// their text); variants with no STRING conversion are rejected at
		// the engine's call layer before reaching this decode. The result is
		// a retained Go copy, safe to hold past the call. Engine string
		// types in signatures cross as call-scoped borrows, except Godot
		// String, which the machinery owns and releases post-call; retain a
		// borrow past the call with its Copy() helper.
		typedValue := arg.ToGoString()
		log.Debug("varcall arg parsed",
			zap.Any("value", typedValue),
			zap.String("type", "string"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Slice:
		// support-native-slice-marshalling: the strict-conversion gate has
		// already accepted this variant against the PACKED_* registration
		// type (Nil is not convertible and never reaches here, pinned by
		// fixture). The elements are deep-copied out of an owned conversion
		// that is released before returning (design D1/D4).
		typedValue := sliceDecodeVarcallArg(arg, t.Elem())
		log.Debug("varcall arg parsed",
			zap.Int("len", typedValue.Len()),
			zap.String("type", t.String()),
		)
		return typedValue, nil
	case reflect.Interface:
		switch {
		case t.Implements(refType):
			// A nil argument has no class to guard against, so the zero Ref
			// is handed back before any registry lookup. The previous order
			// reached the lookup first and log.Fatal'ed on the empty name a
			// bare builtin.Ref derives -- killing the host for a legal null.
			if arg.IsNil() {
				return reflect.Zero(t), nil
			}
			engineObject := GodotObjectPtrFromVariant(&arg)
			if engineObject == nil {
				return reflect.Zero(t), nil
			}
			ref, err := decodeRefArgument((GDExtensionObjectPtr)(engineObject), t)
			if err != nil {
				return reflect.Value{}, err
			}
			log.Debug("varcall arg parsed",
				zap.String("value", arg.Stringify()),
				zap.String("type", t.Name()),
			)
			return reflect.ValueOf(ref), nil
		case t.Implements(gdObjectType):
			if arg.IsNil() {
				return reflect.Zero(t), nil
			}
			// Resolve the caller's object through its instance binding instead
			// of re-wrapping by class name. A user-defined extension class has
			// no GDNativeConstructors entry at all, so the name-keyed path
			// could never resolve one; the binding already holds the Go
			// wrapper created for this exact object.
			engineObject := GodotObjectPtrFromVariant(&arg)
			if engineObject == nil {
				return reflect.Zero(t), nil
			}
			obj, err := resolveObjectArgument((GDExtensionObjectPtr)(engineObject), t)
			if err != nil {
				return reflect.Value{}, fmt.Errorf(
					"cannot resolve object argument for parameter type %s: %w", t.String(), err)
			}
			log.Debug("varcall arg parsed",
				zap.String("class", obj.GetClassName()),
				zap.String("type", t.String()),
			)
			return reflect.ValueOf(obj), nil
		default:
			log.Panic("unsupported interface type",
				zap.Any("type", t),
			)
		}
	case reflect.Array:
		v := reflect.Zero(t)
		inst := v.Interface()
		switch inst.(type) {
		case Variant:
			v := VariantViewFromConstPtr(arg.NativeConstPtr())
			return reflect.ValueOf(v), nil
		case Vector2:
			v := arg.ToVector2()
			return reflect.ValueOf(v), nil
		case Vector2i:
			v := arg.ToVector2i()
			return reflect.ValueOf(v), nil
		case Vector3:
			v := arg.ToVector3()
			return reflect.ValueOf(v), nil
		case Vector3i:
			v := arg.ToVector3i()
			return reflect.ValueOf(v), nil
		case Vector4:
			v := arg.ToVector4()
			return reflect.ValueOf(v), nil
		case Vector4i:
			v := arg.ToVector4i()
			return reflect.ValueOf(v), nil
		case Rect2:
			v := Rect2Encoder.DecodeVariantPtr(arg.NativeConstPtr())
			return reflect.ValueOf(v), nil
		case Rect2i:
			v := Rect2iEncoder.DecodeVariantPtr(arg.NativeConstPtr())
			return reflect.ValueOf(v), nil
		case Transform2D:
			v := arg.ToTransform2D()
			return reflect.ValueOf(v), nil
		case Plane:
			v := arg.ToPlane()
			return reflect.ValueOf(v), nil
		case Quaternion:
			v := arg.ToQuaternion()
			return reflect.ValueOf(v), nil
		case AABB:
			v := arg.ToAABB()
			return reflect.ValueOf(v), nil
		case Basis:
			v := arg.ToBasis()
			return reflect.ValueOf(v), nil
		case Transform3D:
			v := arg.ToTransform3D()
			return reflect.ValueOf(v), nil
		case Projection:
			v := arg.ToProjection()
			return reflect.ValueOf(v), nil
		case Color:
			// support-variant-reflect-completeness (36-type audit): Color was
			// missing from the varcall decoder even though the mapper and both
			// return encoders handled it; the audit fixture crashed here.
			v := arg.ToColor()
			return reflect.ValueOf(v), nil
		// Container arguments are decoded as owned copies. The Go side acquires
		// a reference it releases after the call (see destroyOwnedContainerArgs,
		// invoked from GoMethodMetadata.Call). A borrowed byte-copy is unsafe
		// here: when the supplied Variant is a different container type than the
		// Go parameter (e.g. an Array[int] converted to a PackedInt64Array), the
		// type-from-variant constructor allocates a fresh buffer that the Variant
		// does not own, so destroying the constructed copy would leave the caller
		// reading freed memory. Holding the owned copy for the duration of the
		// call and releasing it on the way out is correct in both the same-type
		// and cross-type cases.
		case Array:
			v := arg.ToArray()
			return reflect.ValueOf(v), nil
		case PackedByteArray:
			v := arg.ToPackedByteArray()
			return reflect.ValueOf(v), nil
		case PackedInt32Array:
			v := arg.ToPackedInt32Array()
			return reflect.ValueOf(v), nil
		case PackedInt64Array:
			v := arg.ToPackedInt64Array()
			return reflect.ValueOf(v), nil
		case PackedFloat32Array:
			v := arg.ToPackedFloat32Array()
			return reflect.ValueOf(v), nil
		case PackedFloat64Array:
			v := arg.ToPackedFloat64Array()
			return reflect.ValueOf(v), nil
		case PackedStringArray:
			v := arg.ToPackedStringArray()
			return reflect.ValueOf(v), nil
		case PackedVector2Array:
			v := arg.ToPackedVector2Array()
			return reflect.ValueOf(v), nil
		case PackedVector3Array:
			v := arg.ToPackedVector3Array()
			return reflect.ValueOf(v), nil
		case PackedVector4Array:
			v := arg.ToPackedVector4Array()
			return reflect.ValueOf(v), nil
		case PackedColorArray:
			v := arg.ToPackedColorArray()
			return reflect.ValueOf(v), nil
		case RID:
			// support-variant-reflect-completeness (36-type audit): RID was
			// the one variant type the varcall decoder had no branch for; the
			// handle is a value, so the converted copy goes straight through.
			v := arg.ToRID()
			return reflect.ValueOf(v), nil
		case Dictionary:
			v := arg.ToDictionary()
			return reflect.ValueOf(v), nil
		case Signal:
			v := arg.ToSignal()
			return reflect.ValueOf(v), nil
		case Callable:
			v := arg.ToCallable()
			return reflect.ValueOf(v), nil
		case String:
			v := arg.ToString()
			return reflect.ValueOf(v), nil
		case StringName:
			v := arg.ToStringName()
			var borrow StringName
			for i := 0; i < StringNameSize; i++ {
				borrow[i] = v[i]
			}
			v.Destroy()
			return reflect.ValueOf(borrow), nil
		case NodePath:
			v := arg.ToNodePath()
			var borrow NodePath
			for i := 0; i < NodePathSize; i++ {
				borrow[i] = v[i]
			}
			v.Destroy()
			return reflect.ValueOf(borrow), nil
		default:
			log.Panic("unsupported array type",
				zap.Any("type", t),
			)
		}
	case reflect.Pointer:
		switch {
		// A Ref-protocol pointer type (*RefXxxImpl, *RefBase[T]) can never
		// arrive here: the type mapper refuses it at bind time. What remains
		// is the class-wrapper pointer parameter decoded as an object.
		case t.Implements(gdClassType):
			obj := arg.ToObject()
			// NOTE: add .Elem() if we want to support
			return reflect.ValueOf(obj), nil
		default:
			log.Panic("unsupported pointer type",
				zap.Any("type", t),
			)
		}
	default:
		log.Panic("unsupported type",
			zap.Any("type", t),
		)
	}
	return out, nil
}

func reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs(inst GDClass, suppliedArgs []GDExtensionConstTypePtr, expectedArgTypes []reflect.Type, hasReceiver bool) []reflect.Value {
	argsCount := len(expectedArgTypes)
	base := 0
	if hasReceiver {
		base = 1
	}
	args := make([]reflect.Value, argsCount+base)
	if hasReceiver {
		// add receiver instance as the first argument
		args[0] = reflect.ValueOf(inst)
	}
	for i := 0; i < argsCount; i++ {
		arg := suppliedArgs[i]
		t := expectedArgTypes[i]
		k := t.Kind()
		switch k {
		case reflect.Bool:
			typedValue := *(*bool)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Bool("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "bool"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Int:
			typedValue := *(*int)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Int("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Int8:
			typedValue := *(*int8)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Int8("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int8"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Int16:
			typedValue := *(*int16)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Int16("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int16"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Int32:
			typedValue := *(*int32)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Int32("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int32"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Int64:
			typedValue := *(*int64)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Int64("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int64"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Uint:
			typedValue := *(*uint)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Uint("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Uint8:
			typedValue := *(*uint8)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Uint8("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint8"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Uint16:
			typedValue := *(*uint16)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Uint16("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint16"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Uint32:
			typedValue := *(*uint32)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Uint32("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint32"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Uint64:
			typedValue := *(*uint64)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Uint64("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint64"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Float32:
			rawTypedValue := *(*float64)(arg)
			typedValue := (float32)(rawTypedValue)
			log.Debug("ptrcall arg parsed",
				zap.Float32("value", typedValue),
				zap.Float64("raw_value", rawTypedValue),
				zap.Int("arg_index", i),
				zap.String("type", "float32"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.Float64:
			typedValue := *(*float64)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Float64("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "float64"),
			)
			args[i+base] = reflect.ValueOf(typedValue)
		case reflect.String:
			typedValue := *(*String)(arg)
			str := typedValue.ToUtf8()
			log.Debug("ptrcall arg parsed",
				zap.Any("str", str),
				zap.Int("arg_index", i),
				zap.String("type", "string"),
			)
			args[i+base] = reflect.ValueOf(str)
		case reflect.Slice:
			// support-native-slice-marshalling: arg borrows the engine's
			// Packed*K cell for this call; elements are deep-copied into
			// Go-owned memory (design D1) and nothing is released here.
			sliceValue := sliceFromPackedCell(GDExtensionConstTypePtr(arg), t.Elem())
			log.Debug("ptrcall arg parsed",
				zap.Int("len", sliceValue.Len()),
				zap.Int("arg_index", i),
				zap.String("type", t.String()),
			)
			args[i+base] = sliceValue
		case reflect.Interface:
			// Dispatch on the declared type itself. The previous form was
			//   switch reflect.Zero(t).Interface().(type) { case Object: ... }
			// and could never reach the object case: the zero value of an
			// interface is a nil interface, and a Go type switch matches a
			// concrete case only against a non-nil dynamic type. Every
			// object-typed parameter therefore fell through to default and
			// panicked with "unsupported interface type".
			switch {
			case t.Implements(gdObjectType):
				// args[i] points at the value, and for an object the value is
				// the object pointer, so the cell must be read rather than
				// cast. The dead code this replaces cast the cell address
				// itself and would have wrapped a stack address as an object.
				objPtr := *(*GDExtensionObjectPtr)(arg)
				if objPtr == nil {
					// A null argument has no class name to ask for, and
					// object_get_class_name through a null pointer is not a
					// recoverable failure. Hand back the zero of the declared
					// interface so the callee sees a nil object.
					args[i+base] = reflect.Zero(t)
					break
				}
				// Same resolution order as varcall -- instance binding first,
				// engine-class constructor fallback keyed by the runtime class
				// name -- so a user-defined extension-class instance receives
				// its own Go wrapper on a direct typed call instead of
				// panicking, and wrapper identity matches call() dispatch.
				obj, err := resolveObjectArgument(objPtr, t)
				if err != nil {
					// The ptrcall fast path has no CallError slot, so the
					// failure is loud rather than contained; the message
					// carries the argument index and both class names so the
					// fatal-panic policy change has a clean diagnostic to
					// report.
					log.Panic(fmt.Sprintf("object argument %d rejected: %s", i, err.Error()),
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				log.Debug("ptrcall arg parsed",
					zap.Int("arg_index", i),
					zap.String("class", obj.GetClassName()),
				)
				args[i+base] = reflect.ValueOf(obj)
			case t.Implements(refType):
				// The engine Ref<T> struct is one pointer wide, so the cell
				// reads through ref_get_object whether GDScript handed a Ref
				// or a plain object; the declared pointee guard below covers
				// both. godot-cpp's PtrToArg<Ref<T>>::convert is the same.
				gdRefPtr := (GDExtensionConstRefPtr)(arg)
				objPtr := CallFunc_GDExtensionInterfaceRefGetObject(gdRefPtr)
				if objPtr == nil {
					// A null referenced object decodes to the zero Ref; no
					// engine call may dereference the null pointer.
					args[i+base] = reflect.Zero(t)
					break
				}
				ref, err := decodeRefArgument(objPtr, t)
				if err != nil {
					// Loud for the same ABI reason as the object branch
					// above: ptrcall reports no CallError.
					log.Panic(fmt.Sprintf("ref argument %d rejected: %s", i, err.Error()),
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				log.Debug("ptrcall arg parsed",
					zap.Int("arg_index", i),
					zap.String("type", "Ref"),
					zap.String("class_name", refArgumentClass(t)),
				)
				args[i+base] = reflect.ValueOf(ref)
			default:
				// The type name and index go in the panic value as well as the
				// structured fields, so a caller recovering from this can tell
				// what was refused without scraping the log.
				log.Panic(fmt.Sprintf("unsupported interface type for argument %d: %s", i, t.String()),
					zap.Int("arg_index", i),
					zap.Any("type", t),
				)
			}
		case reflect.Array:
			v := reflect.Zero(t)
			inst := v.Interface()
			switch inst.(type) {
			case Vector4:
				pV := (*Vector4)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector4WithVector4(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Vector2:
				pV := (*Vector2)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector2WithVector2(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Vector2i:
				pV := (*Vector2i)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector2iWithVector2i(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Vector3:
				pV := (*Vector3)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector3WithVector3(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Vector3i:
				pV := (*Vector3i)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector3iWithVector3i(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Rect2:
				pV := (*Rect2)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewRect2WithRect2(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Rect2i:
				pV := (*Rect2i)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewRect2iWithRect2i(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Transform2D:
				pV := (*Transform2D)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewTransform2DWithTransform2D(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Vector4i:
				pV := (*Vector4i)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector4iWithVector4i(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Plane:
				pV := (*Plane)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewPlaneWithPlane(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Quaternion:
				pV := (*Quaternion)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewQuaternionWithQuaternion(*pV)
				args[i+base] = reflect.ValueOf(v)
			case AABB:
				pV := (*AABB)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewAABBWithAABB(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Basis:
				pV := (*Basis)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewBasisWithBasis(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Transform3D:
				pV := (*Transform3D)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewTransform3DWithTransform3D(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Projection:
				pV := (*Projection)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewProjectionWithProjection(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Color:
				// Audit gap: Color had no ptrcall branch either (the crash
				// this fixture line exposed). POD value: borrowed raw copy,
				// same discipline as the StringName/NodePath cases below.
				v := *(*Color)(arg)
				args[i+base] = reflect.ValueOf(v)
			case Variant:
				v := VariantViewFromConstPtr((GDExtensionConstVariantPtr)(arg))
				args[i+base] = reflect.ValueOf(v)
			case PackedInt64Array:
				v := *(*PackedInt64Array)(arg)
				args[i+base] = reflect.ValueOf(v)
			case Array:
				v := *(*Array)(arg)
				args[i+base] = reflect.ValueOf(v)
			case PackedByteArray:
				v := *(*PackedByteArray)(arg)
				args[i+base] = reflect.ValueOf(v)
			case PackedInt32Array:
				v := *(*PackedInt32Array)(arg)
				args[i+base] = reflect.ValueOf(v)
			case PackedFloat32Array:
				v := *(*PackedFloat32Array)(arg)
				args[i+base] = reflect.ValueOf(v)
			case PackedFloat64Array:
				v := *(*PackedFloat64Array)(arg)
				args[i+base] = reflect.ValueOf(v)
			case PackedStringArray:
				v := *(*PackedStringArray)(arg)
				args[i+base] = reflect.ValueOf(v)
			case PackedVector2Array:
				v := *(*PackedVector2Array)(arg)
				args[i+base] = reflect.ValueOf(v)
			case PackedVector3Array:
				v := *(*PackedVector3Array)(arg)
				args[i+base] = reflect.ValueOf(v)
			case PackedVector4Array:
				pV := (*PackedVector4Array)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewPackedVector4ArrayWithPackedVector4Array(*pV)
				args[i+base] = reflect.ValueOf(v)
			case PackedColorArray:
				v := *(*PackedColorArray)(arg)
				args[i+base] = reflect.ValueOf(v)
			case String:
				v := *(*String)(arg)
				args[i+base] = reflect.ValueOf(v)
			case StringName:
				v := *(*StringName)(arg)
				args[i+base] = reflect.ValueOf(v)
			case NodePath:
				v := *(*NodePath)(arg)
				args[i+base] = reflect.ValueOf(v)
			case RID:
				// support-variant-reflect-completeness (36-type audit): the
				// interface exports no ref/unref for RID, so it crosses as a
				// pure value handle, borrowed for the duration of the call.
				v := *(*RID)(arg)
				args[i+base] = reflect.ValueOf(v)
			case Callable:
				// The engine keeps the cell's callable alive for the whole
				// call; the copy is a borrow, released nowhere because nothing
				// was acquired -- same discipline as the Array case above.
				v := *(*Callable)(arg)
				args[i+base] = reflect.ValueOf(v)
			case Signal:
				v := *(*Signal)(arg)
				args[i+base] = reflect.ValueOf(v)
			case Dictionary:
				v := *(*Dictionary)(arg)
				args[i+base] = reflect.ValueOf(v)
			default:
				log.Panic(fmt.Sprintf("MethodBind.Ptrcall reflected as array does not support type: %s", t.Name()))
			}
		case reflect.Struct:
			v := reflect.Zero(t)
			inst := v.Interface()
			switch inst.(type) {
			case Vector4:
				pV := (*Vector4)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector4WithVector4(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Vector2:
				pV := (*Vector2)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector2WithVector2(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Vector2i:
				pV := (*Vector2i)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector2iWithVector2i(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Vector3:
				pV := (*Vector3)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector3WithVector3(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Vector3i:
				pV := (*Vector3i)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector3iWithVector3i(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Rect2:
				pV := (*Rect2)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewRect2WithRect2(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Rect2i:
				pV := (*Rect2i)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewRect2iWithRect2i(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Transform2D:
				pV := (*Transform2D)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewTransform2DWithTransform2D(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Vector4i:
				pV := (*Vector4i)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewVector4iWithVector4i(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Plane:
				pV := (*Plane)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewPlaneWithPlane(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Quaternion:
				pV := (*Quaternion)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewQuaternionWithQuaternion(*pV)
				args[i+base] = reflect.ValueOf(v)
			case AABB:
				pV := (*AABB)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewAABBWithAABB(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Basis:
				pV := (*Basis)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewBasisWithBasis(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Transform3D:
				pV := (*Transform3D)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewTransform3DWithTransform3D(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Projection:
				pV := (*Projection)(unsafe.Pointer(arg))
				if pV == nil {
					log.Panic("GDExtensionConstTypePtr is nil",
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				v := NewProjectionWithProjection(*pV)
				args[i+base] = reflect.ValueOf(v)
			case Color:
				v := *(*Color)(arg)
				args[i+base] = reflect.ValueOf(v)
			case Variant:
				v := VariantViewFromConstPtr((GDExtensionConstVariantPtr)(arg))
				args[i+base] = reflect.ValueOf(v)
			default:
				// No Array-kind type reaches here that could be a Ref: the
				// former gdextension.Ref-prefixed block below it was dead for
				// the package names this codebase no longer has, and Ref-typed
				// parameters are Interface-kind and decode above.
				log.Panic("unsupported struct type",
					zap.Int("arg_index", i),
					zap.Any("type", t),
				)
			}
		case reflect.Pointer:
			switch {
			// A Ref-protocol pointer type can never arrive here: the type
			// mapper refuses it at bind time, replacing the
			// "TODO: missing ref pointer type" panic.
			case t.Implements(gdClassType):
				className := t.Elem().Name()
				constructor, ok := Internal.GDClassConstructors.Get(className)
				if !ok {
					log.Panic("TODO: missing user gd class constructor",
						zap.Int("arg_index", i),
						zap.Any("type", t),
						zap.Any("class_name", className),
					)
				}
				owner := (**GodotObject)(arg)
				obj := constructor(*owner)
				args[i+base] = reflect.ValueOf(obj)
			default:
				log.Panic("unsupported pointer type",
					zap.Int("arg_index", i),
					zap.Any("type", t),
				)
			}
		default:
			log.Panic(fmt.Sprintf("MethodBind.Ptrcall does not support kind %s and type: %s", k, t.Name()))
		}
	}
	log.Debug("argument converted",
		zap.String("args", spew.Sdump(args)),
	)
	return args
}

func validateReturnValues(reflectedRet []reflect.Value, returnStyle ReturnStyle, expectedReturnType reflect.Type) error {
	log.Debug("validateReturnValues called",
		zap.Any("return_style", returnStyle),
		zap.Any("expected_return_type", expectedReturnType),
	)
	switch len(reflectedRet) {
	case 0:
		if returnStyle != NoneReturnStyle {
			log.Panic("unexpected return style: expected none")
		}
		return nil
	case 1:
		if expectedReturnType == nil {
			log.Panic("no return value expected but 1 returned")
		}
		switch returnStyle {
		case NoneReturnStyle:
			log.Panic("expected no return value")
		case ValueReturnStyle:
			en := expectedReturnType.Name()
			if en != reflectedRet[0].Type().Name() {
				log.Panic("unexpected return type",
					zap.String("type", reflectedRet[0].Type().Name()),
				)
			}
		default:
			log.Panic("unexpected value returned")
		}
		return nil
	default:
		log.Panic("too many values returned", zap.Any("ret", reflectedRet))
	}
	return fmt.Errorf("unexpected code path reached")
}
