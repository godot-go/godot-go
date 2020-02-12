package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/godot-go/godot-go/cmd/generate/extensionapi"
	"github.com/godot-go/godot-go/cmd/generate/gdnativewrapper"

	"github.com/spf13/cobra"
)

var (
	verbose         bool
	cleanAll        bool
	cleanGdnative   bool
	cleanTypes      bool
	cleanClasses    bool
	genClangAPI     bool
	genExtensionApi bool
	packagePath     string
	godotPath       string
)

func init() {
	absPath, _ := filepath.Abs(".")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Toggle extra debug output")
	rootCmd.PersistentFlags().BoolVarP(&genClangAPI, "clangApi", "", false, "Generate GDNative C wrapper")
	rootCmd.PersistentFlags().BoolVarP(&genExtensionApi, "extensionApi", "", false, "Generate Extension API")
	rootCmd.PersistentFlags().StringVarP(&packagePath, "package-path", "p", absPath, "Specified package path")
	rootCmd.PersistentFlags().StringVarP(&godotPath, "godot-path", "", "godot", "Specified path where the Godot executable is located")
}

var rootCmd = &cobra.Command{
	Use:   "godot-go",
	Short: "Godot Go",
	RunE: func(cmd *cobra.Command, args []string) error {
		hasGen := false

		if genClangAPI {
			if verbose {
				println("Generating gdnative C wrapper functions...")
			}
			gdnativewrapper.Generate(packagePath)

			hasGen = true
		}

		if genExtensionApi {
			if verbose {
				println("Generating extension api...")
			}
			extensionapi.Generate(packagePath)

			hasGen = true
		}

		if hasGen {
			outputPackageDirectoryPath := filepath.Join(packagePath, "pkg", "gdnative")

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
		log.Panic(fmt.Errorf("error running gofmt: \n%s\n%w", string(output), err))
	}
}

func execGoImports(filePath string) {
	cmd := exec.Command("goimports", "-w", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Panic(fmt.Errorf("error running goimports: \n%s\n%w", string(output), err))
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
