package core

import (
	"fmt"
	"reflect"

	. "github.com/godot-go/godot-go/pkg/builtin"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// toStringGdName is the Godot virtual name a class binds to override how it is
// rendered by to_string().
const toStringGdName = "to_string"

// defaultToStringFormat is the non-empty, identifying string supplied when a
// class registers no to_string virtual. The shape is the one the binding has
// always produced, kept so existing expectations hold.
func defaultToStringFormat(className string, instanceId uint64) string {
	return fmt.Sprintf("[ GDExtension::%s <--> Instance ID:%d ]", className, instanceId)
}

// resolveToString answers Godot's to_string request for inst.
func resolveToString(inst Object) string {
	return resolveToStringFor(inst, inst.GetClassName(), inst.GetInstanceId())
}

// resolveToStringFor applies the to_string dispatch rules to an already-read
// class identity: the registered to_string virtual's returned Go string when one
// is bound, and the default format otherwise.
//
// Everything else — class not registered, no virtual bound, or a virtual that
// does not return exactly one Go string — falls back to the default format. The
// fallback rather than a panic is deliberate: to_string is display-only, and a
// mis-bound override must not take the process down.
//
// Split out from resolveToString so the dispatch rules can be exercised without a
// live engine object behind recv.
func resolveToStringFor(recv any, className string, instanceId uint64) string {
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
	if !ok {
		return defaultToStringFormat(className, instanceId)
	}

	mcmi, ok := ci.VirtualMethodMap[toStringGdName]
	if !ok {
		return defaultToStringFormat(className, instanceId)
	}

	ret := mcmi.GoMethodMetadata.Func.Call([]reflect.Value{reflect.ValueOf(recv)})
	if len(ret) != 1 || ret[0].Kind() != reflect.String {
		log.Warn("to_string virtual did not return a single Go string; using default format",
			zap.String("class", className),
			zap.String("go_method", mcmi.GoMethodMetadata.GoMethodName),
			zap.Int("return_values", len(ret)),
		)
		return defaultToStringFormat(className, instanceId)
	}
	return ret[0].String()
}
