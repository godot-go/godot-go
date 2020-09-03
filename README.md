[![Actions Build Status](https://github.com/godot-go/godot-go/workflows/CI/badge.svg?branch=master)](https://github.com/godot-go/godot-go/actions)

---

# Godot-Go

## What is Godot-go?

Godot-go allows developers to implement their games in [Go](https://golang.org/) to leverage the [Godot](https://github.com/godotengine/godot) cross-platform game engine. Godot-go leverages cgo to interface with Godot's GDNative and NativeScript APIs.


## Getting Started

Please install Go version 1.15 or above to get the latest cgo improvements. I encourage everyone to install Go through [Go Version Manager](https://github.com/moovweb/gvm)


### Testing

Follow up by compiling and running the tests:

    go run mage.go test

Compiling can take more than *10 minutes* because of the use of cgo. Please be patient. Once it finishes, the tests will run and also start the demo app afterwards.


### Generating Codegen

There is a bit of codegen as part of godot-go. If you've made modifications to the generation, the goimports package will need to be installed to run `go generate`:

    go get golang.org/x/tools/cmd/goimports

Codegen files are not checked into the project. You will need to generate those codegen files like so:

    go generate
    

## References

* Inspiration was taken from ShadowApex's earlier [godot-go](https://github.com/ShadowApex/godot-go) project
* Cheatsheet for [cgo data type conversion](https://gist.github.com/zchee/b9c99695463d8902cd33)
* Cross compilation with cgo with [xgo](https://github.com/karalabe/xgo)
* vscode-go patch to [support cgo](https://github.com/golang/go/issues/35721#issuecomment-568543991)
* Check [unsafe pointer conversion](https://blog.gopheracademy.com/advent-2019/safe-use-of-unsafe-pointer/)
* Loading nativescript libraries with a godot server build requires manual modification to the library [.tres](https://godotengine.org/qa/63890/how-to-open-gdnative-projects-with-headless-server-godot).
* Working with [GDB Go extension](https://nanxiao.me/en/the-tips-of-using-gdb-to-debug-golang-program/)
