# Vendored libbacktrace

This directory contains the PE/COFF-only libbacktrace sources that Godot
4.7-stable vendors at `thirdparty/libbacktrace`. godot-go uses them on
Windows to resolve frames in the extension DLL from its embedded DWARF
debug information, matching the MinGW path in Godot's
`platform/windows/crash_handler_windows_signal.cpp`.

- Upstream: Godot 4.7-stable (`5b4e0cb0fd279832bbdd69fed5354d4e5ad26f88`),
  `thirdparty/libbacktrace`.
- Copied files: `alloc.c`, `atomic.c`, `backtrace.h`,
  `backtrace-supported.h`, `config.h`, `dwarf.c`, `fileline.c`,
  `filenames.h`, `internal.h`, `pecoff.c`, `posix.c`, `read.c`,
  `sort.c`, `state.c`, and `LICENSE`.
- Omitted: `backtrace.c`, `simple.c`, and `print.c`. They provide the
  unwind-based `backtrace_full`/`backtrace_simple` entry points and the
  printing wrapper; this project calls only `backtrace_pcinfo`, and
  linking the unwind entry points would add an unnecessary
  `_Unwind_Backtrace`/`libunwind.dll` dependency.
- Godot's `patches/0001-big-files-support.patch` is already applied to
  `read.c`; it switches the reader to `_lseeki64`/`_read` for large PE
  files.
- `config.h` is Godot's generated Windows configuration with
  `HAVE_ZLIB` and `HAVE_LIBLZMA` set to `0`: the build does not need
  compressed debug sections, and avoiding those features keeps zlib/lzma
  out of the link.

The sources are Windows-only in this project. `read.c` uses `_lseeki64`
and `_read`, and the package is imported from a Windows-only file, so
Linux/macOS builds continue to use `execinfo`.

libbacktrace is distributed under the BSD 3-Clause license in `LICENSE`.
