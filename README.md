# Godot Go Binding

[Godot](https://github.com/godotengine/godot) is a cross-platform game engine

## Getting Started

Codegen files are not checked into the project. You will need to generate those codegen files like so:

    go run mage.go generate

Follow up by compiling and running the tests:

    go run mage.go test

Compiling may take more than *10 minutes* because of the use of cgo. Please be patient.


## Releasing

The release process include comitting codegen files on git tags. To package and deploy a release:

    ./package-release.sh 0.0.1 publish


## Dependencies

    go get golang.org/x/tools/cmd/goimports


## References

* Cheatsheet for [cgo data type conversion](https://gist.github.com/zchee/b9c99695463d8902cd33)
* Cross compilation with cgo with [xgo](https://github.com/karalabe/xgo)
* vscode-go patch to [support cgo](https://github.com/golang/go/issues/35721#issuecomment-568543991)
* Check [unsafe pointer conversion](https://blog.gopheracademy.com/advent-2019/safe-use-of-unsafe-pointer/)
