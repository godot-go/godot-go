package log

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/
#cgo windows LDFLAGS: -ldbghelp
// CodeView records for this package are enabled on windows so dbghelp can
// print file:line for capture frames from the linked PDB. The directive is
// deliberately per-package: the build's global CGO_CFLAGS stay on DWARF
// because dwarf2codeview (MinGW-Builds GCC 15.2.0) ICEs on the generated
// pkg/ffi TUs and on every pkg/core cgo preamble TU (design D6). See
// openspec/changes/windows-pdb-line-info/.
#cgo windows CFLAGS: -gcodeview
*/
import "C"
