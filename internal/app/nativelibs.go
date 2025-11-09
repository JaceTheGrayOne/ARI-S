//go:build cgo
// +build cgo

package app

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

// EnsureNativeLibraries extracts embedded native libraries on first run and returns the extraction path.
// Automatically re-extracts if version.txt mismatches.
// Libraries are extracted to the user's AppData directory alongside config.json.
func EnsureNativeLibraries(nativeLibsFS embed.FS) (string, error) {
	// Use the same appdata directory as config
	appDataDir := getAppDataDir()
	nativeLibsDir := filepath.Join(appDataDir, "nativelibs")

	// Check version to determine if we need to extract/re-extract
	versionPath := filepath.Join(nativeLibsDir, "version.txt")
	needsExtraction := false

	// Read embedded version
	embeddedVersion, err := nativeLibsFS.ReadFile("nativelibs/version.txt")
	if err != nil {
		log.Printf("Warning: Could not read embedded version.txt: %v", err)
		embeddedVersion = []byte("unknown")
	}

	// Check if version file exists on disk
	if diskVersion, err := os.ReadFile(versionPath); err == nil {
		// Version file exists, compare versions
		if string(diskVersion) != string(embeddedVersion) {
			log.Printf("Native library version mismatch: disk=%s, embedded=%s. Re-extracting...",
				string(diskVersion), string(embeddedVersion))
			needsExtraction = true
			// Remove old native libraries directory
			if err := os.RemoveAll(nativeLibsDir); err != nil {
				return "", fmt.Errorf("failed to remove old native libraries: %w", err)
			}
		} else {
			log.Println("Native library version matches. Skipping extraction.")
		}
	} else if os.IsNotExist(err) {
		// Version file doesn't exist, need to extract
		log.Println("First run: extracting native libraries...")
		needsExtraction = true
	} else {
		// Some other error occurred
		return "", fmt.Errorf("failed to check for native libraries: %w", err)
	}

	if needsExtraction {
		// Create the target directory
		if err := os.MkdirAll(nativeLibsDir, 0755); err != nil {
			return "", fmt.Errorf("could not create native libraries directory: %w", err)
		}

		// Call the recursive extractor
		if err := extractFS(nativeLibsFS, "nativelibs", nativeLibsDir); err != nil {
			return "", fmt.Errorf("failed to extract native libraries: %w", err)
		}

		log.Println("Native libraries extracted successfully.")
	}

	// Configure Windows DLL search path to include the extracted directory
	if runtime.GOOS == "windows" {
		if err := addDLLSearchPath(nativeLibsDir); err != nil {
			log.Printf("Warning: Failed to add DLL search path: %v", err)
		} else {
			log.Printf("Added DLL search path: %s", nativeLibsDir)
		}
	}

	return nativeLibsDir, nil
}

// addDLLSearchPath adds a directory to the DLL search path on Windows.
// This allows the runtime linker to find DLLs in the specified directory.
func addDLLSearchPath(dir string) error {
	if runtime.GOOS != "windows" {
		return nil // No-op on non-Windows platforms
	}

	// Convert Go string to UTF-16 for Windows API
	dirPtr, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		return fmt.Errorf("failed to convert path to UTF-16: %w", err)
	}

	// Load kernel32.dll
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return fmt.Errorf("failed to load kernel32.dll: %w", err)
	}
	defer kernel32.Release()

	// Get SetDllDirectory function
	setDllDirectory, err := kernel32.FindProc("SetDllDirectoryW")
	if err != nil {
		return fmt.Errorf("failed to find SetDllDirectoryW: %w", err)
	}

	// Call SetDllDirectoryW
	ret, _, callErr := setDllDirectory.Call(uintptr(unsafe.Pointer(dirPtr)))
	if ret == 0 {
		return fmt.Errorf("SetDllDirectoryW failed: %w", callErr)
	}

	return nil
}
