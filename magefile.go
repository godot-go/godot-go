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
	logLevel   string
)

func init() {
	var (
		ok  bool
	)

	if logLevel, ok = os.LookupEnv("LOG_LEVEL"); !ok {
		logLevel = "trace"
	}

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

		if godotBin, err = which("/Applications/Godot.app/Contents/MacOS/Godot"); err == nil {
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
	}

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
	mg.Deps(InstallGoTools)
	return sh.RunV("go", "generate")
}

func InstallGoTools() error {
	return sh.RunV("go", "get", "golang.org/x/tools/cmd/goimports")
}

func BuildTest() error {
	appPath := filepath.Join("test")
	outputPath := filepath.Join(appPath, "project", "libs")

	return buildGodotPlugin(
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
			"GOTRACEBACK": "crash",
			"GODEBUG": "asyncpreemptoff=0,cgocheck=1,invalidptr=1,clobberfree=1,tracebackancestors=0",
			"LOG_LEVEL": "debug",
			"TEST_USE_GINKGO_RECOVER": "1",
			"TEST_USE_GINKGO_WRITER": "1",
		},
		godotBin, "--verbose", "-v", "-d",
		"--path", filepath.Join(appPath, "project"))
}

func buildGodotPlugin(appPath string, outputPath string, platform BuildPlatform) error {
	return sh.RunWith(envWithPlatform(platform), mg.GoCmd(), "build",
		"-gcflags=all=-d=checkptr",
		"-tags", "tools", "-buildmode=c-shared", "-x", "-trimpath",
		"-o", filepath.Join(outputPath, platform.godotPluginCSharedName(appPath)),
		filepath.Join(appPath, "main.go"),
	)
}

func (p BuildPlatform) godotPluginCSharedName(appPath string) string {
	// NOTE: these files needs to line up with CI as well as the naming convention
	//       expected by the test godot project
	switch(p.OS) {
		case "windows":
			return fmt.Sprintf("libgodotgo-test-windows-4.0-%s.dll", p.Arch)
		case "darwin":
			return fmt.Sprintf("libgodotgo-test-darwin-10.6-%s.dylib", p.Arch)
		case "linux":
			return fmt.Sprintf("libgodotgo-test-linux-%s.so", p.Arch)
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
