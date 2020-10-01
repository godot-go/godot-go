package gdnative

/*
#include <gdnative.wrapper.gen.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"github.com/godot-go/godot-go/pkg/log"
)

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

func NewVariantGoString(value string) Variant {
	gs := NewStringFromGoString(value)
	defer gs.Destroy()

	return NewVariantString(gs)
}
