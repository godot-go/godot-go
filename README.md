# godot-go

Regenerate go files:

    go run mage.go generate


## Compiling

    go get golang.org/x/tools/cmd/goimports


## Cgo disaster

    #cgo pkg-config: --define-variable=SRCDIR=${SRCDIR}/../.. ${SRCDIR}/libgodot.pc


## Notes

* Cheatsheet for [cgo data type conversion](https://gist.github.com/zchee/b9c99695463d8902cd33)
* Cross compilation with cgo with [xgo](https://github.com/karalabe/xgo)
* vscode-go patch to [support cgo](https://github.com/golang/go/issues/35721#issuecomment-568543991)
* Check [unsafe pointer conversion](https://blog.gopheracademy.com/advent-2019/safe-use-of-unsafe-pointer/)