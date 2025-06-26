[![godot-go CI on Linux](https://github.com/godot-go/godot-go/actions/workflows/ci_linux.yaml/badge.svg?branch=main)](https://github.com/godot-go/godot-go/actions/workflows/ci_linux.yaml)
[![godot-go CI on Windows](https://github.com/godot-go/godot-go/actions/workflows/ci_windows.yaml/badge.svg?branch=main)](https://github.com/godot-go/godot-go/actions/workflows/ci_windows.yaml)

---

# godot-go: Go bindings for Godot 4.3

[Go](https://golang.org/) bindings for the [Godot Game Engine](https://github.com/godotengine/godot) cross-platform game engine. godot-go integrates into Godot through the Godot GDExtension API through cgo.

The project is currently under heavy development. The API should be considered __EXPERIMENTAL__ and is subject to change. The API is expected to become more stable as we get closer to a 1.0 release.

## Current State of the Project

Here are some high level bullet points:
* You must have a minimum of Godot 4.3.
* The tests confirm basic functionality.
* The bindings are more than likely leaking memory.
* Not all variant types have been tested.
* Go to [godot-go-demo-projects](https://github.com/godot-go/godot-go-demo-projects) to look at demo projects running a mix of godot-go and GDScript classes.

## Getting Started

Requirements:
* clang-format
* gcc
* go 1.21.x
* goimport

### Building Godot-Go

In order for godot-go to work correctly, you must make sure the godot_headers are in sync between your godot binary and godot-go. Development is built and tested off of [Godot 4.3](https://github.com/godotengine/godot-builds/releases/tag/4.3-stable).

    # Install go dependencies
    make installdeps

    # exports the latest gdextension_interface.h and extension_api.json from the godot binary
    GODOT=/some_path/bin/godot make update_godot_headers_from_binary

    # generates code for wrapping gdextension_interface.h and extension_api.json
    make generate

    # build godot-go
    make build

### Test

Once the project successfully builds, run the following commands to generate cached files for the test demo project for the first time (don't be concerned if it fails):

    GODOT=/some_path/bin/godot make ci_gen_test_project_files

From here on out, you will just need to run the following command to iteratively test:

    GODOT=/some_path/bin/godot make test

This will run the demo project in the test directory.

The expected output can be found in the CI logs.



## Documentation

A living doc of godot-go is maintained at [docs/overview.md](docs/overview.md).

## Contact

I'm happy to help out anyone interested in the project. Please leave a message in the [Discussion boards](https://github.com/godot-go/godot-go/discussions) or you can add me (surgical#3758) as a friend on the [Godot Engine Discord](https://discord.gg/qZHMsDg) servers. I primarily frequent the **gdnative-dev** room.

## References

* Go 101 article on [Type-Unsafe Pointers](https://go101.org/article/unsafe.html)
* Cheatsheet for [cgo data type conversion](https://gist.github.com/zchee/b9c99695463d8902cd33)
* Cross compilation with cgo with [xgo](https://github.com/karalabe/xgo)
* vscode-go patch to [support cgo](https://github.com/golang/go/issues/35721#issuecomment-568543991)
* Check [unsafe pointer conversion](https://blog.gopheracademy.com/advent-2019/safe-use-of-unsafe-pointer/)
* Loading nativescript libraries with a godot server build requires manual modification to the library [.tres](https://godotengine.org/qa/63890/how-to-open-gdnative-projects-with-headless-server-godot).
* Working with [GDB Go extension](https://nanxiao.me/en/the-tips-of-using-gdb-to-debug-golang-program/)

## Credit

* Inspiration for the project was taken from ShadowApex's earlier project: [godot-go](https://github.com/ShadowApex/godot-go)
* Inspiration also from [godot-cpp](https://github.com/godotengine/godot-cpp/)
