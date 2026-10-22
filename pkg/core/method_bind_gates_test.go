package core

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
)

// bindGateFixture provides real reflect.Methods for the engine-free bind gates,
// which all run before any engine-dependent metadata setup.
type bindGateFixture struct{}

func (bindGateFixture) Variadic(a ...Variant) {}
func (bindGateFixture) OneReturn() int        { return 1 }
func (bindGateFixture) ThreeReturns() (int, int, int) {
	return 0, 0, 0
}
func (bindGateFixture) NonBoolSecond() (int, string) { return 0, "" }

func bindGateFatal(t *testing.T, methodName string, flags MethodFlags) (string, bool) {
	t.Helper()
	m, ok := reflect.TypeOf(&bindGateFixture{}).MethodByName(methodName)
	if !ok {
		t.Fatalf("%s method not found", methodName)
	}
	return recoverFatal(func() {
		NewGoMethodMetadata(m, "bindGateFixture", strings.ToLower(methodName), methodName, nil, nil, flags)
	})
}

// TestBindRejectsVariadicFlagMismatch pins the two directions of the vararg
// alignment gate: the Go signature and the bind flag must agree.
func TestBindRejectsVariadicFlagMismatch(t *testing.T) {
	t.Run("variadic method without the flag", func(t *testing.T) {
		msg, fired := bindGateFatal(t, "Variadic", METHOD_FLAGS_DEFAULT)
		if !fired {
			t.Fatal("a variadic Go method bound as ordinary must fail at bind time")
		}
		if !strings.Contains(msg, "not variadic aligned") {
			t.Fatalf("expected a variadic-alignment diagnostic, got: %s", msg)
		}
	})

	t.Run("vararg flag on a non-variadic method", func(t *testing.T) {
		msg, fired := bindGateFatal(t, "OneReturn", METHOD_FLAG_VARARG)
		if !fired {
			t.Fatal("a non-variadic Go method bound with the vararg flag must fail at bind time")
		}
		if !strings.Contains(msg, "not variadic aligned") {
			t.Fatalf("expected a variadic-alignment diagnostic, got: %s", msg)
		}
	})
}

// TestBindRejectsMoreThanTwoReturns pins the return-arity gate: the binding can
// deliver a value or a (value, handled) pair, never more.
func TestBindRejectsMoreThanTwoReturns(t *testing.T) {
	msg, fired := bindGateFatal(t, "ThreeReturns", METHOD_FLAGS_DEFAULT)
	if !fired {
		t.Fatal("more than two returns must fail at bind time")
	}
	if !strings.Contains(msg, "more than 1 type") {
		t.Fatalf("expected a return-arity diagnostic, got: %s", msg)
	}
}

// TestBindRejectsNonBoolSecondReturn pins the shape of the handled flag: the
// second return value exists to report whether the call was handled, so it must
// be a bool.
func TestBindRejectsNonBoolSecondReturn(t *testing.T) {
	msg, fired := bindGateFatal(t, "NonBoolSecond", METHOD_FLAGS_DEFAULT)
	if !fired {
		t.Fatal("a non-bool second return must fail at bind time")
	}
	if !strings.Contains(msg, "2nd return value must be of type bool") {
		t.Fatalf("expected a second-return-type diagnostic, got: %s", msg)
	}
}
