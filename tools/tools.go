//go:build tools
// +build tools

// idiomatic way of tracking your tool dependencies:
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

package tools

import (
	_ "golang.org/x/tools/cmd/goimports"
)
