# Godot Go Binding

[Godot](https://github.com/godotengine/godot) is a cross-platform game engine.

## Getting Started

Please install Go version 1.15 or above to get the latest cgo improvements. I encourage everyone to install Go through [Go Version Manager](https://github.com/moovweb/gvm)

### Generating Codegen

The goimports package will need to be installed to run `go generate`:

    go get golang.org/x/tools/cmd/goimports

Codegen files are not checked into the project. You will need to generate those codegen files like so:

    go generate
    
    
### Testing

Follow up by compiling and running the tests:

    go run mage.go test

Compiling can take more than *10 minutes* because of the use of cgo. Please be patient. Once it finishes, the tests will run and also start the demo app afterwards.


## Releasing

The release process include comitting codegen files on git tags. To package and deploy a release:

    ./package-release.sh 0.0.1 publish

Generating releases manually should never be required.

---

[![Actions Build Status](https://github.com/godot-go/godot-go/workflows/CI/badge.svg?branch=master)](https://github.com/godot-go/godot-go/actions)
