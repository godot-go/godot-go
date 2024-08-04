# Development

## Debugging with VSCode

To debug with `gdb`, you can setup Visual Studio Code `launch.json`:
```
{
    "version": "0.2.0",
    "configurations": [
        {
            "type": "gdb",
            "request": "launch",
            "name": "Godot",
            "target": "/home/pcting/bin/godot",
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