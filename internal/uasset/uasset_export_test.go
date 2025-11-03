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

func TestUAssetExport_Success_CreatesJSONFiles(t *testing.T) {
	tempDir := t.TempDir()
	exportFolder := filepath.Join(tempDir, "uassets")

	if err := os.MkdirAll(exportFolder, 0755); err != nil {
		t.Fatalf("Failed to create export folder: %v", err)
	}

	// Create mock .uasset and .uexp files
	mockAsset := filepath.Join(exportFolder, "TestAsset.uasset")
	mockExp := filepath.Join(exportFolder, "TestAsset.uexp")
	if err := os.WriteFile(mockAsset, []byte("mock asset"), 0644); err != nil {
		t.Fatalf("Failed to create mock uasset: %v", err)
	}
	if err := os.WriteFile(mockExp, []byte("mock exp"), 0644); err != nil {
		t.Fatalf("Failed to create mock uexp: %v", err)
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
	result := service.ExportUAssets(ctx, exportFolder, "")

	if result.Duration == "" {
		t.Error("Expected duration to be recorded")
	}

	t.Logf("Export result: %+v", result)
}

func TestUAssetExport_WithMappings_PassesMappingsPath(t *testing.T) {
	tempDir := t.TempDir()
	exportFolder := filepath.Join(tempDir, "uassets")
	mappingsPath := filepath.Join(tempDir, "test.usmap")

	if err := os.MkdirAll(exportFolder, 0755); err != nil {
		t.Fatalf("Failed to create export folder: %v", err)
	}

	// Create mock mappings file
	if err := os.WriteFile(mappingsPath, []byte("mock mappings"), 0644); err != nil {
		t.Fatalf("Failed to create mappings file: %v", err)
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
	result := service.ExportUAssets(ctx, exportFolder, mappingsPath)

	t.Logf("Export with mappings result: %+v", result)
}

func TestUAssetExport_MissingBridge_ReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	depsDir := filepath.Join(tempDir, "nonexistent_deps")

	appInstance := app.NewApp()
	service := NewUAssetService(appInstance, depsDir)

	ctx := context.Background()
	result := service.ExportUAssets(ctx, "/test/folder", "")

	if result.Success {
		t.Error("Expected operation to fail with missing bridge")
	}

	expectedErr := "UAssetBridge.exe not found"
	if !strings.Contains(result.Error, expectedErr) {
		t.Errorf("Expected error containing '%s', got: %s", expectedErr, result.Error)
	}
}

func TestUAssetExport_MissingFolder_ReturnsError(t *testing.T) {
	tempDir := t.TempDir()
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

	nonExistentFolder := filepath.Join(tempDir, "nonexistent")

	ctx := context.Background()
	result := service.ExportUAssets(ctx, nonExistentFolder, "")

	if result.Success {
		t.Error("Expected operation to fail with missing folder")
	}

	expectedErr := "does not exist"
	if !strings.Contains(result.Error, expectedErr) {
		t.Errorf("Expected error containing '%s', got: %s", expectedErr, result.Error)
	}
}

func TestUAssetExport_CountFiles_Accurate(t *testing.T) {
	tempDir := t.TempDir()
	testFolder := filepath.Join(tempDir, "test_assets")
	if err := os.MkdirAll(testFolder, 0755); err != nil {
		t.Fatalf("Failed to create test folder: %v", err)
	}

	// Create 3 .uasset and 3 .uexp files
	for i := 1; i <= 3; i++ {
		assetPath := filepath.Join(testFolder, fmt.Sprintf("Asset%d.uasset", i))
		expPath := filepath.Join(testFolder, fmt.Sprintf("Asset%d.uexp", i))

		if err := os.WriteFile(assetPath, []byte("mock"), 0644); err != nil {
			t.Fatalf("Failed to create mock file: %v", err)
		}
		if err := os.WriteFile(expPath, []byte("mock"), 0644); err != nil {
			t.Fatalf("Failed to create mock file: %v", err)
		}
	}

	// Create 1 .txt file (should be ignored)
	txtPath := filepath.Join(testFolder, "readme.txt")
	if err := os.WriteFile(txtPath, []byte("readme"), 0644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}

	app := NewApp()
	depsDir := filepath.Join(tempDir, "deps")
	service := NewUAssetService(app, depsDir)

	ctx := context.Background()
	uassetCount, uexpCount, err := service.CountUAssetFiles(ctx, testFolder)

	if err != nil {
		t.Fatalf("Failed to count files: %v", err)
	}

	if uassetCount != 3 {
		t.Errorf("Expected 3 .uasset files, got %d", uassetCount)
	}

	if uexpCount != 3 {
		t.Errorf("Expected 3 .uexp files, got %d", uexpCount)
	}
}
