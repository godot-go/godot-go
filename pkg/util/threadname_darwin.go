//go:build darwin

package util

// SetThreadName is a no-op on macOS, matching the Windows stub: every caller
// uses it for debugging output, and pthread_setname_np would pull cgo into this
// package for that.
func SetThreadName(name string) {}
