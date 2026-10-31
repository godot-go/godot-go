package core

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// trim-dispatch-overhead task 2.1 — the lazy-payload witnesses.
//
// The capability claim: work performed to describe an entry happens only
// when the threshold will write that entry. These tests prove it at two
// levels: a capturing write syncer around a zap core (the mechanism every
// dispatch site now uses), and the reflective Ptrcall dispatch site driven
// end-to-end engine-free — suppressed writes nothing, raised writes the same
// field names and content the eager form carried.
//
// Engine-free reach note: marshaling a non-empty variantCollection asks the
// engine to stringify each element, so the engine-free tier exercises the
// empty collection and the resolved-vs-supplied dedup (which by design
// performs zero stringifications); the non-empty content assertions live in
// the in-engine dispatch_diag fixture. That reach split is itself a proof
// direction: a regression that reintroduces eager payloads faults here on
// the engine-free tests, where any Variant stringification would dereference
// the unloaded interface table.

// capturingSyncer collects everything a zap core writes.
type capturingSyncer struct {
	buf bytes.Buffer
}

func (c *capturingSyncer) Write(p []byte) (int, error) { return c.buf.Write(p) }
func (c *capturingSyncer) Sync() error                 { return nil }

func (c *capturingSyncer) Lines() string { return c.buf.String() }

// newCapturingLogger builds a console-encoder logger (the production encoder
// shape from pkg/log) over a capturing syncer at the given level.
func newCapturingLogger(level zapcore.Level) (*zap.Logger, *capturingSyncer, zap.AtomicLevel) {
	sink := &capturingSyncer{}
	atomicLevel := zap.NewAtomicLevelAt(level)
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = ""
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), sink, atomicLevel)
	return zap.New(core), sink, atomicLevel
}

// countingMarshaler records every time the encoder asks it to marshal.
type countingMarshaler struct{ calls int }

func (c *countingMarshaler) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	c.calls++
	enc.AddString("marker", "counted")
	return nil
}

// TestLazyPayloadsNotMarshaledWhileSuppressed is the core falsification:
// entries below the threshold must not invoke a single marshaler or Stringer.
func TestLazyPayloadsNotMarshaledWhileSuppressed(t *testing.T) {
	logger, sink, level := newCapturingLogger(zapcore.WarnLevel)
	cm := &countingMarshaler{}
	textCalls := 0
	fields := []zap.Field{
		zap.Object("bind", cm),
		zap.Stringer("ret", lazyText(func() string {
			textCalls++
			return "described"
		})),
	}

	logger.Info("suppressed-entry", fields...)
	if got := sink.Lines(); got != "" {
		t.Errorf("suppressed entry wrote %q", got)
	}
	if cm.calls != 0 {
		t.Errorf("ObjectMarshaler invoked %d times while suppressed", cm.calls)
	}
	if textCalls != 0 {
		t.Errorf("Stringer invoked %d times while suppressed", textCalls)
	}

	// Raising the threshold writes the entry and marshals each payload once.
	level.SetLevel(zapcore.InfoLevel)
	logger.Info("written-entry", fields...)
	out := sink.Lines()
	if !strings.Contains(out, "written-entry") {
		t.Fatalf("raising the threshold did not write the entry: %q", out)
	}
	if cm.calls != 1 {
		t.Errorf("ObjectMarshaler invoked %d times for one written entry (want 1)", cm.calls)
	}
	if textCalls != 1 {
		t.Errorf("Stringer invoked %d times for one written entry (want 1)", textCalls)
	}
	if !strings.Contains(out, "counted") || !strings.Contains(out, "described") {
		t.Errorf("written entry missing payload content: %q", out)
	}
}

// newReflectiveGateMetadata is a name-string-style bind (Func set, no typed
// closure) over a plain Go func — the reflective fallback shape the dispatch
// sites log around.
func newReflectiveGateMetadata() *GoMethodMetadata {
	return &GoMethodMetadata{
		ClassName:        "SpotStub",
		GdMethodName:     "gate_triple",
		GoMethodName:     "GateTriple",
		Func:             reflect.ValueOf(func(s *spotStubClass, n int64) int64 { return n }),
		GoReturnType:     reflect.TypeFor[int64](),
		GoReturnStyle:    ValueReturnStyle,
		GoArgumentTypes:  []reflect.Type{reflect.TypeFor[int64]()},
		gdeArgumentTypes: []GDExtensionVariantType{GDEXTENSION_VARIANT_TYPE_INT},
	}
}

// TestPtrcallSiteSuppressedWritesNothingRaisedRestoresContent drives the
// real dispatch site through the global logger, redirecting entries into a
// capture buffer (the logger supports the swap since the stdoutSink fix).
func TestPtrcallSiteSuppressedWritesNothingRaisedRestoresContent(t *testing.T) {
	InitPrimativeTypeEncodersForUnitTests()

	md := newReflectiveGateMetadata()
	stub := &spotStubClass{}
	var cell int64 = 7
	args := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(&cell)}
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])

	var buf bytes.Buffer
	prev := log.SetWriteSyncer(&buf)
	defer log.SetWriteSyncer(prev)
	prevLevel := log.GetLevel()
	defer log.SetLevel(prevLevel)

	log.SetLevel(log.WarnLevel)
	buf.Reset()
	md.Ptrcall(stub, args, retPtr)
	if got := buf.String(); got != "" {
		t.Errorf("Ptrcall entry written while suppressed: %q", got)
	}

	log.SetLevel(log.InfoLevel)
	buf.Reset()
	md.Ptrcall(stub, args, retPtr)
	got := buf.String()
	if !strings.Contains(got, "Ptrcall") {
		t.Fatalf("raised threshold did not write the Ptrcall entry: %q", got)
	}
	// Same field names as the eager form...
	for _, key := range []string{"bind", "ret"} {
		if !strings.Contains(got, key) {
			t.Errorf("entry lost field %q: %q", key, got)
		}
	}
	// ...and the same identification content: class, method, the paired
	// Go/variant argument type, the return type. The console encoder renders
	// the nested bind object as JSON, so the flattened probe keeps the
	// quoting; the variant type is the full enum name the old String() used
	// (see the red-run "GDEXTENSION_VARIANT_TYPE_VARIANT_MAX" line).
	flat := strings.ReplaceAll(got, " ", "")
	for _, want := range []string{
		`"class":"SpotStub"`,
		`"method":"GateTriple"`,
		`"argument_types":["int64/GDEXTENSION_VARIANT_TYPE_INT"]`,
		`"return_type":"int64"`,
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("entry bind content missing %q in %q", want, got)
		}
	}
	if !strings.Contains(got, "int64/int64") {
		t.Errorf("entry ret content missing the type/kind description: %q", got)
	}
}

// TestGoMethodMetadataMarshalStructure asserts the bind field's structured
// content exactly, through a JSON encoder for deterministic rendering.
func TestGoMethodMetadataMarshalStructure(t *testing.T) {
	md := newReflectiveGateMetadata()

	var buf bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zap.NewAtomicLevelAt(zapcore.InfoLevel),
	))
	logger.Info("bind-shape", zap.Object("bind", md))

	var parsed struct {
		Bind map[string]any `json:"bind"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("bind did not marshal as a JSON object: %v (%s)", err, buf.String())
	}
	bind := parsed.Bind
	if bind["class"] != "SpotStub" || bind["method"] != "GateTriple" {
		t.Errorf("bind identification drifted: %v", bind)
	}
	if bind["return_type"] != "int64" {
		t.Errorf("bind return_type drifted: %v", bind)
	}
	args, ok := bind["argument_types"].([]any)
	if !ok || len(args) != 1 || args[0] != "int64/GDEXTENSION_VARIANT_TYPE_INT" {
		t.Errorf("bind argument_types drifted (paired Go/variant descriptions expected): %v", bind["argument_types"])
	}
}

// TestNilGoMethodMetadataMarshalErrorsNotPanics: the marshaler runs inside
// the encoder, where the proportionate signal is an error, not the fatal
// path String() takes.
func TestNilGoMethodMetadataMarshalErrorsNotPanics(t *testing.T) {
	var md *GoMethodMetadata
	enc := zapcore.NewMapObjectEncoder()
	if err := md.MarshalLogObject(enc); err == nil {
		t.Error("nil metadata marshal must return an error")
	}
}

// TestVariantCollectionEmptyMarshalsWithoutEngine: an empty collection is
// zero stringifications, so it renders engine-free. A regression that
// reintroduces eager engine stringification faults the package tests instead
// of passing silently (any real element's Stringify dereferences the
// unloaded interface table).
func TestVariantCollectionEmptyMarshalsWithoutEngine(t *testing.T) {
	logger, sink, _ := newCapturingLogger(zapcore.InfoLevel)
	logger.Info("empty-args", zap.Object("gd_args", variantCollection{}))
	if got := sink.Lines(); !strings.Contains(strings.ReplaceAll(got, " ", ""), `"values":[]`) {
		t.Errorf("empty variant collection rendered %q", got)
	}
}

// TestResolvedCollectionDeduplicatesIdenticalFill: when fillCallArgs copied
// the supplied cells unchanged, the resolved_args field records the identity
// instead of re-describing every value — the scenario "a value is described
// at most once per entry" holds for the written form, and it happens with
// zero engine crossings (which is what makes this assertion engine-free).
func TestResolvedCollectionDeduplicatesIdenticalFill(t *testing.T) {
	supplied := []Variant{plantedVariant(4), plantedVariant(9)}
	resolved := append([]Variant{}, supplied...)

	logger, sink, _ := newCapturingLogger(zapcore.InfoLevel)
	logger.Info("resolved-shape",
		zap.Object("resolved_args", resolvedCollection{resolved: resolved, supplied: supplied}))
	got := sink.Lines()
	if !strings.Contains(got, "same as gd_args") {
		t.Errorf("unchanged resolved collection did not record the identity: %q", got)
	}
}

// TestSameVariantCells unit check.
func TestSameVariantCells(t *testing.T) {
	a := []Variant{plantedVariant(1), plantedVariant(2)}
	b := append([]Variant{}, a...)
	if !sameVariantCells(a, b) {
		t.Error("copied collection must compare equal")
	}
	b[1] = plantedVariant(3)
	if sameVariantCells(a, b) {
		t.Error("changed cell must compare unequal")
	}
	if sameVariantCells(a, a[:1]) {
		t.Error("different lengths must compare unequal")
	}
}
