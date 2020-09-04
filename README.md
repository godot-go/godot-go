[![Actions Build Status](https://github.com/godot-go/godot-go/workflows/godot-go%20CI/badge.svg)](https://github.com/godot-go/godot-go/actions?query=workflow%3Agodot-go+branch%3Amaster)

---

# godot-go

## What is godot-go?

godot-go allows developers to implement their games in [Go](https://golang.org/) to leverage the [Godot](https://github.com/godotengine/godot) cross-platform game engine. godot-go leverages cgo to interface with Godot's GDNative and NativeScript APIs.


## Getting Started

Please install Go version 1.15 or above to get the latest cgo improvements. I encourage everyone to install Go through [Go Version Manager](https://github.com/moovweb/gvm)

Once you've checked out the project, you can go ahead and run the embedded project that's used for automated testing:

    GODOT_BIN=godot go run mage.go test

Please adjust `GODOT_BIN` to point to your godot executable. Compiling can take more than *10 minutes* because of the use of cgo. Please be patient. Once it finishes, the tests will run and also start the demo app afterwards.

Subsequent changes to the test project will be quick.

For those that are familiar with GDScript and are curious to see the code, here's an example class representing a [Player Charater](test/pkg/gdnativetest/object_player_character.go) in a 2D Top-down game.


### Support

godot-go has been tested to work with godot 3.2 in the following platforms and architectures:

| Platform      | Builds Cross-Compile from Linux | Builds from native OS | Test Pass |
| ------------- | ------------------------------- | --------------------- | --------- |
| linux/amd64   | Yes                             | Yes                   | Yes       |
| darwin/amd64  | Yes                             | Unknown               | Unknown   |
| windows/amd64 | Yes                             | Yes                   | Unknown   |
| windows/386   | Yes                             | Unknown               | Unknown   |

* The Github Workflow [test_windows](.github/workflows/test_windows.yaml) tests building on a Windows machine. However, tests fail to finish when running with godot windows distributed binaries.

My development environment is on Ubuntu 20.04; therefore, support for Linux is primary for the project. The project also compiles for Windows and MacOS, but issues may pop up that might not be caught by the continuous integration process. Please feel free to file issues as they come up.

You can find me (surgical#3758) on the [Godot Engine Discord](https://discord.gg/qZHMsDg) servers. I primarily focus in the **gdnative-dev** room.

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


## Credit

* Test project art assets taken from [Free RPG Asset Pack](https://biloumaster.itch.io/free-rpg-asset)
