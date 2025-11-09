//go:build !cgo
// +build !cgo

package main

// initNativeLibraries is a no-op when CGO is disabled.
// The IPC-based implementation doesn't require native libraries.
func initNativeLibraries() error {
	// No-op: native libraries only needed for CGO builds
	return nil
}
