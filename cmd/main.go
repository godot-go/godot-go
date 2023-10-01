package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/godot-go/godot-go/cmd/generate/extensionapi"
	"github.com/godot-go/godot-go/cmd/generate/gdextensionwrapper"

	"github.com/spf13/cobra"
)

var (
	verbose          bool
	cleanAll         bool
	cleanGdextension bool
	cleanTypes       bool
	cleanClasses     bool
	genClangAPI      bool
	genExtensionApi  bool
	packagePath      string
	godotPath        string
	parsedASTPath    string
	buildConfig      string
)

func init() {
	absPath, _ := filepath.Abs(".")
	var (
		defaultBuildConfig string
	)
	if strings.Contains(runtime.GOARCH, "32") {
		defaultBuildConfig = "float_32"
	} else {
		defaultBuildConfig = "float_64"
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Toggle extra debug output")
	rootCmd.PersistentFlags().BoolVarP(&genClangAPI, "clang-api", "", false, "Generate GDExtension C wrapper")
	rootCmd.PersistentFlags().BoolVarP(&genExtensionApi, "extension-api", "", false, "Generate Extension API")
	rootCmd.PersistentFlags().StringVarP(&packagePath, "package-path", "p", absPath, "Specified package path")
	rootCmd.PersistentFlags().StringVarP(&godotPath, "godot-path", "", "godot", "Specified path where the Godot executable is located")
	rootCmd.PersistentFlags().StringVarP(&parsedASTPath, "parsed-ast-path", "", "_debug_parsed_ast.json", "Specified path where the AST structure should be written to")
	rootCmd.PersistentFlags().StringVarP(&buildConfig, "build-config", "", defaultBuildConfig, "Specified build configuration for built-in class sizes")
}

var rootCmd = &cobra.Command{
	Use:   "godot-go",
	Short: "Godot Go",
	RunE: func(cmd *cobra.Command, args []string) error {
		hasGen := false

		if verbose {
			println(fmt.Sprintf(`build configuration "%s" selected`, buildConfig))
		}

		if genClangAPI {
			if verbose {
				println("Generating gdextension C wrapper functions...")
			}
			gdextensionwrapper.Generate(packagePath, parsedASTPath)

			hasGen = true
		}

		if genExtensionApi {
			if verbose {
				println("Generating extension api...")
			}
			extensionapi.Generate(packagePath, buildConfig)

			hasGen = true
		}

		if hasGen {
			outputPackageDirectoryPath := filepath.Join(packagePath, "pkg", "gdextension")

			log.Println("running go fmt on files.")
			execGoFmt(outputPackageDirectoryPath)

			// log.Println("running goimports on files.")
			// execGoImports(outputPackageDirectoryPath)
		}

		if verbose {
			println("cli tool done")
		}

		return nil
	},
}

func execGoFmt(filePath string) {
	cmd := exec.Command("gofmt", "-w", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Panic(fmt.Errorf("error running gofmt: \n%s\n%w", output, err))
	}
}

func execGoImports(filePath string) {
	cmd := exec.Command("goimports", "-w", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Panic(fmt.Errorf("error running goimports: \n%s\n%w", output, err))
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
