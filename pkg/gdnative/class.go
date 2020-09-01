package gdnative

/*
#include <nativescript_wrappergen.h>
#include <cgo_gateway_register_class.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"math/rand"
	"reflect"
	"unsafe"
)

type UserDataIdentifiable interface {
	GetUserData() UserData
	generateUserData(tt TypeTag)
}

type UserDataIdentifiableImpl struct {
	UserData UserData
}

func (u *UserDataIdentifiableImpl) GetUserData() UserData {
	return u.UserData
}

func (u *UserDataIdentifiableImpl) generateUserData(tt TypeTag) {
	// TODO: provide better method to reduce collision; maybe
	//       an auto-increment method?
	u.UserData = UserData(rand.Int())
}

type NativeScriptClass interface {
	Class
	UserDataIdentifiable
	Init()
	OnClassRegistered(e ClassRegisteredEvent)
}

type Class interface {
	GetOwnerObject() *GodotObject
	GetTypeTag() TypeTag
	ClassName() string
	BaseClass() string
	Free()
}

type ClassRegisteredEvent struct {
	ClassName       string
	ClassType       reflect.Type
	ClassTypeTag    TypeTag
	BaseName        string
	BaseTypeTag     TypeTag
	_pCharClassName *C.char
	_pCharBaseName  *C.char
}

func (e *ClassRegisteredEvent) Destroy() {
	C.free(unsafe.Pointer(e._pCharClassName))
	C.free(unsafe.Pointer(e._pCharBaseName))
}

func NewClassRegisteredEvent(
	className string,
	classType reflect.Type,
	classTypeTag TypeTag,
	baseName string,
	baseTypeTag TypeTag,
) ClassRegisteredEvent {
	_pCharClassName := C.CString(className)
	_pCharBaseName := C.CString(baseName)

	return ClassRegisteredEvent{
		ClassName:       className,
		ClassType:       classType,
		ClassTypeTag:    classTypeTag,
		BaseName:        baseName,
		BaseTypeTag:     baseTypeTag,
		_pCharClassName: _pCharClassName,
		_pCharBaseName:  _pCharBaseName,
	}
}

// CreateClassFunc internal version of CreateNativeScriptClassFunc
type CreateClassFunc func(owner *GodotObject, typeTag TypeTag) Class

// UserData must be unique to the instance
type UserData int
type UserDataMap map[UserData]NativeScriptClass

type CreateFunc func(*GodotObject, MethodData) UserData
type DestroyFunc func(*GodotObject, MethodData, UserData)
type FreeFunc func(MethodData)
