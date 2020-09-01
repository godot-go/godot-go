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

Compiling can take more than *10 minutes* because of the use of cgo. Please be patient.


## Releasing

The release process include comitting codegen files on git tags. To package and deploy a release:

    ./package-release.sh 0.0.1 publish

Generating releases manually should never be required.

## Dependencies

    go get golang.org/x/tools/cmd/goimports


## References

* Cheatsheet for [cgo data type conversion](https://gist.github.com/zchee/b9c99695463d8902cd33)
* Cross compilation with cgo with [xgo](https://github.com/karalabe/xgo)
* vscode-go patch to [support cgo](https://github.com/golang/go/issues/35721#issuecomment-568543991)
* Check [unsafe pointer conversion](https://blog.gopheracademy.com/advent-2019/safe-use-of-unsafe-pointer/)
* Loading nativescript libraries require modifying the library [.tres](https://godotengine.org/qa/63890/how-to-open-gdnative-projects-with-headless-server-godot).
* Working with [GDB Go extension](https://nanxiao.me/en/the-tips-of-using-gdb-to-debug-golang-program/)

---

[![Actions Build Status](https://github.com/pcting/godot-go/workflows/CI/badge.svg?branch=master)](https://github.com/pcting/godot-go/actions)
