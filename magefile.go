//+build mage

// This is the build script for Mage. The install target is all you really need.
// The release target is for generating official releases and is really only
// useful to project admins.
package main

import (
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type BuildPlatform struct {
	OS   string
	Arch string
}

var (
	cForGoCmd = "c-for-go"
)

func envWithPlatform(platform BuildPlatform) map[string]string {
	return map[string]string{
		"GOOS":        platform.OS,
		"GOARCH":      platform.Arch,
		"CGO_ENABLED": "1",
	}
}

func Generate() error {
	mg.Deps(Clean)

	return sh.RunV(cForGoCmd, "-ccdefs", "-ccincl", "-debug", "godot-go.yaml")
}

// Remove the temporarily generated files from Release.
func Clean() error {
	files := []string{"cgo_helpers.go", "cgo_helpers.c", "cgo_helpers.h", "const.go", "godot.go", "types.go"}

	for _, f := range files {
		p := filepath.Join("pkg", "godot", f)
		if err := sh.Rm(p); err != nil {
			return err
		}
	}

	return nil
}

// tag returns the git tag for the current branch or "" if none.
func tag() string {
	s, _ := sh.Output("git", "describe", "--tags")
	return s
}

// hash returns the git hash for the current repo or "" if none.
func hash() string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return hash
}

var Default = Generate
