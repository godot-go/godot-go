package main

import (
	"github.com/godot-go/godot-go/cmd/generate/classes"
	"github.com/godot-go/godot-go/cmd/generate/gdnativewrapper"
	"github.com/godot-go/godot-go/cmd/generate/types"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	verbose         bool
	cleanAll        bool
	cleanGdnative   bool
	cleanTypes      bool
	cleanClasses    bool
	genGodotApiJson bool
	genGdnative     bool
	genTypes        bool
	genClasses      bool
	packagePath     string
	godotPath       string
)

func init() {
	absPath, _ := filepath.Abs(".")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Toggle extra debug output")
	rootCmd.PersistentFlags().BoolVarP(&cleanAll, "clean", "", false, "Clean all generated files")
	rootCmd.PersistentFlags().BoolVarP(&cleanGdnative, "clean-gdnative", "", false, "Clean generated GDNative files")
	rootCmd.PersistentFlags().BoolVarP(&cleanTypes, "clean-types", "", false, "Clean generated Godot types files")
	rootCmd.PersistentFlags().BoolVarP(&cleanClasses, "clean-classes", "", false, "Clean generated Godot classes files")
	rootCmd.PersistentFlags().BoolVarP(&genGodotApiJson, "godot-api-json", "", false, "Generate Godot APIVersion JSON")
	rootCmd.PersistentFlags().BoolVarP(&genGdnative, "gdnative", "", false, "Generate gdnative")
	rootCmd.PersistentFlags().BoolVarP(&genTypes, "types", "", false, "Generate types")
	rootCmd.PersistentFlags().BoolVarP(&genClasses, "classes", "", false, "Generate classes")
	rootCmd.PersistentFlags().StringVarP(&packagePath, "package-path", "p", absPath, "Specified package path")
	rootCmd.PersistentFlags().StringVarP(&godotPath, "godot-path", "", "godot", "Specified path where the Godot executable is located")
}

func expandPattern(globPattern string) ([]string, error) {
	absGlobPattern, err := filepath.Abs(globPattern)
	if err != nil {
		return nil, err
	}

	files, err := filepath.Glob(absGlobPattern)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func removeFiles(files []string) error {
	for _, f := range files {
		println("removing file: ", f)
		if err := os.Remove(f); err != nil {
			return err
		}
	}

	return nil
}

func cleanFiles(globPatterns []string) error {
	var (
		genFiles []string
	)

	for _, p := range globPatterns {
		f, err := expandPattern(path.Join(packagePath, p))

		if err != nil {
			return err
		}

		genFiles = append(genFiles, f...)
	}

	if len(genFiles) == 0 {
		println("No generated files found to clean.")
	} else {
		// TODO: test actual remove
		//pretty.Println(genFiles)
		if err := removeFiles(genFiles); err != nil {
			return err
		}
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "godot-go",
	Short: "Godot Go",
	RunE: func(cmd *cobra.Command, args []string) error {
		hasGen := false

		if cleanAll || cleanGdnative {
			globPatterns := []string{
				"/pkg/gdnative/*.wrapper.gen.c",
				"/pkg/gdnative/*.wrapper.gen.h",
				"/pkg/gdnative/*.wrapper.gen.go",
				"/pkg/gdnative/wrappers.gen.c",
				"/pkg/gdnative/wrappers.gen.h",
				"/pkg/gdnative/wrappers.gen.go",
			}
			if err := cleanFiles(globPatterns); err != nil {
				return err
			}
		}

		if cleanAll || cleanTypes {
			globPatterns := []string{
				"/pkg/gdnative/types.gen.go",
			}
			if err := cleanFiles(globPatterns); err != nil {
				return err
			}
		}

		if cleanAll || cleanClasses {
			globPatterns := []string{
				"/pkg/gdnative/classes.gen.go",
			}
			if err := cleanFiles(globPatterns); err != nil {
				return err
			}
		}

		if genGodotApiJson {
			if verbose {
				println("Generating godot_api.json from godot --gdnative-generate-json-api...")
			}
			c := exec.Command("godot",
				"--gdnative-generate-json-api",
				path.Join(packagePath, "cmd/generate/templates/godot_api.json"),
				"--no-window")

			if err := c.Run(); err != nil {
				return err
			}
		}

		if genGdnative {
			if verbose {
				println("Generating gdnative wrapper functions...")
			}
			gdnativewrapper.Generate(packagePath)

			hasGen = true
		}

		if genTypes {
			if verbose {
				println("Generating gdnative types...")
			}
			types.Generate(packagePath)

			hasGen = true
		}

		if genClasses {
			if verbose {
				println("Generating gdnative classes...")
			}
			classes.Generate(packagePath)

			hasGen = true
		}

		if hasGen {
			outputPackageDirectoryPath := filepath.Join(packagePath, "pkg", "gdnative")

			log.Println("running go fmt on files.")
			execGoFmt(outputPackageDirectoryPath)

			log.Println("running goimports on files.")
			execGoImports(outputPackageDirectoryPath)
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
