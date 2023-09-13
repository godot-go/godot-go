[![Build Status](https://github.com/godot-go/godot-go/actions/workflows/ci_linux.yaml/badge.svg)](https://github.com/godot-go/godot-go/actions/workflows/ci_linux.yaml)

---

# godot-go: Go bindings for Godot 4.2-dev

[Go](https://golang.org/) bindings for the [Godot Game Engine](https://github.com/godotengine/godot) cross-platform game engine. godot-go integrates into Godot through the Godot GDExtension API through cgo.

The project is currently under heavy development. The API should be considered __EXPERIMENTAL__ and is subject to change. The API is expected to become more stable as we get closer to a 1.0 release.

## Current State of the Project

Although the tests confirm positive results, the godot-go bindings are currently not useable until [Add functions for non-ptr style virtual calls in GDExtension](https://github.com/godotengine/godot/pull/80671) is merged. Further development won't make much sense until this is addressed.

## Getting Started

Requirements:
* clang-format
* gcc
* go 1.21.x

TODO


### Building Godot-Go

In order for godot-go to work correctly, you must make sure the godot_headers are in sync between your godot binary and godot-go. Development is built and tested off of [Godot 4.2 dev](https://github.com/fuzzybinary/godot/tree/gdextension-virtuals). There are breaking changes from the Godot 4.0 GDExtension interface. Please make sure to run godot-go against this fork of Godot 4.2-dev as it contains important virtual call changes.


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

This is the expected output:

```
$ make test
CI=1 \
LOG_LEVEL=info \
GOTRACEBACK=1 \
GODEBUG=sbrk=1,gctrace=1,asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=5 \
/home/pcting/bin/godot --headless --path test/demo/ --quit
INFO    gdextension/godot.go:59 godot version   {"major": 4, "minor": 2}
Godot Engine v4.2.dev.custom_build.53553f07d - https://godotengine.org

INFO    gdextension/classdb.go:422      gdclass registered      {"class": "ExampleRef", "parent_type": "RefCounted"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:ExampleRef.GetId() int32"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:ExampleRef.SetId(int32)"}
INFO    gdextension/classdb.go:422      gdclass registered      {"class": "Example", "parent_type": "Control"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.V_Ready()"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.V_Input(Ref)"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.V_Set(string,Variant) bool"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.V_Get(string) Variant"}
INFO    gdextension/method_bind.go:314  Create Variadic ClassMethodInfoFromMethodBind   {"bind": "MethodBind:Example.VarargsFunc() Variant"}
INFO    gdextension/method_bind.go:314  Create Variadic ClassMethodInfoFromMethodBind   {"bind": "MethodBind:Example.VarargsFuncVoid()"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.SimpleFunc()"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.SimpleConstFunc(int64)"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.ReturnSomething(string,float32,float64,int,int8,int16,int32,int64) string"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.ReturnSomethingConst() Viewport"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.GetV4() Vector4"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.DefArgs(int32,int32) int32"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.TestArray() Array"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.TestDictionary() Dictionary"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.TestStringOps() string"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.GetCustomPosition() Vector2"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.SetCustomPosition(Vector2)"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.GetPropertyFromList() Vector3"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.SetPropertyFromList(Vector3)"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.EmitCustomSignal(string,int64)"}
INFO    gdextension/method_bind.go:333  Create Normal ClassMethodInfoFromMethodBind     {"bind": "MethodBind:Example.TestCastTo()"}
INFO    gdextension/wrapped_gdclass.go:174      GDClass instance created        {"object_id": 24528291065, "class_name": "Example", "parent_name": "Control", "inst": "0x7fa7ec19ca10", "owner": "0xcfe5db0", "object": "0xcfe5db0", "inst.GetGodotObjectOwner": "0xcfe5db0"}
INFO    gdextension/classdb_callback.go:45      GoCallback_ClassCreationInfoGetVirtualCallWithData called       {"class_name_from_user_data": "Example", "method_name": "_get_minimum_size"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_get_minimum_size"}
INFO    gdextension/classdb_callback.go:75      no virtual method found {"className": "Example", "method": "_get_minimum_size"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_get_minimum_size"}
INFO    gdextension/classdb_callback.go:75      no virtual method found {"className": "Example", "method": "_get_minimum_size"}
INFO    gdextension/classdb_callback.go:45      GoCallback_ClassCreationInfoGetVirtualCallWithData called       {"class_name_from_user_data": "Example", "method_name": "_enter_tree"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_enter_tree"}
INFO    gdextension/classdb_callback.go:75      no virtual method found {"className": "Example", "method": "_enter_tree"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_get_minimum_size"}
INFO    gdextension/classdb_callback.go:75      no virtual method found {"className": "Example", "method": "_get_minimum_size"}
INFO    gdextension/classdb_callback.go:45      GoCallback_ClassCreationInfoGetVirtualCallWithData called       {"class_name_from_user_data": "Example", "method_name": "_input"}
INFO    gdextension/classdb_callback.go:45      GoCallback_ClassCreationInfoGetVirtualCallWithData called       {"class_name_from_user_data": "Example", "method_name": "_shortcut_input"}
INFO    gdextension/classdb_callback.go:45      GoCallback_ClassCreationInfoGetVirtualCallWithData called       {"class_name_from_user_data": "Example", "method_name": "_unhandled_input"}
INFO    gdextension/classdb_callback.go:45      GoCallback_ClassCreationInfoGetVirtualCallWithData called       {"class_name_from_user_data": "Example", "method_name": "_unhandled_key_input"}
INFO    gdextension/classdb_callback.go:45      GoCallback_ClassCreationInfoGetVirtualCallWithData called       {"class_name_from_user_data": "Example", "method_name": "_process"}
INFO    gdextension/classdb_callback.go:45      GoCallback_ClassCreationInfoGetVirtualCallWithData called       {"class_name_from_user_data": "Example", "method_name": "_physics_process"}
INFO    gdextension/classdb_callback.go:45      GoCallback_ClassCreationInfoGetVirtualCallWithData called       {"class_name_from_user_data": "Example", "method_name": "_ready"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_ready"}
INFO    pkg/example.go:302      Example_Ready called    {"inst": "0x7fa7ec19ca10"}
INFO    pkg/example.go:307      Vector3: Created (1.1, 2.2, 3.3)        {"x": 1.100000023841858, "y": 2.200000047683716, "z": 3.299999952316284}
INFO    pkg/example.go:313      Vector3: Multiply Vector3 by 2  {"x": 2.200000047683716, "y": 4.400000095367432, "z": 6.599999904632568}
INFO    pkg/example.go:319      Vector3: Add (1,2,3)    {"x": 12.199999809265137, "y": 24.399999618530273, "z": 36.599998474121094}
INFO    pkg/example.go:325      Vector3: Multiply (5,10,15)     {"x": 61, "y": 244, "z": 549}
INFO    pkg/example.go:331      Vector3: Substract (x,y,0)      {"x": 0, "y": 0, "z": 549}
INFO    pkg/example.go:337      Vector3: Normalized     {"x": 0, "y": 0, "z": 1}
INFO    pkg/example.go:343      Vector3: Equality Check {"x": 0, "y": 0, "z": 1, "equal": true}
INFO    gdextension/method_bind.go:241  Ptrcall {"bind": "MethodBind:Example.V_Ready()", "ret": "()"}
INFO    pkg/example.go:181      EmitCustomSignal called {"name": "Button", "value": 42}
INFO    gdextension/method_bind.go:241  Ptrcall {"bind": "MethodBind:Example.EmitCustomSignal(string,int64)", "ret": "()"}
INFO    gdextension/classdb_callback.go:208     GoCallback_ClassCreationInfoSet called  {"class": "Example", "name": "property_from_list", "value": "(100, 200, 300)"}
INFO    gdextension/classdb_callback.go:232     reflect method called   {"ret": "(bool(<bool Value>))"}
INFO    gdextension/classdb_callback.go:152     GoCallback_ClassCreationInfoGet called  {"class": "Example", "method_name": "property_from_list"}
INFO    gdextension/classdb_callback.go:188     reflect method called   {"ret": "(Variant(<gdextension.Variant Value>),bool(<bool Value>))", "v": "(100, 200, 300)"}
  Simple func called.
INFO    pkg/example.go:181      EmitCustomSignal called {"name": "simple_func", "value": 3}
INFO    gdextension/method_bind.go:241  Ptrcall {"bind": "MethodBind:Example.SimpleFunc()", "ret": "()"}
  Simple const func called 123.
INFO    pkg/example.go:181      EmitCustomSignal called {"name": "simple_const_func", "value": 4}
INFO    gdextension/method_bind.go:241  Ptrcall {"bind": "MethodBind:Example.SimpleConstFunc(int64)", "ret": "()"}
  Return something called (8 values cancatenated as a string).
INFO    gdextension/method_bind.go:241  Ptrcall {"bind": "MethodBind:Example.ReturnSomething(string,float32,float64,int,int8,int16,int32,int64) string", "ret": "(string(1. some string42, 2. 1.166667, 3. 1166.666667, 4. 2147483647, 5. -127, 6. -32768, 7. 2147483647, 8. 9223372036854775807))"}
  Return something const called.
viewport instance id: 23655874846
INFO    gdextension/method_bind.go:241  Ptrcall {"bind": "MethodBind:Example.ReturnSomethingConst() Viewport", "ret": "(Viewport(<gdextension.Viewport Value>))"}
INFO    gdextension/method_bind.go:186  Call Variadic   {"bind": "MethodBind:Example.VarargsFunc() Variant", "gd_args": "[]Variant(some, arguments, to, test)", "resolved_args": "[]Variant(some)", "ret": "(Variant(<gdextension.Variant Value>))"}
INFO    gdextension/method_bind.go:186  Call Variadic   {"bind": "MethodBind:Example.VarargsFunc() Variant", "gd_args": "[]Variant(some)", "resolved_args": "[]Variant(some)", "ret": "(Variant(<gdextension.Variant Value>))"}
INFO    pkg/example.go:181      EmitCustomSignal called {"name": "varargs_func_void", "value": 5}
INFO    gdextension/method_bind.go:186  Call Variadic   {"bind": "MethodBind:Example.VarargsFuncVoid()", "gd_args": "[]Variant(some, arguments, to, test)", "resolved_args": "[]Variant(some)", "ret": "()"}
INFO    pkg/example.go:91       DefArgs called  {"sum": 300}
INFO    gdextension/method_bind.go:208  Call    {"bind": "MethodBind:Example.DefArgs(int32,int32) int32", "gd_args": "[]Variant()", "resolved_args": "[]Variant(100, 200)", "ret": "(int32(<int32 Value>))"}
INFO    pkg/example.go:91       DefArgs called  {"sum": 250}
INFO    gdextension/method_bind.go:208  Call    {"bind": "MethodBind:Example.DefArgs(int32,int32) int32", "gd_args": "[]Variant(50)", "resolved_args": "[]Variant(50, 200)", "ret": "(int32(<int32 Value>))"}
INFO    pkg/example.go:91       DefArgs called  {"sum": 150}
INFO    gdextension/method_bind.go:241  Ptrcall {"bind": "MethodBind:Example.DefArgs(int32,int32) int32", "ret": "(int32(<int32 Value>))"}
INFO    pkg/example.go:114      arr size        {"size": 2, "v[0]": "1", "v[1]": "2"}
INFO    pkg/example.go:145      pick random     {"val": 2}
INFO    gdextension/method_bind.go:241  Ptrcall {"bind": "MethodBind:Example.TestArray() Array", "ret": "(Array(<gdextension.Array Value>))"}
INFO    gdextension/method_bind.go:241  Ptrcall {"bind": "MethodBind:Example.TestDictionary() Dictionary", "ret": "(Dictionary(<gdextension.Dictionary Value>))"}
INFO    gdextension/char_string.go:136  decoded utf32   {"str": "ABCĎE\u0000"}
INFO    gdextension/method_bind.go:241  Ptrcall {"bind": "MethodBind:Example.TestStringOps() string", "ret": "(string(ABCĎE\u0000))"}
INFO    gdextension/classdb_callback.go:152     GoCallback_ClassCreationInfoGet called  {"class": "Example", "method_name": "group_subgroup_custom_position"}
INFO    gdextension/method_bind.go:208  Call    {"bind": "MethodBind:Example.GetCustomPosition() Vector2", "gd_args": "[]Variant()", "resolved_args": "[]Variant()", "ret": "(Vector2(<gdextension.Vector2 Value>))"}
INFO    gdextension/classdb_callback.go:208     GoCallback_ClassCreationInfoSet called  {"class": "Example", "name": "group_subgroup_custom_position", "value": "(50, 50)"}
INFO    gdextension/classdb_callback.go:232     reflect method called   {"ret": "(bool(<bool Value>))"}
INFO    gdextension/method_bind.go:208  Call    {"bind": "MethodBind:Example.SetCustomPosition(Vector2)", "gd_args": "[]Variant((50, 50))", "resolved_args": "[]Variant((50, 50))", "ret": "()"}
INFO    gdextension/classdb_callback.go:152     GoCallback_ClassCreationInfoGet called  {"class": "Example", "method_name": "group_subgroup_custom_position"}
INFO    gdextension/method_bind.go:208  Call    {"bind": "MethodBind:Example.GetCustomPosition() Vector2", "gd_args": "[]Variant()", "resolved_args": "[]Variant()", "ret": "(Vector2(<gdextension.Vector2 Value>))"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_input"}
INFO    pkg/example.go:181      EmitCustomSignal called {"name": "_input: H", "value": 72}
INFO    gdextension/method_bind.go:241  Ptrcall {"bind": "MethodBind:Example.V_Input(Ref)", "ret": "()"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_shortcut_input"}
INFO    gdextension/classdb_callback.go:75      no virtual method found {"className": "Example", "method": "_shortcut_input"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_unhandled_key_input"}
INFO    gdextension/classdb_callback.go:75      no virtual method found {"className": "Example", "method": "_unhandled_key_input"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_unhandled_input"}
INFO    gdextension/classdb_callback.go:75      no virtual method found {"className": "Example", "method": "_unhandled_input"}

 ==== TESTS FINISHED ====

   PASSES: 22
   FAILURES: 0

 ******** PASSED ********

INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_physics_process"}
INFO    gdextension/classdb_callback.go:75      no virtual method found {"className": "Example", "method": "_physics_process"}
INFO    gdextension/classdb_callback.go:45      GoCallback_ClassCreationInfoGetVirtualCallWithData called       {"class_name_from_user_data": "Example", "method_name": "_draw"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_draw"}
INFO    gdextension/classdb_callback.go:75      no virtual method found {"className": "Example", "method": "_draw"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_process"}
INFO    gdextension/classdb_callback.go:75      no virtual method found {"className": "Example", "method": "_process"}
INFO    gdextension/classdb_callback.go:45      GoCallback_ClassCreationInfoGetVirtualCallWithData called       {"class_name_from_user_data": "Example", "method_name": "_exit_tree"}
INFO    gdextension/classdb_callback.go:63      GoCallback_ClassCreationInfoCallVirtualWithData called  {"type": "Example", "userData": "Example", "method": "_exit_tree"}
INFO    gdextension/classdb_callback.go:75      no virtual method found {"className": "Example", "method": "_exit_tree"}
INFO    gdextension/classdb_callback.go:108     GoCallback_ClassCreationInfoFreeInstance called {"type_name": "Example", "ptr": "0x7fa7ec19cba8", "to_string": "[ GDExtension::Example <--> Instance ID:24528291065 ]", "GodotObjectOwner()": "0xcfe5db0"}
INFO    gdextension/classdb_callback.go:119     GDClass instance freed  {"id": 24528291065}
```

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
