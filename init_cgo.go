//go:build cgo
// +build cgo

package main

import (
	"embed"
	"log"

	"github.com/JaceTheGrayOne/ARI-S/internal/app"
)

// nativeLibsFS holds the embedded native libraries (UAssetBridge.dll).
// This directory is extracted on first run to AppData for native AOT integration.
// Only included when building with CGO enabled.
//
//go:embed nativelibs
var nativeLibsFS embed.FS

// initNativeLibraries extracts embedded native libraries when CGO is enabled.
// This is called early in main() before any services are created.
func initNativeLibraries() error {
	extractedLibsDir, err := app.EnsureNativeLibraries(nativeLibsFS)
	if err != nil {
		return err
	}
	log.Printf("Native libraries directory: %s", extractedLibsDir)
	return nil
}
