package core

import (
	"errors"
	"reflect"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"go.uber.org/zap/zapcore"
)

// Lazy dispatch payloads (trim-dispatch-overhead D3).
//
// The dispatch entries ("Call", "Call Variadic", "Ptrcall", "Calling", ...)
// used to describe their call through zap.String fields whose values were
// function calls: formatting the bind, stringifying every argument Variant
// through the engine, dumping return values. Go evaluates those arguments
// before zap ever compares the entry level against the threshold, so an entry
// the threshold discarded still paid for the whole description — including
// the engine crossings Variant.Stringify makes (measured before this change:
// 8 crossings for a discarded two-argument fallback varcall, 20 for five
// arguments; see the change evidence).
//
// The payload types below move that work behind the encoder interface: zap
// invokes MarshalLogObject / String only while writing an entry that passed
// the threshold. A discarded entry now costs the threshold comparison and its
// own message, and the engine is never asked to stringify a value belonging to
// a call nobody reads.
//
// Field names are unchanged (bind, gd_args, resolved_args, ret); the content
// each field carries is the same identification as before, structured where
// the value is composed (bind, the variant collections) and text-identical
// where it was already text (ret, spew dumps). The one deliberate removal is
// the variadic entry's resolved_args field: that dispatch hands the collected
// Variants to the handler unchanged, so its resolved_args was always a
// second stringification of the identical gd_args collection.

// lazyText defers a description to write time. The function runs at most
// once per entry actually written.
type lazyText func() string

// String implements fmt.Stringer for zap.Stringer.
func (l lazyText) String() string { return l() }

// variantCollection describes the Variants handed to a dispatch. Each
// element's description crosses into the engine (variant_stringify), which is
// exactly why the description may not happen for a discarded entry.
type variantCollection []Variant

// MarshalLogObject implements zapcore.ObjectMarshaler: one description per
// element, in call order.
func (p variantCollection) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	return enc.AddArray("values", variantCollectionArray(p))
}

// variantCollectionArray stringifies the elements at encode time.
type variantCollectionArray []Variant

// MarshalLogArray implements zapcore.ArrayMarshaler.
func (a variantCollectionArray) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for i := range a {
		enc.AppendString(a[i].Stringify())
	}
	return nil
}

// resolvedCollection describes the argument list after the trailing-default
// fill. When the fill changed nothing (the engine pads every default before
// dispatching, so the full-supply case is the norm), re-stringifying the
// collection would describe the same values a second time inside one entry;
// the payload then records the identity instead, keeping "each value is
// described at most once per entry" true for the written form too. The
// comparison runs at encode time only — a discarded entry never pays it.
type resolvedCollection struct {
	resolved []Variant
	supplied []Variant
}

// MarshalLogObject implements zapcore.ObjectMarshaler.
func (p resolvedCollection) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if sameVariantCells(p.resolved, p.supplied) {
		enc.AddString("resolution", "same as gd_args")
		return nil
	}
	return variantCollection(p.resolved).MarshalLogObject(enc)
}

// sameVariantCells reports whether both collections hold byte-identical
// Variant cells. A filled slot is engine-supplied default storage, distinct
// from any supplied cell it replaced only when the fill actually ran.
func sameVariantCells(a, b []Variant) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler for the bind field:
// the same identification of the call that String() produced as one joined
// text — class, Go method, per-argument Go type and Godot variant type, and
// the return type — as structured fields instead.
func (md *GoMethodMetadata) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if md == nil {
		// String() treats a null metadata as fatal; the marshaler runs inside
		// the encoder, where an error field is the proportionate signal (a
		// dispatch site always holds a non-nil receiver, so this guards
		// against a future caller threading a nil through).
		return errors.New("GoMethodMetadata cannot be null")
	}
	enc.AddString("class", md.ClassName)
	enc.AddString("method", md.GoMethodName)
	enc.AddArray("argument_types", goArgumentTypeNames{types: md.GoArgumentTypes, md: md})
	if md.GoReturnType != nil {
		enc.AddString("return_type", md.GoReturnType.Name())
	}
	return nil
}

// goArgumentTypeNames writes the paired Go-type / Godot-variant-type
// descriptions ("int64/INT") the old String() built.
type goArgumentTypeNames struct {
	types []reflect.Type
	md    *GoMethodMetadata
}

// MarshalLogArray implements zapcore.ArrayMarshaler.
func (a goArgumentTypeNames) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for i := range a.types {
		variantType := ""
		if i < len(a.md.gdeArgumentTypes) {
			variantType = GDExtensionVariantTypeStringMap[a.md.gdeArgumentTypes[i]]
		}
		enc.AppendString(a.types[i].Name() + "/" + variantType)
	}
	return nil
}
