# Development - Visual Studio Code

## Setup the Launcher

To debug with `gdb` with `vscode`, it's important to setup the proper `pathSubstitutions` mappings for breakpoints to work. Here's an example `launch.json`:
```
{
    "version": "0.2.0",
    "configurations": [
        {
            "type": "gdb",
            "request": "launch",
            "name": "Godot",
            "target": "__YOUR_GODOT_BIN_PATH__/godot",
            "cwd": "${workspaceFolder:godot-go}",
            "valuesFormatting": "parseText",
            "arguments": "--headless --verbose --path test/demo/ --quit",
            "env": {
                "CI": "1",
                "LOG_LEVEL": "info",
                "GOTRACEBACK": "1",
                "GODEBUG": "sbrk=1,gctrace=1,asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=5",
            },
            "printCalls": true,
            "stopAtEntry": true,
            "pathSubstitutions": {
                "/_/github.com/godot-go/godot-go/": "__YOUR_WORKSPACE_FOLDER__",
                "github.com/godot-go/godot-go/": "__YOUR_WORKSPACE_FOLDER__",
                "/_/GOROOT": "__YOUR_GOROOT__"
            }
        }
    ]
}
```

### Debug GODEBUG Flags

The launcher's `GODEBUG` is deliberately not the set `make test` uses:

- `sbrk=1` grows the Go heap from the program break instead of the runtime's
  reserved `mmap` arena, so the heap shows up as ordinary process memory in a
  `gdb` session (the runtime also skips its mmap hint generation in this mode).
  `sbrk` is implemented on Linux only; other platforms silently fall back to
  `mmap`. The headless test run leaves it out because nothing there inspects the
  heap.
- `cgocheck=0` relaxes the cgo pointer-passing checks that the test run keeps on
  with `cgocheck=1`.
- `asyncpreemptoff=1` keeps the runtime from preempting goroutines at
  signal-driven safe points. The test run dropped it once async preemption was
  validated to coexist with Godot; a debug session keeps it for deterministic
  stop points.
- `gctrace=1` prints GC cycles to stderr and `tracebackancestors=5` retains the
  last five goroutine creation stacks per traceback; both are diagnostics only.

### Path Substitution

The path substitions entries are required if you want your breakpoints to work in vscode. The paths can be determined by inspecting the compiled library files:

```
$ readelf -wi test/demo/lib/libgodotgo-test-linux-amd64.so | grep '/_/github.com'
    <1daddd4>   DW_AT_comp_dir    : (indirect line string, offset: 0x340): /_/github.com/godot-go/godot-go/pkg/builtin
    <1dafe92>   DW_AT_comp_dir    : (indirect line string, offset: 0x56d): /_/github.com/godot-go/godot-go/pkg/core
    <1db064b>   DW_AT_comp_dir    : (indirect line string, offset: 0x56d): /_/github.com/godot-go/godot-go/pkg/core
    <1db07ca>   DW_AT_comp_dir    : (indirect line string, offset: 0x56d): /_/github.com/godot-go/godot-go/pkg/core
    <1dbded0>   DW_AT_comp_dir    : (indirect line string, offset: 0x6e2): /_/github.com/godot-go/godot-go/pkg/ffi
    <1dc5fa1>   DW_AT_comp_dir    : (indirect line string, offset: 0x7c0): /_/github.com/godot-go/godot-go/pkg/log
    <1dc93f0>   DW_AT_comp_dir    : (indirect line string, offset: 0xb15): /_/github.com/godot-go/godot-go/pkg/gdclassinit
```