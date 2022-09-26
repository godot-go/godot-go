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


### Support

Development is generally bulding and testing off of Godot's master branch. As of 2022 Sept 26, runs off of [ca1ebf9fee8d58718d41b2c08d22d484764b7f54](https://github.com/godotengine/godot/tree/581db8b4e60c2a2fa4d0be076030b326784c69bb)


## Building Godot
```
$ git checkout ca1ebf9fee8d58718d41b2c08d22d484764b7f54
$ scons platform=linuxbsd use_llvm=yes
```


### Building Godot-Go from Source

To quickly build from source, check out the code and run the following commands:

    make generate && make build


### Test

Once the project successfully builds, run the following commands to test:

    make test

This will run the demo project in the test directory.

**NOTE:** The GDExtension interface isn't stable. If the tests fail to pass, please rebuild godot against the latest version documented here under [Support](#Support)


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
