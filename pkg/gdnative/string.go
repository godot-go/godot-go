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
		log.Info("unable to parse utf-8", StringField("value", value))
	}
	return str
}

// NewStringNameFromGoString create a new string and fills it with the go string.
func NewStringNameFromGoString(value string) StringName {
	gds := NewStringFromGoString(value)
	defer gds.Destroy()
	return NewStringName(gds)
}

func (x *String) AsGoString() string {
	a := x.Ascii()
	defer a.Destroy()
	return a.GetData()
}

func (x *StringName) AsGoString() string {
	a := x.GetName()
	return a.AsGoString()
}
