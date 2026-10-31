package gdclassimpl

/*
	A-before/A-after benchmark for isNilObjectWrapper (trim-dispatch-overhead
	task 3.3). isNilObjectWrapperReflectOld is the pre-switch reflect form,
	kept as a same-run control so the mechanism delta is measured on one
	machine. The fake wrappers come from object_return_test.go (same
	package); both shapes answer through declared methods, so nothing here
	touches the engine. Run:

		go test ./pkg/gdclassimpl/ -run XXXnoTests -bench 'IsNilObjectWrapper' -benchmem
*/

import (
	"reflect"
	"testing"

	"github.com/godot-go/godot-go/pkg/builtin"
)

func isNilObjectWrapperReflectOld(obj builtin.Object) bool {
	if obj == nil {
		return true
	}
	v := reflect.ValueOf(obj)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

func BenchmarkIsNilObjectWrapperTypedNil(b *testing.B) {
	var obj builtin.Object = (*fakeLabelImpl)(nil)
	for b.Loop() {
		if !isNilObjectWrapper(obj) {
			b.Fatal("typed nil must report nil")
		}
	}
}

func BenchmarkIsNilObjectWrapperTypedNilReflectOld(b *testing.B) {
	var obj builtin.Object = (*fakeLabelImpl)(nil)
	for b.Loop() {
		if !isNilObjectWrapperReflectOld(obj) {
			b.Fatal("typed nil must report nil")
		}
	}
}

func BenchmarkIsNilObjectWrapperLive(b *testing.B) {
	obj := &fakeLabelImpl{}
	for b.Loop() {
		if isNilObjectWrapper(obj) {
			b.Fatal("live wrapper must not report nil")
		}
	}
}

func BenchmarkIsNilObjectWrapperLiveReflectOld(b *testing.B) {
	obj := &fakeLabelImpl{}
	for b.Loop() {
		if isNilObjectWrapperReflectOld(obj) {
			b.Fatal("live wrapper must not report nil")
		}
	}
}
