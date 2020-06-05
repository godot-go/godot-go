package gdnative

/*
#include <gdnative.gen.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"github.com/pcting/godot-go/pkg/log"
)

// // StringName as declared in gdnative/string_name.h:47
// type StringName C.godot_string_name

// // String as declared in gdnative/string.h:50
// type String C.godot_string

// // CharString as declared in gdnative/string.h:58
// type CharString C.godot_char_string

// TODO: implement wchar String constructors

// func ToPChar(value string) *C.char {
// 	ret, _ := unpackPCharString(value)

// 	return ret
// }

// func PCharAppend(base *C.char, postfix string) *C.char {
// 	return ToPChar(packPCharString(base) + postfix)
// }

// NewStringFromGoString create a new string and fills it with the go string.
func NewStringFromGoString(value string) String {
	str := NewString()

	// https://github.com/godotengine/godot/blob/ef5891091bceef2800b4fae4cd85af219e791467/core/ustring.h#L300
	// return true on error
	if str.ParseUtf8(value) {
		log.WithField("value", value).Info("unable to parse '%s' as utf-8", value)
	}
	return str
}

func (x *String) AsGoString() string {
	a := x.Ascii()
	defer a.Destroy()
	return a.GetData()
}

// // ToCharArray must free
// func (gdt *String) ToCharArray() *C.char {
// 	gschar := C.go_godot_string_ascii(CoreApi, (*C.godot_string)(unsafe.Pointer(gdt)))

// 	return (*C.char)(unsafe.Pointer(C.go_godot_char_string_get_data(CoreApi, &gschar)))
// }
