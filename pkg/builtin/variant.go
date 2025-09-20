package builtin

// #include <godot/gdextension_interface.h>
import "C"
import (
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"github.com/godot-go/godot-go/pkg/util"
	"go.uber.org/zap"
)

const (
	VariantSize = 24
)

type Variant [VariantSize]uint8

func VariantInitBindings() {
	log.Debug("VariantInitBindings called")
	for i := GDExtensionVariantType(1); i < GDEXTENSION_VARIANT_TYPE_VARIANT_MAX; i++ {
		variantFromTypeConstructor[i] = CallFunc_GDExtensionInterfaceGetVariantFromTypeConstructor(i)
		typeFromVariantConstructor[i] = CallFunc_GDExtensionInterfaceGetVariantToTypeConstructor(i)
	}
	initPrimativeTypeEncoders()
	initBuiltinClassEncoders()
	builtinClassesInitBindings()
}

func NewVariantNativeCopy(NativeConstPtr GDExtensionConstVariantPtr) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	CallFunc_GDExtensionInterfaceVariantNewCopy(ptr, NativeConstPtr)
	return ret
}

func NewVariantCopy(dst, src Variant) {
	CallFunc_GDExtensionInterfaceVariantNewCopy((GDExtensionUninitializedVariantPtr)(dst.NativePtr()), src.NativeConstPtr())
}

func NewVariantNil() Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrWithNil(ptr)
	return ret
}

func GDExtensionVariantPtrWithNil(rOut GDExtensionUninitializedVariantPtr) {
	pnr.Pin(rOut)
	CallFunc_GDExtensionInterfaceVariantNewNil(rOut)
}

func NewVariantCopyWithGDExtensionConstVariantPtr(ptr GDExtensionConstVariantPtr) Variant {
	ret := Variant{}
	typedSrc := (*[VariantSize]uint8)(ptr)
	pnr.Pin(ptr)
	for i := range VariantSize {
		ret[i] = typedSrc[i]
	}
	return ret
}

func NewVariantGodotObject(owner *GodotObject) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	pnr.Pin(owner)
	GDExtensionVariantPtrFromGodotObjectPtr(owner, ptr)
	return ret
}

func GDExtensionVariantPtrFromGodotObjectPtr(owner *GodotObject, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	pnr.Pin(rOut)
	ownerPtr := unsafe.Pointer(&owner)
	pnr.Pin(ownerPtr)
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(ownerPtr),
	)
}

func (c *Variant) ToObject() Object {
	if c.IsNil() {
		return nil
	}
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	var engineObject *GodotObject
	engineObjectPtr := &engineObject
	pnr.Pin(engineObjectPtr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(engineObjectPtr),
		c.NativePtr(),
	)
	ret := getObjectInstanceBinding(engineObject)
	return ret
}

func getObjectInstanceBinding(engineObject *GodotObject) Object {
	if engineObject == nil {
		return nil
	}
	// Get existing instance binding, if one already exists.
	instPtr := (*Object)(CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		(GDExtensionObjectPtr)(engineObject),
		FFI.Token,
		nil))
	if instPtr != nil && *instPtr != nil {
		return *instPtr
	}
	snClassName := StringName{}
	snClassNamePtr := snClassName.NativePtr()
	pnr.Pin(snClassNamePtr)
	cok := CallFunc_GDExtensionInterfaceObjectGetClassName(
		(GDExtensionConstObjectPtr)(engineObject),
		FFI.Library,
		(GDExtensionUninitializedStringNamePtr)(snClassNamePtr),
	)
	if cok == 0 {
		log.Panic("failed to get class name",
			zap.Any("owner", engineObject),
		)
	}
	pnr.Pin(snClassNamePtr)
	// defer snClassName.Destroy()
	className := snClassName.ToUtf8()
	// const GDExtensionInstanceBindingCallbacks *binding_callbacks = nullptr;
	// Otherwise, try to look up the correct binding callbacks.
	cbs, ok := GDExtensionBindingGDExtensionInstanceBindingCallbacks.Get(className)
	if !ok {
		log.Warn("unable to find callbacks for Object")
		return nil
	}
	cbsPtr := &cbs
	pnr.Pin(cbsPtr)
	pnr.Pin(engineObject)
	pnr.Pin(FFI.Token)

	util.CgoTestCall(unsafe.Pointer(cbsPtr))
	util.CgoTestCall(unsafe.Pointer(engineObject))
	util.CgoTestCall(FFI.Token)
	instPtr = (*Object)(CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		(GDExtensionObjectPtr)(engineObject),
		FFI.Token,
		cbsPtr))
	runtime.KeepAlive(engineObject)
	runtime.KeepAlive(FFI.Token)
	runtime.KeepAlive(cbsPtr)
	if instPtr == nil || *instPtr == nil {
		log.Panic("unable to get instance")
		return nil
	}
	pnr.Pin(instPtr)
	wrapperClassName := (*instPtr).GetClassName()
	gdStrClassName := (*instPtr).GetClass()
	defer gdStrClassName.Destroy()
	log.Info("GetObjectInstanceBinding casted",
		zap.String("class", gdStrClassName.ToUtf8()),
		zap.String("className", wrapperClassName),
	)
	return *instPtr
}

func NewVariantGoString(v string) Variant {
	gdStr := NewStringWithUtf8Chars(v)
	defer gdStr.Destroy()
	ret := Variant{}
	ptr := ret.NativePtr()
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromString(gdStr, (GDExtensionUninitializedVariantPtr)(ptr))
	return ret
}

func (c *Variant) ToGoString() string {
	gdStr := c.ToString()
	defer gdStr.Destroy()
	return gdStr.ToUtf8()
}

func (c *Variant) NativeConstPtr() GDExtensionConstVariantPtr {
	return (GDExtensionConstVariantPtr)(c)
}

func (c *Variant) NativePtr() GDExtensionVariantPtr {
	return (GDExtensionVariantPtr)(c)
}

func (c *Variant) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return (GDExtensionConstTypePtr)(c)
}

func (c *Variant) AsGDExtensionTypePtr() GDExtensionTypePtr {
	return (GDExtensionTypePtr)(c)
}

var (
	needsDeinit = [GDEXTENSION_VARIANT_TYPE_VARIANT_MAX]bool{
		false, //NIL,
		false, //BOOL,
		false, //INT,
		false, //FLOAT,
		true,  //STRING,
		false, //VECTOR2,
		false, //VECTOR2I,
		false, //RECT2,
		false, //RECT2I,
		false, //VECTOR3,
		false, //VECTOR3I,
		true,  //TRANSFORM2D,
		false, //PLANE,
		false, //QUATERNION,
		true,  //AABB,
		true,  //BASIS,
		true,  //TRANSFORM,

		// misc types
		false, //COLOR,
		true,  //STRING_NAME,
		true,  //NODE_PATH,
		false, //RID,
		true,  //OBJECT,
		true,  //CALLABLE,
		true,  //SIGNAL,
		true,  //DICTIONARY,
		true,  //ARRAY,

		// typed arrays
		true, //PACKED_BYTE_ARRAY,
		true, //PACKED_INT32_ARRAY,
		true, //PACKED_INT64_ARRAY,
		true, //PACKED_FLOAT32_ARRAY,
		true, //PACKED_FLOAT64_ARRAY,
		true, //PACKED_STRING_ARRAY,
		true, //PACKED_VECTOR2_ARRAY,
		true, //PACKED_VECTOR3_ARRAY,
		true, //PACKED_COLOR_ARRAY,
	}
)

func (c *Variant) Call(
	method string,
	args []*Variant,
) (*Variant, error) {
	var (
		callArgs *GDExtensionConstVariantPtr
		r_ret    Variant
	)
	sn := NewStringNameWithLatin1Chars(method)
	defer sn.Destroy()
	callArgs = (*GDExtensionConstVariantPtr)(unsafe.Pointer(unsafe.SliceData(args)))

	callArgCount := len(args)
	var err GDExtensionCallError
	CallFunc_GDExtensionInterfaceVariantCall(
		(GDExtensionVariantPtr)(c.NativePtr()),
		(GDExtensionConstStringNamePtr)(sn.NativePtr()),
		callArgs,
		(GDExtensionInt)(callArgCount),
		(GDExtensionUninitializedVariantPtr)(r_ret.NativePtr()),
		&err,
	)
	if err.Ok() {
		return &r_ret, nil
	}
	return nil, err
}

func (c *Variant) CallStatic(
	vt GDExtensionVariantType,
	method string,
	args []*Variant,
	r_ret *Variant,
) error {
	var (
		callArgs *GDExtensionConstVariantPtr
	)
	sn := NewStringNameWithLatin1Chars(method)
	defer sn.Destroy()
	callArgs = (*GDExtensionConstVariantPtr)(unsafe.Pointer(unsafe.SliceData(args)))
	callArgCount := len(args)
	var err GDExtensionCallError
	CallFunc_GDExtensionInterfaceVariantCallStatic(
		vt,
		(GDExtensionConstStringNamePtr)(sn.NativePtr()),
		callArgs,
		(GDExtensionInt)(callArgCount),
		(GDExtensionUninitializedVariantPtr)(r_ret.NativePtr()),
		&err,
	)
	if err.Ok() {
		return nil
	}
	return err
}

func (c *Variant) GetType() GDExtensionVariantType {
	return CallFunc_GDExtensionInterfaceVariantGetType((GDExtensionConstVariantPtr)(c.NativePtr()))
}

func (c *Variant) Clear() {
	if needsDeinit[(int)(c.GetType())] {
		CallFunc_GDExtensionInterfaceVariantDestroy((GDExtensionVariantPtr)(c.NativePtr()))
	}
	CallFunc_GDExtensionInterfaceVariantNewNil((GDExtensionUninitializedVariantPtr)(c.NativePtr()))
}

var (
	ErrOutOfBounds = fmt.Errorf("out of bounds")
	ErrInvalid     = fmt.Errorf("invalid")
)

func (c *Variant) Set(key Variant, value Variant) error {
	var valid GDExtensionBool
	CallFunc_GDExtensionInterfaceVariantSet(
		c.NativePtr(),
		key.NativeConstPtr(), value.NativeConstPtr(), &valid)
	if valid != 0 {
		return ErrInvalid
	}
	return nil
}

func (c *Variant) SetNamed(name StringName, value Variant) error {
	var valid GDExtensionBool
	CallFunc_GDExtensionInterfaceVariantSetNamed(
		(GDExtensionVariantPtr)(unsafe.Pointer(c.NativePtr())),
		(GDExtensionConstStringNamePtr)(unsafe.Pointer(name.NativePtr())),
		(GDExtensionConstVariantPtr)(unsafe.Pointer(value.NativePtr())), &valid)
	if valid != 0 {
		return ErrInvalid
	}
	return nil
}

func (c *Variant) SetIndexed(index int, value Variant) error {
	var valid, oob GDExtensionBool
	CallFunc_GDExtensionInterfaceVariantSetIndexed(
		(GDExtensionVariantPtr)(unsafe.Pointer(c.NativePtr())),
		(GDExtensionInt)(index), value.NativeConstPtr(), &valid, &oob)
	if valid == 0 {
		return ErrInvalid
	}
	if oob != 0 {
		return ErrOutOfBounds
	}
	return nil
}

func (c *Variant) SetKeyed(key, value Variant) bool {
	var valid GDExtensionBool
	CallFunc_GDExtensionInterfaceVariantSetKeyed(
		(GDExtensionVariantPtr)(unsafe.Pointer(c.NativePtr())),
		key.NativeConstPtr(),
		value.NativeConstPtr(),
		&valid)
	return valid != 0
}

func (c *Variant) GetIndexed(index int) (Variant, error) {
	var (
		result Variant
		valid  GDExtensionBool
		oob    GDExtensionBool
	)
	CallFunc_GDExtensionInterfaceVariantGetIndexed(
		c.NativeConstPtr(), (GDExtensionInt)(index), (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(result.NativePtr())), &valid, &oob)
	if valid == 0 {
		return result, ErrInvalid
	}
	if oob != 0 {
		return result, ErrOutOfBounds
	}
	return result, nil
}

func (c *Variant) GetKeyed(key Variant) (Variant, error) {
	var (
		result Variant
		valid  GDExtensionBool
	)
	CallFunc_GDExtensionInterfaceVariantGetKeyed(
		c.NativeConstPtr(), key.NativeConstPtr(), (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(result.NativePtr())), &valid)
	if valid == 0 {
		return result, ErrInvalid
	}
	return result, nil
}

func (c *Variant) Destroy() {
	CallFunc_GDExtensionInterfaceVariantDestroy((GDExtensionVariantPtr)(c.NativePtr()))
}

func (c *Variant) Stringify() string {
	ret := NewString()
	defer ret.Destroy()
	CallFunc_GDExtensionInterfaceVariantStringify((GDExtensionConstVariantPtr)(c.NativePtr()), (GDExtensionStringPtr)(ret.NativePtr()))
	return ret.ToUtf8()
}

func (c *Variant) IsNil() bool {
	if c == nil {
		return true
	}
	for i := range c {
		if c[i] != 0 {
			return false
		}
	}
	return true
}

func Stringify(v Variant) string {
	return v.Stringify()
}

func VariantSliceToString(values []Variant) string {
	sb := strings.Builder{}
	sb.WriteString("[]Variant(")
	for i := range values {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(values[i].Stringify())
	}
	sb.WriteString(")")
	return sb.String()
}

func VariantPtrSliceToString(values []*Variant) string {
	sb := strings.Builder{}
	sb.WriteString("[]Variant(")
	for i := range values {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(values[i].Stringify())
	}
	sb.WriteString(")")
	return sb.String()
}

func ZapVariant(key string, v Variant) zap.Field {
	value := v.Stringify()
	return zap.String(key, value)
}

func ZapVector2(key string, v Vector2) zap.Field {
	variant := NewVariantVector2(v)
	// defer variant.Destroy()
	value := variant.Stringify()
	return zap.String(key, value)
}

func ZapVector3(key string, v Vector3) zap.Field {
	variant := NewVariantVector3(v)
	// defer variant.Destroy()
	value := variant.Stringify()
	return zap.String(key, value)
}

func ZapVector4(key string, v Vector4) zap.Field {
	variant := NewVariantVector4(v)
	// defer variant.Destroy()
	value := variant.Stringify()
	return zap.String(key, value)
}
