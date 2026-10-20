//go:build windows

// Copyright (c) 2026 the godot-go authors.

package core

/*
#include <windows.h>
#include <stdint.h>

// Pin this DLL in the process address space. GET_MODULE_HANDLE_EX_FLAG_PIN
// adds a reference that only process exit releases, so the engine's
// FreeLibrary during Main::cleanup cannot unmap the module.
static BOOL godotgo_pin_module(void) {
	HMODULE mod = NULL;
	return GetModuleHandleExA(
		GET_MODULE_HANDLE_EX_FLAG_PIN | GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS,
		(LPCSTR)(uintptr_t)&godotgo_pin_module,
		&mod);
}
*/
import "C"

import (
	"github.com/godot-go/godot-go/pkg/log"
)

func init() {
	// A Go c-shared library cannot be unloaded: the Go runtime keeps its own
	// threads (sysmon, GC workers) running, and its exception handlers live in
	// this module. Godot releases every GDExtension library during shutdown
	// (Main::cleanup -> unregister_core_types -> ~GDExtensionManager ->
	// FreeLibrary); unmapping the module makes the next runtime-thread
	// instruction an access violation that no in-process handler can report.
	//
	// Consequence: the module stays mapped until process exit, so the engine's
	// unload is a no-op on Windows. Hot-reload/unload of the extension is not
	// supported there; restarting the process is the supported path.
	if C.godotgo_pin_module() == 0 {
		log.Warn("failed to pin the extension module; the process may fault at shutdown")
	}
}
