//go:build windows

package log

// Link the vendored PE/COFF libbacktrace objects into the extension. The C
// capture code in stacktrace.c calls backtrace_* directly; this import puts
// the package archive on the external link line.
import _ "github.com/godot-go/godot-go/pkg/log/libbacktrace"
