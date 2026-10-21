package log

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/
// The vendored PE/COFF libbacktrace resolves Windows frames from DWARF.
// GCC 15 already defaults to DWARF 5 here, but pin it so a toolchain default
// cannot silently drop back to a version libbacktrace cannot read.
#cgo windows CFLAGS: -I${SRCDIR}/libbacktrace -gdwarf-5
*/
import "C"
