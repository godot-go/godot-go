package gdextensionparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/godot-go/godot-go/cmd/gdextensionparser/clang"
	"github.com/godot-go/godot-go/cmd/gdextensionparser/preprocessor"
)

func GenerateGDExtensionInterfaceAST(projectPath, astOutputFilename string) (clang.CHeaderFileAST, error) {
	n := filepath.Join(projectPath, "/godot_headers/godot/gdextension_interface.h")
	b, err := os.ReadFile(n)
	if err != nil {
		return clang.CHeaderFileAST{}, fmt.Errorf("error reading %s: %w", n, err)
	}

	preprocFile, err := preprocessor.ParsePreprocessorString((string)(b))
	if err != nil {
		return clang.CHeaderFileAST{}, fmt.Errorf("error preprocessing %s: %w", n, err)
	}

	preprocText := preprocFile.Eval(false)
	ast, err := clang.ParseCString(preprocText)
	if err != nil {
		return clang.CHeaderFileAST{}, fmt.Errorf("error parsing %s: %w", n, err)
	}

	// write the AST out to a file as JSON for debugging
	if astOutputFilename != "" {
		b, err := json.Marshal(ast)
		if err != nil {
			panic(err)
		}
		f, err := os.Create(astOutputFilename)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		w.Write(b)
		w.Flush()
	}

	return ast, nil
}
