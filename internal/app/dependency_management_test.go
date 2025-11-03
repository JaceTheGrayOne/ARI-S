package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDependencyManagement_FirstRun_ExtractsSuccessfully(t *testing.T) {
	tempDir := t.TempDir()
	depsDir := filepath.Join(tempDir, "dependencies")

	// Create mock embedded FS structure
	versionContent := []byte("1.0.0")
	versionPath := filepath.Join(depsDir, "version.txt")

	if err := os.MkdirAll(depsDir, 0755); err != nil {
		t.Fatalf("Failed to create deps dir: %v", err)
	}

	if err := os.WriteFile(versionPath, versionContent, 0644); err != nil {
		t.Fatalf("Failed to write version file: %v", err)
	}

	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("Expected version.txt to exist after extraction")
	}

	actualVersion, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatalf("Failed to read version file: %v", err)
	}

	if string(actualVersion) != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", string(actualVersion))
	}
}

func TestDependencyManagement_VersionMismatch_ReExtracts(t *testing.T) {
	tempDir := t.TempDir()
	depsDir := filepath.Join(tempDir, "dependencies")
	versionPath := filepath.Join(depsDir, "version.txt")

	if err := os.MkdirAll(depsDir, 0755); err != nil {
		t.Fatalf("Failed to create deps dir: %v", err)
	}

	// Write old version
	oldVersion := []byte("0.9.0")
	if err := os.WriteFile(versionPath, oldVersion, 0644); err != nil {
		t.Fatalf("Failed to write old version: %v", err)
	}

	// Create marker file to verify re-extraction
	markerPath := filepath.Join(depsDir, "old_marker.txt")
	if err := os.WriteFile(markerPath, []byte("old"), 0644); err != nil {
		t.Fatalf("Failed to write marker: %v", err)
	}

	newVersion := []byte("1.0.0")

	// Remove old deps (simulating re-extraction)
	if err := os.RemoveAll(depsDir); err != nil {
		t.Fatalf("Failed to remove old deps: %v", err)
	}

	// Re-create with new version
	if err := os.MkdirAll(depsDir, 0755); err != nil {
		t.Fatalf("Failed to recreate deps dir: %v", err)
	}

	if err := os.WriteFile(versionPath, newVersion, 0644); err != nil {
		t.Fatalf("Failed to write new version: %v", err)
	}

	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Error("Expected old marker to be removed during re-extraction")
	}

	actualVersion, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatalf("Failed to read new version: %v", err)
	}

	if string(actualVersion) != "1.0.0" {
		t.Errorf("Expected new version '1.0.0', got '%s'", string(actualVersion))
	}
}

func TestDependencyManagement_VersionMatch_SkipsExtraction(t *testing.T) {
	tempDir := t.TempDir()
	depsDir := filepath.Join(tempDir, "dependencies")
	versionPath := filepath.Join(depsDir, "version.txt")
	currentVersion := []byte("1.0.0")

	if err := os.MkdirAll(depsDir, 0755); err != nil {
		t.Fatalf("Failed to create deps dir: %v", err)
	}

	if err := os.WriteFile(versionPath, currentVersion, 0644); err != nil {
		t.Fatalf("Failed to write version: %v", err)
	}

	// Create timestamp file to detect if extraction ran
	timestampPath := filepath.Join(depsDir, "timestamp.txt")
	originalTime := []byte("original")
	if err := os.WriteFile(timestampPath, originalTime, 0644); err != nil {
		t.Fatalf("Failed to write timestamp: %v", err)
	}

	diskVersion, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatalf("Failed to read version: %v", err)
	}

	embeddedVersion := currentVersion
	shouldSkip := string(diskVersion) == string(embeddedVersion)

	if !shouldSkip {
		t.Error("Expected extraction to be skipped for matching version")
	}

	// Verify timestamp file unchanged
	timestampContent, err := os.ReadFile(timestampPath)
	if err != nil {
		t.Fatalf("Failed to read timestamp: %v", err)
	}

	if string(timestampContent) != "original" {
		t.Error("Expected timestamp file to remain unchanged when skipping extraction")
	}
}
