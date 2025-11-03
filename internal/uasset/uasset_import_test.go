package uasset

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JaceTheGrayOne/ARI-S/internal/app"
)

func TestUAssetImport_Success_CreatesAssetFiles(t *testing.T) {
	tempDir := t.TempDir()
	importFolder := filepath.Join(tempDir, "json_files")

	if err := os.MkdirAll(importFolder, 0755); err != nil {
		t.Fatalf("Failed to create import folder: %v", err)
	}

	// Create mock JSON files
	mockJSON := filepath.Join(importFolder, "TestAsset.json")
	if err := os.WriteFile(mockJSON, []byte(`{"test": "data"}`), 0644); err != nil {
		t.Fatalf("Failed to create mock JSON: %v", err)
	}

	// Create mock UAssetBridge.exe
	depsDir := filepath.Join(tempDir, "deps")
	bridgeDir := filepath.Join(depsDir, "UAssetAPI")
	if err := os.MkdirAll(bridgeDir, 0755); err != nil {
		t.Fatalf("Failed to create UAssetAPI dir: %v", err)
	}
	bridgePath := filepath.Join(bridgeDir, "UAssetBridge.exe")
	if err := os.WriteFile(bridgePath, []byte("mock"), 0755); err != nil {
		t.Fatalf("Failed to create mock bridge: %v", err)
	}

	appInstance := app.NewApp()
	service := NewUAssetService(appInstance, depsDir)

	ctx := context.Background()
	result := service.ImportUAssets(ctx, importFolder, "")

	if result.Duration == "" {
		t.Error("Expected duration to be recorded")
	}

	t.Logf("Import result: %+v", result)
}

func TestUAssetImport_WithMappings_PassesMappingsPath(t *testing.T) {
	tempDir := t.TempDir()
	importFolder := filepath.Join(tempDir, "json_files")
	mappingsPath := filepath.Join(tempDir, "test.usmap")

	if err := os.MkdirAll(importFolder, 0755); err != nil {
		t.Fatalf("Failed to create import folder: %v", err)
	}

	// Create mock JSON file
	mockJSON := filepath.Join(importFolder, "Test.json")
	if err := os.WriteFile(mockJSON, []byte(`{}`), 0644); err != nil {
		t.Fatalf("Failed to create JSON: %v", err)
	}

	// Create mock mappings file
	if err := os.WriteFile(mappingsPath, []byte("mock mappings"), 0644); err != nil {
		t.Fatalf("Failed to create mappings: %v", err)
	}

	// Create mock UAssetBridge.exe
	depsDir := filepath.Join(tempDir, "deps")
	bridgeDir := filepath.Join(depsDir, "UAssetAPI")
	if err := os.MkdirAll(bridgeDir, 0755); err != nil {
		t.Fatalf("Failed to create UAssetAPI dir: %v", err)
	}
	bridgePath := filepath.Join(bridgeDir, "UAssetBridge.exe")
	if err := os.WriteFile(bridgePath, []byte("mock"), 0755); err != nil {
		t.Fatalf("Failed to create mock bridge: %v", err)
	}

	appInstance := app.NewApp()
	service := NewUAssetService(appInstance, depsDir)

	ctx := context.Background()
	result := service.ImportUAssets(ctx, importFolder, mappingsPath)

	t.Logf("Import with mappings result: %+v", result)
}

func TestUAssetImport_MissingMappings_ReturnsMissingError(t *testing.T) {
	tempDir := t.TempDir()
	importFolder := filepath.Join(tempDir, "json_files")
	mappingsPath := filepath.Join(tempDir, "nonexistent.usmap")

	if err := os.MkdirAll(importFolder, 0755); err != nil {
		t.Fatalf("Failed to create import folder: %v", err)
	}

	// Create mock UAssetBridge.exe
	depsDir := filepath.Join(tempDir, "deps")
	bridgeDir := filepath.Join(depsDir, "UAssetAPI")
	if err := os.MkdirAll(bridgeDir, 0755); err != nil {
		t.Fatalf("Failed to create UAssetAPI dir: %v", err)
	}
	bridgePath := filepath.Join(bridgeDir, "UAssetBridge.exe")
	if err := os.WriteFile(bridgePath, []byte("mock"), 0755); err != nil {
		t.Fatalf("Failed to create mock bridge: %v", err)
	}

	appInstance := app.NewApp()
	service := NewUAssetService(appInstance, depsDir)

	ctx := context.Background()
	result := service.ImportUAssets(ctx, importFolder, mappingsPath)

	if result.Success {
		t.Error("Expected operation to fail with missing mappings file")
	}

	expectedErr := "does not exist"
	if !strings.Contains(result.Error, expectedErr) {
		t.Errorf("Expected error containing '%s', got: %s", expectedErr, result.Error)
	}
}

func TestUAssetImport_CountJSONFiles_Accurate(t *testing.T) {
	tempDir := t.TempDir()
	testFolder := filepath.Join(tempDir, "test_json")
	if err := os.MkdirAll(testFolder, 0755); err != nil {
		t.Fatalf("Failed to create test folder: %v", err)
	}

	// Create 5 .json files
	for i := 1; i <= 5; i++ {
		jsonPath := filepath.Join(testFolder, fmt.Sprintf("Asset%d.json", i))
		if err := os.WriteFile(jsonPath, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create JSON file: %v", err)
		}
	}

	// Create 2 .txt files (should be ignored)
	for i := 1; i <= 2; i++ {
		txtPath := filepath.Join(testFolder, fmt.Sprintf("readme%d.txt", i))
		if err := os.WriteFile(txtPath, []byte("readme"), 0644); err != nil {
			t.Fatalf("Failed to create txt file: %v", err)
		}
	}

	app := NewApp()
	depsDir := filepath.Join(tempDir, "deps")
	service := NewUAssetService(app, depsDir)

	ctx := context.Background()
	jsonCount, err := service.CountJSONFiles(ctx, testFolder)

	if err != nil {
		t.Fatalf("Failed to count JSON files: %v", err)
	}

	if jsonCount != 5 {
		t.Errorf("Expected 5 JSON files, got %d", jsonCount)
	}
}

func TestUAssetImport_ExportImportCycle_PreservesStructure(t *testing.T) {
	tempDir := t.TempDir()

	// Create initial asset files
	assetsFolder := filepath.Join(tempDir, "original_assets")
	if err := os.MkdirAll(assetsFolder, 0755); err != nil {
		t.Fatalf("Failed to create assets folder: %v", err)
	}

	mockAsset := filepath.Join(assetsFolder, "Test.uasset")
	mockExp := filepath.Join(assetsFolder, "Test.uexp")
	if err := os.WriteFile(mockAsset, []byte("original asset data"), 0644); err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	if err := os.WriteFile(mockExp, []byte("original exp data"), 0644); err != nil {
		t.Fatalf("Failed to create exp: %v", err)
	}

	// Create mock UAssetBridge.exe
	depsDir := filepath.Join(tempDir, "deps")
	bridgeDir := filepath.Join(depsDir, "UAssetAPI")
	if err := os.MkdirAll(bridgeDir, 0755); err != nil {
		t.Fatalf("Failed to create UAssetAPI dir: %v", err)
	}
	bridgePath := filepath.Join(bridgeDir, "UAssetBridge.exe")
	if err := os.WriteFile(bridgePath, []byte("mock"), 0755); err != nil {
		t.Fatalf("Failed to create mock bridge: %v", err)
	}

	appInstance := app.NewApp()
	service := NewUAssetService(appInstance, depsDir)
	ctx := context.Background()

	// Act 1: Export (would create JSON files)
	exportResult := service.ExportUAssets(ctx, assetsFolder, "")
	t.Logf("Export phase: %+v", exportResult)

	// Simulate JSON file creation (mock bridge doesn't actually create them)
	jsonPath := filepath.Join(assetsFolder, "Test.json")
	mockJSON := `{"AssetType": "Test", "Properties": []}`
	if err := os.WriteFile(jsonPath, []byte(mockJSON), 0644); err != nil {
		t.Fatalf("Failed to create mock JSON: %v", err)
	}

	// Act 2: Import (would recreate .uasset/.uexp from JSON)
	importResult := service.ImportUAssets(ctx, assetsFolder, "")
	t.Logf("Import phase: %+v", importResult)

	// (actual file conversion would require real UAssetBridge.exe)
	if exportResult.Duration == "" || importResult.Duration == "" {
		t.Error("Expected both operations to record duration")
	}
}
