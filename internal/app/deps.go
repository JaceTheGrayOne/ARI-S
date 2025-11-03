package app

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsureDependencies extracts embedded dependencies on first run and returns the extraction path.
// Automatically re-extracts if version.txt mismatches.
func EnsureDependencies(depsFS embed.FS) (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not get executable path: %w", err)
	}
	appDir := filepath.Dir(exePath)
	depsDir := filepath.Join(appDir, "dependencies")

	// Check version to determine if we need to extract/re-extract
	versionPath := filepath.Join(depsDir, "version.txt")
	needsExtraction := false

	// Read embedded version
	embeddedVersion, err := depsFS.ReadFile("dependencies/version.txt")
	if err != nil {
		log.Printf("Warning: Could not read embedded version.txt: %v", err)
		embeddedVersion = []byte("unknown")
	}

	// Check if version file exists on disk
	if diskVersion, err := os.ReadFile(versionPath); err == nil {
		// Version file exists, compare versions
		if string(diskVersion) != string(embeddedVersion) {
			log.Printf("Version mismatch: disk=%s, embedded=%s. Re-extracting dependencies...",
				string(diskVersion), string(embeddedVersion))
			needsExtraction = true
			// Remove old dependencies directory
			if err := os.RemoveAll(depsDir); err != nil {
				return "", fmt.Errorf("failed to remove old dependencies: %w", err)
			}
		} else {
			log.Println("Dependencies version matches. Skipping extraction.")
			return depsDir, nil
		}
	} else if os.IsNotExist(err) {
		// Version file doesn't exist, need to extract
		log.Println("First run: extracting dependencies...")
		needsExtraction = true
	} else {
		// Some other error occurred
		return "", fmt.Errorf("failed to check for dependencies: %w", err)
	}

	if needsExtraction {
		// Create the target directory
		if err := os.MkdirAll(depsDir, 0755); err != nil {
			return "", fmt.Errorf("could not create dependencies directory: %w", err)
		}

		// Call the recursive extractor
		if err := extractFS(depsFS, "dependencies", depsDir); err != nil {
			return "", fmt.Errorf("failed to extract dependencies: %w", err)
		}

		log.Println("Dependencies extracted successfully.")
	}

	return depsDir, nil
}

func isExecutable(path string) bool {
	// On Windows, .exe files are executable
	if runtime.GOOS == "windows" {
		return strings.HasSuffix(strings.ToLower(path), ".exe")
	}

	// On Unix-like systems, check for known executables
	base := filepath.Base(path)
	baseLower := strings.ToLower(base)

	// Add known executables here (without .exe extension for cross-platform builds)
	knownExecutables := []string{
		"retoc",
		"uassetbridge",
		"createdump",
	}

	for _, exe := range knownExecutables {
		if baseLower == exe {
			return true
		}
	}

	return false
}

func extractFS(efs embed.FS, embedRoot, destDir string) error {
	return fs.WalkDir(efs, embedRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == embedRoot {
			return nil
		}

		// Calculate relative path from embed root
		relPath, err := filepath.Rel(embedRoot, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			// Create directory with appropriate permissions
			return os.MkdirAll(destPath, 0755)
		}

		// Read the embedded file's content
		data, err := efs.ReadFile(path)
		if err != nil {
			return err
		}

		// Determine file permissions
		perm := fs.FileMode(0644) // Default: readable/writable by owner, readable by others
		if isExecutable(path) {
			perm = 0755 // Executable: readable/executable by all, writable by owner
		}

		// Write the file to disk
		if err := os.WriteFile(destPath, data, perm); err != nil {
			return err
		}

		// On Unix-like systems, we may need to explicitly set permissions
		// even after WriteFile, especially for executables
		if runtime.GOOS != "windows" && isExecutable(path) {
			if err := os.Chmod(destPath, 0755); err != nil {
				log.Printf("Warning: Failed to set executable permissions on %s: %v", destPath, err)
			}
		}

		return nil
	})
}
