package gdclassimpl

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	_ "embed"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
	"github.com/iancoleman/strcase"
)

var (
	//go:embed classes.go.tmpl
	classesText string

	//go:embed classes.refs.go.tmpl
	classesRefsText string

	//go:embed virtuals.go.tmpl
	virtualsText string
)

// Generate will generate Go wrappers for all Godot base types
func Generate(projectPath string, eapi extensionapiparser.ExtensionApi) {
	var (
		err error
	)
	if err = GenerateClasses(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateClassRefs(projectPath, eapi); err != nil {
		panic(err)
	}
	if err = GenerateVirtuals(projectPath, eapi); err != nil {
		panic(err)
	}
}

// VirtualClassInfo is one engine class's Category-A virtual surface.
type VirtualClassInfo struct {
	Name       string // Godot class name, e.g. "Control".
	StructName string // Generated Go struct name, e.g. "Control".
	Methods    []VirtualMethodInfo
}

// VirtualMethodInfo is one Category-A virtual with its generated qualified name.
type VirtualMethodInfo struct {
	extensionapiparser.ClassMethod
	QualifiedName string // e.g. "V_Control_GetMaximumSize".
}

// DeriveCategoryA classifies every is_virtual method of the API into the
// Category-A set: non-void return and a plain wrapper named without the
// leading underscore exists somewhere on the class hierarchy (the engine
// consumes the result through that wrapper). The returned map is keyed by
// Godot class name. This is the same criterion documented in the census
// fixture cmd/generate/gdclassimpl/virtual_census.json.
func DeriveCategoryA(eapi extensionapiparser.ExtensionApi) map[string][]extensionapiparser.ClassMethod {
	methods := make(map[string]map[string]struct{}, len(eapi.Classes))
	inherits := make(map[string]string, len(eapi.Classes))
	for _, c := range eapi.Classes {
		set := make(map[string]struct{}, len(c.Methods))
		for _, m := range c.Methods {
			if !m.IsVirtual {
				set[m.Name] = struct{}{}
			}
		}
		methods[c.Name] = set
		if c.Inherits != "" {
			inherits[c.Name] = c.Inherits
		}
	}

	hasWrapper := func(className string, wrapper string) bool {
		for cls := className; cls != ""; {
			if _, ok := methods[cls][wrapper]; ok {
				return true
			}
			var ok2 bool
			cls, ok2 = inherits[cls]
			if !ok2 {
				break
			}
		}
		return false
	}

	out := make(map[string][]extensionapiparser.ClassMethod)
	for _, c := range eapi.Classes {
		for _, m := range c.Methods {
			if !m.IsVirtual {
				continue
			}
			rt := m.ReturnValue.Type
			if rt == "" || rt == "void" {
				continue
			}
			wrapper := strings.TrimPrefix(m.Name, "_")
			if !hasWrapper(c.Name, wrapper) {
				continue
			}
			out[c.Name] = append(out[c.Name], m)
		}
	}
	return out
}

// buildVirtualClassesView prepares the template data: every class with
// Category-A virtuals, in declaration order, with qualified method names.
func buildVirtualClassesView(eapi extensionapiparser.ExtensionApi) []VirtualClassInfo {
	categoryA := DeriveCategoryA(eapi)
	structName := make(map[string]string, len(eapi.Classes))
	view := make([]VirtualClassInfo, 0, len(categoryA))
	for _, c := range eapi.Classes {
		ms, ok := categoryA[c.Name]
		if !ok || len(ms) == 0 {
			continue
		}
		sn := c.Name
		structName[c.Name] = sn
		info := VirtualClassInfo{Name: c.Name, StructName: sn}
		for _, m := range ms {
			info.Methods = append(info.Methods, VirtualMethodInfo{
				ClassMethod:   m,
				QualifiedName: fmt.Sprintf("V_%s_%s", sn, strcase.ToCamel(strings.TrimPrefix(m.Name, "_"))),
			})
		}
		view = append(view, info)
	}
	return view
}

type virtualsView struct {
	extensionapiparser.ExtensionApi
	VirtualClasses []VirtualClassInfo
}

// GenerateVirtuals emits pkg/gdclassimpl/virtuals.gen.go: per-class
// declaration-only interfaces cataloging the Category-A virtual surface.
func GenerateVirtuals(projectPath string, eapi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("virtuals.gen.go").
		Funcs(template.FuncMap{
			"isSetterMethodName":   isSetterMethodName,
			"goVariantConstructor": goVariantConstructor,
			"goMethodName":         goMethodName,
			"goArgumentName":       goArgumentName,
			"goArgumentType":       goArgumentType,
			"goVariantFunc":        goVariantFunc,
			"goReturnType":         goReturnType,
			"goClassEnumName":      goClassEnumName,
			"goClassStructName":    goClassStructName,
			"goClassInterfaceName": goClassInterfaceName,
			"goEncoder":            goEncoder,
			"goEncodeIsReference":  goEncodeIsReference,
			"coalesce":             coalesce,
			"typeOrMeta":           typeOrMeta,
		}).
		Parse(virtualsText)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, virtualsView{ExtensionApi: eapi, VirtualClasses: buildVirtualClassesView(eapi)})
	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdclassimpl", "virtuals.gen.go")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(b.Bytes())
	return err
}

func GenerateClasses(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.gen.go").
		Funcs(template.FuncMap{
			"isSetterMethodName":   isSetterMethodName,
			"goVariantConstructor": goVariantConstructor,
			"goMethodName":         goMethodName,
			"goArgumentName":       goArgumentName,
			"goArgumentType":       goArgumentType,
			"goVariantFunc":        goVariantFunc,
			"goReturnType":         goReturnType,
			"goClassEnumName":      goClassEnumName,
			"goClassStructName":    goClassStructName,
			"goClassInterfaceName": goClassInterfaceName,
			"goEncoder":            goEncoder,
			"goEncodeIsReference":  goEncodeIsReference,
			"coalesce":             coalesce,
			"typeOrMeta":           typeOrMeta,
		}).
		Parse(classesText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdclassimpl", fmt.Sprintf("classes.gen.go"))

	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(b.Bytes())

	if err != nil {
		return err
	}

	return nil
}

func GenerateClassRefs(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("classes.refs.gen.go").
		Funcs(template.FuncMap{
			"isSetterMethodName":   isSetterMethodName,
			"goVariantConstructor": goVariantConstructor,
			"goMethodName":         goMethodName,
			"goArgumentName":       goArgumentName,
			"goArgumentType":       goArgumentType,
			"goVariantFunc":        goVariantFunc,
			"goReturnType":         goReturnType,
			"goClassEnumName":      goClassEnumName,
			"goClassStructName":    goClassStructName,
			"goClassInterfaceName": goClassInterfaceName,
			"goEncoder":            goEncoder,
			"goEncodeIsReference":  goEncodeIsReference,
			"coalesce":             coalesce,
			"typeOrMeta":           typeOrMeta,
		}).
		Parse(classesRefsText)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, extensionApi)

	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdclassimpl", fmt.Sprintf("classes.refs.gen.go"))

	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(b.Bytes())

	if err != nil {
		return err
	}

	return nil
}
