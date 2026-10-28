package typedwrapper

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed wrappers.go.tmpl
var wrappersText string

//go:embed ret_encoders.go.tmpl
var retEncodersText string

//go:embed binds.go.tmpl
var bindsText string

type templateData struct {
	ArityRange []int
}

// retEncoderType is one copy-return builtin: the generated retEncodeInto
// delegates to this type's engine copy constructor. Membership is the audited
// return-encoder matrix's copy-constructor set (design D3); String is
// deliberately absent (raw store per StringEncoder.EncodeTypePtrArg).
type retEncoderType struct {
	Type string
	Ctor string
}

var retEncoderTypes = []retEncoderType{
	{"StringName", "StringNameCopyConstructor"},
	{"NodePath", "NodePathCopyConstructor"},
	{"Callable", "CallableCopyConstructor"},
	{"Signal", "SignalCopyConstructor"},
	{"Array", "ArrayCopyConstructor"},
	{"Dictionary", "DictionaryCopyConstructor"},
	{"PackedByteArray", "PackedByteArrayCopyConstructor"},
	{"PackedInt32Array", "PackedInt32ArrayCopyConstructor"},
	{"PackedInt64Array", "PackedInt64ArrayCopyConstructor"},
	{"PackedFloat32Array", "PackedFloat32ArrayCopyConstructor"},
	{"PackedFloat64Array", "PackedFloat64ArrayCopyConstructor"},
	{"PackedStringArray", "PackedStringArrayCopyConstructor"},
	{"PackedVector2Array", "PackedVector2ArrayCopyConstructor"},
	{"PackedVector3Array", "PackedVector3ArrayCopyConstructor"},
	{"PackedVector4Array", "PackedVector4ArrayCopyConstructor"},
	{"PackedColorArray", "PackedColorArrayCopyConstructor"},
}

// Generate runs every typed-dispatch emission in the project rooted at
// projectPath: the ptrcall wrapper family for pkg/core, the copy-return
// encoder methods for pkg/builtin, and the typed bind overloads for pkg/core.
func Generate(projectPath string) error {
	if err := generateWrappers(projectPath); err != nil {
		return err
	}
	if err := generateRetEncoders(projectPath); err != nil {
		return err
	}
	return generateBinds(projectPath)
}

// generateWrappers renders wrappers.go.tmpl and writes pkg/core/wrappers.gen.go.
func generateWrappers(projectPath string) error {
	tmpl, err := template.New("wrappers.gen.go").Funcs(template.FuncMap{
		// counts returns the zero-based argument indexes for one arity, so
		// the template can render both the A1..An parameter list and the
		// matching args[i] indices from one loop.
		"counts": func(n int) []int {
			out := make([]int, n)
			for i := range out {
				out[i] = i
			}
			return out
		},
		"inc": func(i int) int { return i + 1 },
	}).Parse(wrappersText)
	if err != nil {
		return err
	}

	arity := 16
	data := templateData{ArityRange: make([]int, 0, arity+1)}
	for i := 0; i <= arity; i++ {
		data.ArityRange = append(data.ArityRange, i)
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return err
	}

	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return fmt.Errorf("format generated wrappers: %w", err)
	}

	filename := filepath.Join(projectPath, "pkg", "core", "wrappers.gen.go")

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(formatted)

	return err
}

// generateRetEncoders renders ret_encoders.go.tmpl and writes
// pkg/builtin/typed_ret_encoders.gen.go.
func generateRetEncoders(projectPath string) error {
	tmpl, err := template.New("typed_ret_encoders.gen.go").Parse(retEncodersText)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, struct{ Types []retEncoderType }{retEncoderTypes}); err != nil {
		return err
	}

	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return fmt.Errorf("format generated ret encoders: %w", err)
	}

	filename := filepath.Join(projectPath, "pkg", "builtin", "typed_ret_encoders.gen.go")

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(formatted)

	return err
}

// generateBinds renders binds.go.tmpl and writes pkg/core/typed_binds.gen.go,
// sharing the wrapper template's arity funcmap.
func generateBinds(projectPath string) error {
	tmpl, err := template.New("typed_binds.gen.go").Funcs(template.FuncMap{
		"counts": func(n int) []int {
			out := make([]int, n)
			for i := range out {
				out[i] = i
			}
			return out
		},
		"inc": func(i int) int { return i + 1 },
	}).Parse(bindsText)
	if err != nil {
		return err
	}

	arity := 16
	data := templateData{ArityRange: make([]int, 0, arity+1)}
	for i := 0; i <= arity; i++ {
		data.ArityRange = append(data.ArityRange, i)
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return err
	}

	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return fmt.Errorf("format generated binds: %w", err)
	}

	filename := filepath.Join(projectPath, "pkg", "core", "typed_binds.gen.go")

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(formatted)

	return err
}
