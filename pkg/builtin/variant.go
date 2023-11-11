package builtin

// #include <godot/gdextension_interface.h>
import "C"
import (
	"fmt"
	"strings"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
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
	CallFunc_GDExtensionInterfaceVariantNewCopy((GDExtensionUninitializedVariantPtr)(ret.NativePtr()), NativeConstPtr)
	return ret
}

func NewVariantCopy(dst, src Variant) {
	CallFunc_GDExtensionInterfaceVariantNewCopy((GDExtensionUninitializedVariantPtr)(dst.NativePtr()), src.NativeConstPtr())
}

func NewVariantNil() Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	GDExtensionVariantPtrWithNil(ptr)
	return ret
}

func GDExtensionVariantPtrWithNil(rOut GDExtensionUninitializedVariantPtr) {
	CallFunc_GDExtensionInterfaceVariantNewNil(rOut)
}

func NewVariantCopyWithGDExtensionConstVariantPtr(ptr GDExtensionConstVariantPtr) Variant {
	ret := Variant{}
	typedSrc := (*[VariantSize]uint8)(ptr)

	for i := 0; i < VariantSize; i++ {
		ret[i] = typedSrc[i]
	}
	return ret
}

func NewVariantGodotObject(owner *GodotObject) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromGodotObjectPtr(owner, (GDExtensionUninitializedVariantPtr)(ret.NativePtr()))
	return ret
}

func GDExtensionVariantPtrFromGodotObjectPtr(owner *GodotObject, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(unsafe.Pointer(&owner)),
	)
}

func (c *Variant) ToObject() Object {
	if c.IsNil() {
		return nil
	}
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	var engineObject *GodotObject
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(unsafe.Pointer(&engineObject)),
		c.NativePtr(),
	)
	ret := GetObjectInstanceBinding(engineObject)
	return ret
}

func NewVariantGoString(v string) Variant {
	gdStr := NewStringWithUtf8Chars(v)
	defer gdStr.Destroy()
	ret := Variant{}
	GDExtensionVariantPtrFromString(gdStr, (GDExtensionUninitializedVariantPtr)(ret.NativePtr()))
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
	defer variant.Destroy()
	value := variant.Stringify()
	return zap.String(key, value)
}

func ZapVector3(key string, v Vector3) zap.Field {
	variant := NewVariantVector3(v)
	defer variant.Destroy()
	value := variant.Stringify()
	return zap.String(key, value)
}

func ZapVector4(key string, v Vector4) zap.Field {
	variant := NewVariantVector4(v)
	defer variant.Destroy()
	value := variant.Stringify()
	return zap.String(key, value)
}
