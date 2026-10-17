package core

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
)

// recoverFatal runs fn and reports whether it died through a fatal log
// (log.Panic), returning the formatted panic payload.
func recoverFatal(fn func()) (msg string, fired bool) {
	defer func() {
		if r := recover(); r != nil {
			msg, fired = fmt.Sprint(r), true
		}
	}()
	fn()
	return "", false
}

func staticFixtureAdd(a int64, b float64) int64 {
	return a + int64(b)
}

func TestGoMethodMetadataIsStatic(t *testing.T) {
	cases := []struct {
		name  string
		flags MethodFlags
		want  bool
	}{
		{"default flags are not static", METHOD_FLAGS_DEFAULT, false},
		{"static flag is static", METHOD_FLAG_STATIC, true},
		{"static combined with vararg is static", METHOD_FLAG_STATIC | METHOD_FLAG_VARARG, true},
		{"virtual alone is not static", METHOD_FLAG_VIRTUAL, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			md := &GoMethodMetadata{MethodFlags: tc.flags}
			if got := md.IsStatic(); got != tc.want {
				t.Fatalf("IsStatic() with flags %v = %v, want %v", tc.flags, got, tc.want)
			}
		})
	}
}

func TestNewGoMethodMetadataStaticRejectsNonCallable(t *testing.T) {
	msg, fired := recoverFatal(func() {
		newGoMethodMetadataStatic(42, "Example", "answer", nil, nil, METHOD_FLAG_STATIC)
	})
	if !fired {
		t.Fatal("non-callable fn must fail fatally at bind time")
	}
	if !strings.Contains(msg, "Example") || !strings.Contains(msg, "answer") {
		t.Fatalf("fatal message must name class and method, got: %s", msg)
	}
}

func TestNewGoMethodMetadataStaticRejectsVirtualFlagCombination(t *testing.T) {
	_, fired := recoverFatal(func() {
		newGoMethodMetadataStatic(
			staticFixtureAdd, "Example", "add_nums",
			[]string{"a", "b"}, nil, METHOD_FLAG_STATIC|METHOD_FLAG_VIRTUAL,
		)
	})
	if !fired {
		t.Fatal("static binds must never carry the virtual flag")
	}
}

func TestNewGoMethodMetadataStaticArityValidation(t *testing.T) {
	t.Run("excess argument names", func(t *testing.T) {
		_, fired := recoverFatal(func() {
			newGoMethodMetadataStatic(
				staticFixtureAdd, "Example", "add_nums",
				[]string{"a", "b", "c"}, nil, METHOD_FLAG_STATIC,
			)
		})
		if !fired {
			t.Fatal("more argument names than parameters must fail fatally")
		}
	})
	t.Run("excess default arguments", func(t *testing.T) {
		_, fired := recoverFatal(func() {
			newGoMethodMetadataStatic(
				staticFixtureAdd, "Example", "add_nums",
				[]string{"a", "b"}, []Variant{{}, {}, {}}, METHOD_FLAG_STATIC,
			)
		})
		if !fired {
			t.Fatal("more defaults than parameters must fail fatally")
		}
	})
	t.Run("variadic flag mismatch", func(t *testing.T) {
		_, fired := recoverFatal(func() {
			newGoMethodMetadataStatic(
				staticFixtureAdd, "Example", "add_nums",
				[]string{"a", "b"}, nil, METHOD_FLAG_STATIC|METHOD_FLAG_VARARG,
			)
		})
		if !fired {
			t.Fatal("a non-variadic function bound with the vararg flag must fail fatally")
		}
	})
}

// TestValidateBindSignatureArgumentBaseOffset pins that arity is counted as
// NumIn minus the argument base: the static path (base 0) sees every
// parameter of a two-parameter package function, while the instance path
// (base 1) treats In(0) as the receiver. Running through the pure helper
// keeps the assertion engine-free.
func TestValidateBindSignatureArgumentBaseOffset(t *testing.T) {
	mt := reflect.TypeOf(staticFixtureAdd)
	t.Run("static base accepts both parameters", func(t *testing.T) {
		_, fired := recoverFatal(func() {
			validateBindSignature(mt, 0, []string{"a", "b"}, nil, METHOD_FLAG_STATIC, "Example", "add_nums")
		})
		if fired {
			t.Fatalf("two named arguments on a two-parameter function must pass, got: %s", mt)
		}
	})
	t.Run("instance base rejects the same shape", func(t *testing.T) {
		_, fired := recoverFatal(func() {
			validateBindSignature(mt, 1, []string{"a", "b"}, nil, METHOD_FLAGS_DEFAULT, "Example", "add_nums")
		})
		if !fired {
			t.Fatal("with a receiver, a two-parameter function has one argument; two names must fail fatally")
		}
	})
	t.Run("static rejects the virtual flag combination", func(t *testing.T) {
		_, fired := recoverFatal(func() {
			validateBindSignature(mt, 0, []string{"a", "b"}, nil, METHOD_FLAG_STATIC|METHOD_FLAG_VIRTUAL, "Example", "add_nums")
		})
		if !fired {
			t.Fatal("static binds must never carry the virtual flag")
		}
	})
	t.Run("instance base keeps virtual allowed", func(t *testing.T) {
		m, ok := reflect.TypeOf(&staticRecvFixture{}).MethodByName("V_Node_Ready")
		if !ok {
			t.Fatal("fixture method missing")
		}
		_, fired := recoverFatal(func() {
			validateBindSignature(m.Type, 1, nil, nil, METHOD_FLAG_VIRTUAL, "Example", "_ready")
		})
		if fired {
			t.Fatal("virtual must remain a legal flag on the instance path")
		}
	})
}

type staticRecvFixture struct{}

func (s *staticRecvFixture) V_Node_Ready() {}

// TestReflectFuncCallArgsStaticOmitsReceiver pins that the argument builders
// allocate a receiver slot only for instance binds: a static dispatch must
// hand md.Func exactly the decoded parameters, with no leading value, while
// the instance shape keeps the receiver at index 0.
func TestReflectFuncCallArgsStaticOmitsReceiver(t *testing.T) {
	t.Run("varcall static", func(t *testing.T) {
		args, err := reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs(nil, nil, nil, false)
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if len(args) != 0 {
			t.Fatalf("static varcall dispatch must not allocate a receiver slot, got %d values", len(args))
		}
	})
	t.Run("varcall instance keeps receiver", func(t *testing.T) {
		args, err := reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs(nil, nil, nil, true)
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if len(args) != 1 {
			t.Fatalf("instance varcall dispatch must carry the receiver at index 0, got %d values", len(args))
		}
	})
	t.Run("ptrcall static", func(t *testing.T) {
		args := reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs(nil, nil, nil, false)
		if len(args) != 0 {
			t.Fatalf("static ptrcall dispatch must not allocate a receiver slot, got %d values", len(args))
		}
	})
	t.Run("ptrcall instance keeps receiver", func(t *testing.T) {
		args := reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs(nil, nil, nil, true)
		if len(args) != 1 {
			t.Fatalf("instance ptrcall dispatch must carry the receiver at index 0, got %d values", len(args))
		}
	})
}

// TestIsPtrcallBorrowEchoCandidateOffset pins that borrow-echo scanning skips
// the receiver slot only when the argument slice carries one: the first real
// argument of a static bind must remain an echo candidate.
func TestIsPtrcallBorrowEchoCandidateOffset(t *testing.T) {
	borrow := StringName{}
	echo := StringName{}
	ret := reflect.ValueOf(echo)
	t.Run("static first argument is a candidate", func(t *testing.T) {
		args := []reflect.Value{reflect.ValueOf(borrow)}
		if !isPtrcallBorrowEcho(ret, args, false) {
			t.Fatal("a static bind's first argument must be checked for a borrow echo")
		}
	})
	t.Run("instance receiver slot is skipped", func(t *testing.T) {
		args := []reflect.Value{reflect.ValueOf("receiver-not-a-string-name"), reflect.ValueOf(borrow)}
		if !isPtrcallBorrowEcho(ret, args, true) {
			t.Fatal("the instance-path scan must find the echo at index 1")
		}
	})
	t.Run("slot zero skipped when a receiver is present", func(t *testing.T) {
		args := []reflect.Value{reflect.ValueOf(borrow)}
		if isPtrcallBorrowEcho(ret, args, true) {
			t.Fatal("with a receiver slot, index 0 is the receiver and must not be an echo candidate")
		}
	})
}
