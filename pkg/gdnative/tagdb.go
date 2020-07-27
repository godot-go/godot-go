package gdnative

/*
#include <nativescript.wrappergen.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	"hash/fnv"
	"unsafe"

	"github.com/pcting/godot-go/pkg/log"
)

const (
	EmptyTypeTag   TypeTag   = TypeTag(0)
	EmptyMethodTag MethodTag = MethodTag(0)
)

type classMethod struct {
	className  string
	methodName string
}

func (c classMethod) String() string {
	return c.className + "::" + c.methodName
}

type TypeTag uint

type MethodTag uint

type tagDB struct {
	parentTo   map[TypeTag]TypeTag
	typeTags   map[TypeTag]string
	methodTags map[MethodTag]classMethod
}

type TagDBStats struct {
	ParentCount    int
	TypeTagCount   int
	MethodTagCount int
}

func newTypeTagFromString(className string) TypeTag {
	h := fnv.New32a()
	h.Write([]byte(className))
	return TypeTag(uint(h.Sum32()))
}

func newMethodTagFromString(cm classMethod) MethodTag {
	h := fnv.New32a()
	h.Write([]byte(cm.String()))
	return MethodTag(uint(h.Sum32()))
}

func (t tagDB) Stats() TagDBStats {
	return TagDBStats{
		ParentCount:    len(t.parentTo),
		TypeTagCount:   len(t.typeTags),
		MethodTagCount: len(t.methodTags),
	}
}

func (t tagDB) RegisterType(typeTag, baseTypeTag string) (TypeTag, TypeTag) {
	tt := newTypeTagFromString(typeTag)
	btt := newTypeTagFromString(baseTypeTag)

	if tt == btt {
		log.WithField(
			"tagType", typeTag,
		).WithField(
			"baseTypeTag", baseTypeTag,
		).Panic("hash collision with tag and base tag")
	}

	if existing, ok := t.typeTags[tt]; ok {
		log.WithField(
			"new", typeTag,
		).WithField(
			"existing", existing,
		).Panic("hash collision with new tag and existing tag")
	}

	t.parentTo[tt] = btt
	t.typeTags[tt] = typeTag

	return tt, btt
}

func (t tagDB) RegisterMethod(className, methodName string) MethodTag {
	cm := classMethod{
		className:  className,
		methodName: methodName,
	}
	mt := newMethodTagFromString(cm)

	if existing, ok := t.methodTags[mt]; ok {
		log.WithField(
			"className", className,
		).WithField(
			"methodName", methodName,
		).WithField(
			"existing", fmt.Sprintf("%+v", existing),
		).Panic("hash collision with new and existing method tag")
	}

	t.methodTags[mt] = cm

	return mt
}

func (t tagDB) IsTypeKnown(typeTag TypeTag) bool {
	_, ok := t.parentTo[typeTag]
	return ok
}

func (t tagDB) GetRegisteredClassName(typeTag TypeTag) string {
	if tt, ok := t.typeTags[typeTag]; ok {
		return tt
	}

	// log.Panic("unable to find type tag")

	return ""
}

func (t tagDB) IsMethodKnown(methodTag MethodTag) bool {
	_, ok := t.methodTags[methodTag]
	return ok
}

func (t tagDB) GetRegisteredMethodName(methodTag MethodTag) string {
	if cm, ok := t.methodTags[methodTag]; ok {
		return cm.String()
	}

	// log.Panic("unable to find method tag")

	return ""
}

func (t tagDB) RegisterGlobalType(name string, typeTag, baseTypeTag string) (TypeTag, TypeTag) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	tt, btt := t.RegisterType(typeTag, baseTypeTag)

	C.go_godot_nativescript_set_global_type_tag(Nativescript11Api, RegisterState.LanguageIndex, cname, unsafe.Pointer(uintptr(tt)))

	return tt, btt
}

func (t tagDB) IsTypeCompatible(askTag, haveTag TypeTag) bool {
	if haveTag == EmptyTypeTag {
		return false
	}

	// traverse up the hierarchy until matched
	for tag := haveTag; tag != EmptyTypeTag; tag = t.parentTo[tag] {
		if tag == askTag {
			return true
		}
	}

	return false
}

//export get_class_name_from_type_tag
func get_class_name_from_type_tag(tt C.long) *C.char {
	name := RegisterState.TagDB.GetRegisteredClassName(TypeTag(tt))
	return C.CString(name)
}

//export get_method_name_from_method_tag
func get_method_name_from_method_tag(mt C.long) *C.char {
	name := RegisterState.TagDB.GetRegisteredMethodName(MethodTag(mt))
	return C.CString(name)
}
