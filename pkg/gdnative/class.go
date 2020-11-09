package gdnative

/*
#include <nativescript.wrapper.gen.h>
#include <cgo_gateway_register_class.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"math/rand"
	"reflect"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
)

// UserDataIdentifiable returns data for gdnative class registration.
type UserDataIdentifiable interface {
	GetUserData() UserData
	setUserDataFromTypeTag(tt TypeTag)
}

// UserDataIdentifiableImpl is the default implementation of UserDataIdentifiable.
type UserDataIdentifiableImpl struct {
	UserData UserData
}

// GetUserData returns the unique identifier for the instance.
func (u *UserDataIdentifiableImpl) GetUserData() UserData {
	return u.UserData
}

func (u *UserDataIdentifiableImpl) setUserDataFromTypeTag(_ TypeTag) {
	// TODO: provide better method to reduce collisiosn; maybe
	//       an auto-increment method?
	u.UserData = UserData(rand.Int())
}

// NativeScriptClass represents the base interface for creating
// custom Nativescript classes.
type NativeScriptClass interface {
	Class
	UserDataIdentifiable
	Init()
	OnClassRegistered(e ClassRegisteredEvent)
}

// Class represents gdnative's
type Class interface {
	Wrapped
	ClassName() string
	BaseClass() string
	Destroy()
}

// ClassRegisteredEvent contains the event context for when Godot has
// registered the specified class.
type ClassRegisteredEvent struct {
	ClassName       string
	ClassType       reflect.Type
	ClassTypeTag    TypeTag
	BaseName        string
	BaseTypeTag     TypeTag
	_pCharClassName *C.char
	_pCharBaseName  *C.char
}

func (e *ClassRegisteredEvent) destroy() {
	if e._pCharClassName == nil || e._pCharBaseName == nil {
		log.Warn("ClassRegisteredEvent.destroy called more than once")
		return
	}

	C.free(unsafe.Pointer(e._pCharClassName))
	C.free(unsafe.Pointer(e._pCharBaseName))

	e._pCharClassName = nil
	e._pCharBaseName = nil
}

func newClassRegisteredEvent(
	className string,
	classType reflect.Type,
	classTypeTag TypeTag,
	baseName string,
	baseTypeTag TypeTag,
) ClassRegisteredEvent {
	_pCharClassName := C.CString(className)
	_pCharBaseName := C.CString(baseName)

	e := ClassRegisteredEvent{
		ClassName:       className,
		ClassType:       classType,
		ClassTypeTag:    classTypeTag,
		BaseName:        baseName,
		BaseTypeTag:     baseTypeTag,
		_pCharClassName: _pCharClassName,
		_pCharBaseName:  _pCharBaseName,
	}

	return e
}

// CreateClassFunc internal version of CreateNativeScriptClassFunc.
type CreateClassFunc func(owner *GodotObject, typeTag TypeTag) Class

// UserData must be unique to the instance.
type UserData int

// UserDataMap is an identity map key on by NativeScriptClass.GetUserData()
type UserDataMap map[UserData]NativeScriptClass

// CreateFunc is a callback for Godot to call whenever a NativeScriptClass is
// requested to be instantiated.
type CreateFunc func(*GodotObject, MethodData) UserData

// DestroyFunc is a callback for Godot to call whenever a NativeScriptClass is
// destroyed.
type DestroyFunc func(*GodotObject, MethodData, UserData)

// FreeFunc is a callback to clean up CreateFunc and DestroyFunc.
type FreeFunc func(MethodData)
