//+build mage

// This is the build script for Mage. The install target is all you really need.
// The release target is for generating official releases and is really only
// useful to project admins.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type BuildPlatform struct {
	OS   string
	Arch string
}

var (
	godotBin   string
	ci         bool
	targetOS   string
	targetArch string
)

func init() {
	var (
		ok  bool
	)

	if targetOS, ok = os.LookupEnv("TARGET_OS"); !ok {
		targetOS = runtime.GOOS
	}

	if targetArch, ok = os.LookupEnv("TARGET_ARCH"); !ok {
		targetArch = runtime.GOARCH
	}

	envCI, _ := os.LookupEnv("CI")
	ci = envCI == "true"
}

func initGodotBin() {
	var (
		err error
	)

	godotBin, _ = os.LookupEnv("GODOT_BIN")

	if godotBin, err = which(godotBin); err == nil {
		fmt.Printf("GODOT_BIN = %s\n", godotBin)
		return
	}

	if !ci {
		if godotBin, err = which("godot3"); err == nil {
			fmt.Printf("GODOT_BIN = %s\n", godotBin)
			return
		}

		if godotBin, err = which("godot"); err == nil {
			fmt.Printf("GODOT_BIN = %s\n", godotBin)
			return
		}
	}

	panic(err)
}

func envWithPlatform(platform BuildPlatform) map[string]string {
	envs := map[string]string{
		"GOOS":                   targetOS,
		"GOARCH":                 targetArch,
		"CGO_ENABLED":            "1",
		"asyncpremptoff":         "1",
	}

	// enable for cross-compiling from linux
	// case "windows":
	// 	envs["CC"] = "i686-w64-mingw32-gcc"
	// }

	return envs
}

func CleanGdnative() error {
	return sh.RunV("go", "run", "cmd/main.go", "--verbose", "--clean-gdnative")
}

func CleanTypes() error {
	return sh.RunV("go", "run", "cmd/main.go", "--verbose", "--clean-types")
}

func CleanClasses() error {
	return sh.RunV("go", "run", "cmd/main.go", "--verbose", "--clean-classes")
}

func Clean() error {
	mg.Deps(CleanClasses, CleanTypes, CleanGdnative)
	return nil
}

func Generate() error {
	return sh.RunV("go", "generate", "main.go")
}

func InstallGoTools() error {
	return sh.RunV("go", "get", "golang.org/x/tools/cmd/goimports")
}

func BuildTest() error {
	appPath := filepath.Join("test")
	outputPath := filepath.Join(appPath, "project", "libs")

	return buildGodotPlugin(
		"test",
		appPath,
		outputPath,
		BuildPlatform{
			OS:   targetOS,
			Arch: targetArch,
		},
	)
}

func Test() error {
	mg.Deps(BuildTest)

	appPath := filepath.Join("test")

	return runPlugin(appPath)
}

func runPlugin(appPath string) error {
	mg.Deps(initGodotBin)

	return sh.RunWith(
		map[string]string{
			"asyncpremptoff": "1",
			"cgocheck": "2",
			"LOG_LEVEL": "trace",
			"TEST_USE_GINKGO_WRITER": "1",
		}, 
		godotBin, "--verbose", "-v", "-d", 
		"--path", filepath.Join(appPath, "project"))
}

func buildGodotPlugin(name string, appPath string, outputPath string, platform BuildPlatform) error {
	return sh.RunWith(envWithPlatform(platform), mg.GoCmd(), "build",
		"-buildmode=c-shared", "-x", "-trimpath",
		"-o", filepath.Join(outputPath, platform.godotPluginCSharedName(appPath, name)),
		filepath.Join(appPath, "main.go"),
	)
}

func (p BuildPlatform) godotPluginCSharedName(appPath string, varargs ...string) string {
	switch(p.OS) {
		case "windows":
			return fmt.Sprintf("libgodotgo-%s-windows-4.0-%s.dll", strings.Join(varargs, "-"), p.Arch)
		case "linux":
			ext := "so"
			
			if len(varargs) == 0 {
				return fmt.Sprintf("libgodotgo-%s-%s.%s", p.OS, p.Arch, ext)
			}
			
			return fmt.Sprintf("libgodotgo-%s-%s-%s.%s", strings.Join(varargs, "-"), p.OS, p.Arch, ext)
		default:
			panic(fmt.Errorf("unsupported build platform: %s", p.OS))
	}
}

func which(filename string) (string, error) {
	if len(filename) == 0 {
		return "", fmt.Errorf("no filename specified")
	}
	return sh.Output("which", filename)
}

var Default = BuildTest
