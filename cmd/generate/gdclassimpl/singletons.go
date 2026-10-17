package gdclassimpl

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	_ "embed"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
)

//go:embed singletons.go.tmpl
var singletonsText string

// SingletonAccessorInfo is one generated typed singleton accessor. Names are
// derived from distinct fields of the API entry: the engine lookup string and
// the accessor name come from the singleton name, while the Go type and the
// wrapper constructor come from the entry's class type.
type SingletonAccessorInfo struct {
	Singleton       string // engine singleton name, e.g. "OS".
	ClassName       string // generated class interface, e.g. "OS".
	FuncName        string // accessor function, e.g. "GetOSSingleton".
	BuilderName     string // cold-path builder, e.g. "buildOSSingleton".
	SlotVar         string // atomic holder slot, e.g. "osSingletonSlot".
	LockVar         string // cold-path mutex, e.g. "osSingletonLock".
	ConstructorName string // e.g. "NewOSWithGodotOwnerObject".
}

// lowerCamel lowercases the leading uppercase run of an identifier following
// Go conventions: "OS" -> "os", "InputMap" -> "inputMap",
// "GDScriptLanguageProtocol" -> "gdScriptLanguageProtocol". The final capital
// of a run longer than one starts the next word and stays capital, unless the
// whole string is the run.
func lowerCamel(s string) string {
	run := 0
	for run < len(s) && s[run] >= 'A' && s[run] <= 'Z' {
		run++
	}
	if run == 0 {
		return s
	}
	if run == len(s) {
		return strings.ToLower(s)
	}
	if run > 1 {
		run--
	}
	return strings.ToLower(s[:run]) + s[run:]
}

// DeriveSingletonAccessors maps every engine singleton declared in the API
// metadata to one accessor descriptor, in declaration order.
func DeriveSingletonAccessors(eapi extensionapiparser.ExtensionApi) []SingletonAccessorInfo {
	infos := make([]SingletonAccessorInfo, 0, len(eapi.Singletons))
	for _, s := range eapi.Singletons {
		prefix := lowerCamel(s.Name)
		infos = append(infos, SingletonAccessorInfo{
			Singleton:       s.Name,
			ClassName:       s.Type,
			FuncName:        "Get" + s.Name + "Singleton",
			BuilderName:     "build" + s.Name + "Singleton",
			SlotVar:         prefix + "SingletonSlot",
			LockVar:         prefix + "SingletonLock",
			ConstructorName: "New" + s.Type + "WithGodotOwnerObject",
		})
	}
	return infos
}

// singletonsTemplateData is the root value rendered by singletons.go.tmpl.
type singletonsTemplateData struct {
	Singletons []SingletonAccessorInfo
}

// GenerateSingletons writes typed accessors for every engine singleton in the
// API metadata to pkg/gdclassimpl/singletons.gen.go.
func GenerateSingletons(projectPath string, extensionApi extensionapiparser.ExtensionApi) error {
	tmpl, err := template.New("singletons.gen.go").Parse(singletonsText)
	if err != nil {
		return err
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, singletonsTemplateData{
		Singletons: DeriveSingletonAccessors(extensionApi),
	})
	if err != nil {
		return err
	}

	filename := filepath.Join(projectPath, "pkg", "gdclassimpl", "singletons.gen.go")

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(b.Bytes())

	return err
}
