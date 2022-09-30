[![Actions Build Status](https://github.com/godot-go/godot-go/workflows/godot-go%20CI/badge.svg)](https://github.com/godot-go/godot-go/actions?query=workflow%3Agodot-go+branch%3Amaster)

---

# godot-go: Go bindings for Godot 4

[Go](https://golang.org/) bindings for the [Godot Game Engine](https://github.com/godotengine/godot) cross-platform game engine. godot-go integrates into Godot through the Godot GDExtension API through cgo.

The project is currently under heavy development. The API should be considered __EXPERIMENTAL__ and is subject to change. The API is expected to become more stable as we get closer to a 1.0 release.


## Getting Started

Requirements:
* clang-format
* gcc
* go 1.19.x

TODO


### Building Godot-Go

**The GDExtension interface hasn't been stable.** In order for godot-go to work correctly, you must make sure the godot_headers are in sync between your godot binary and godot-go. There's a helper task in the Makefile for this:

    make update_godot_headers_from_github
    make generate

If your godot binary is built from source, you can run this:

    GODOT_SRC=/usr/local/src/godot/ make update_godot_headers_from_source
    make generate

Development is built and tested off of Godot 4 beta2 [snapshot](https://github.com/godotengine/godot-headers/tree/f8745f2f71c79972df66f17a3da75f6e328bc55d).

Once that's done, you can run the following to build:

    make build


### Test

Once the project successfully builds, run the following commands to test:

    make test

This will run the demo project in the test directory.

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
