[![Build Status](https://github.com/godot-go/godot-go/actions/workflows/ci_linux.yaml/badge.svg)](https://github.com/godot-go/godot-go/actions/workflows/ci_linux.yaml)

---

# godot-go: Go bindings for Godot 4.2-dev

[Go](https://golang.org/) bindings for the [Godot Game Engine](https://github.com/godotengine/godot) cross-platform game engine. godot-go integrates into Godot through the Godot GDExtension API through cgo.

The project is currently under heavy development. The API should be considered __EXPERIMENTAL__ and is subject to change. The API is expected to become more stable as we get closer to a 1.0 release.

## Current State of the Project

Here are some high level bullet points:
* A bleeding edge version of godot 4.2 will be required.
* The tests confirm basic functionality.
* The bindings are more than likely leaking memory.
* Not all variant types have been tested.
* A working demo project should be possible. We'll have one avaialble and linked out in the next coming week.

## Getting Started

Requirements:
* clang-format
* gcc
* go 1.21.x

TODO


### Building Godot-Go

In order for godot-go to work correctly, you must make sure the godot_headers are in sync between your godot binary and godot-go. Development is built and tested off of [Godot@44e399e](https://github.com/godotengine/godot/tree/44e399ed5fa895f760b2995e59788bdb49782666). The bindings will not work unless running with a bleeding edge version of Godot as of 2023-10-01 as the bindings are dependent on changes to the GDExtensions interface that bring better virtual call support.


    # exports the latest gdextension_interface.h and extension_api.json from the godot binary
    GODOT=/godot_folder_path/bin/godot make update_godot_headers_from_binary

    # generates code for wrapping gdextension_interface.h and extension_api.json
    make generate

    # build godot-go
    make build


### Test

Once the project successfully builds, run the following commands to generate cached files for the test demo project for the first time (don't be concerned if it fails):

    make ci_gen_test_project_files

From here on out, you will just need to run the following command to iteratively test:

    make test

This will run the demo project in the test directory.

The expected output can be found in [CI logs](https://github.com/godot-go/godot-go/actions/runs/6373041267/job/17296433849#step:11:131)

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
