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
	godotBin string
	ci       bool
)

func init() {
	envCI, _ := os.LookupEnv("CI")

	ci = envCI == "true"

	var (
		ok bool
		err error
	)

	if ci {
		if godotBin, err = which("godot3-server"); err == nil {
			return
		}
	} else {
		godotBin = "godot_engine/bin/godot.x11.tools.64.llvm"

		if ok = fileExists(godotBin); ok {
			return
		}

		if godotBin, err = which("godot3"); err == nil {
			return
		}

		if godotBin, err = which("godot"); err == nil {
			return
		}
	}

	panic(err)
}

func envWithPlatform(platform BuildPlatform) map[string]string {
	return map[string]string{
		"GOOS":                   platform.OS,
		"GOARCH":                 platform.Arch,
		"GODEBUG":                "cgocheck=2",
		"CGO_LDFLAGS_ALLOW":      "pkg-config",
		"CGO_CFLAGS_ALLOW":       "pkg-config",
		"CGO_ENABLED":            "1",
		"asyncpremptoff":         "1",
	}
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
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	)
}

func Test() error {
	mg.Deps(BuildTest)

	appPath := filepath.Join("test")

	return runPlugin(appPath)
}

func runPlugin(appPath string) error {
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
	return sh.RunWith(envWithPlatform(platform), mg.GoCmd(), "build", "-x", "-work",
		"-buildmode=c-shared", "-ldflags=\"-d=checkptr -compressdwarf=false\"", "-gcflags=\"all=-N -l\"",
		"-o", filepath.Join(outputPath, platform.godotPluginCSharedName(appPath, name)),
		filepath.Join(appPath, "main.go"),
	)
}

func (p *BuildPlatform) godotPluginCSharedName(appPath string, varargs ...string) string {
	if len(varargs) == 0 {
		return fmt.Sprintf("libgodotgo-%s-%s.so", p.OS, p.Arch)
	}
	
	return fmt.Sprintf("libgodotgo-%s-%s-%s.so", strings.Join(varargs, "-"), p.OS, p.Arch)
}

func which(filename string) (string, error) {
	return sh.Output("which", filename)
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

var Default = BuildTest
