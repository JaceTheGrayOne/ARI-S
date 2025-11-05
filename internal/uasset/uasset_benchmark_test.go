//go:build cgo
// +build cgo

package uasset

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JaceTheGrayOne/ARI-S/internal/app"
)

// BenchmarkUAssetExport_IPC benchmarks the old IPC-based export.
// This requires the UAssetBridge.exe to be present.
func BenchmarkUAssetExport_IPC(b *testing.B) {
	// Setup: Create test directory with a sample asset
	testDir := setupBenchmarkAssets(b)
	defer os.RemoveAll(testDir)

	// Create IPC service (old method)
	appInstance := app.NewApp()
	depsDir := findDepsDir(b)
	service := NewUAssetService(appInstance, depsDir)

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := service.ExportUAssets(ctx, testDir, "")
		if !result.Success {
			b.Fatalf("IPC export failed: %s", result.Error)
		}

		// Clean up JSON files for next iteration
		cleanupJsonFiles(testDir)
	}
}

// BenchmarkUAssetExport_Native benchmarks the new NativeAOT-based export.
func BenchmarkUAssetExport_Native(b *testing.B) {
	// Setup: Create test directory with a sample asset
	testDir := setupBenchmarkAssets(b)
	defer os.RemoveAll(testDir)

	// Create native service (new method)
	appInstance := app.NewApp()
	service := NewUAssetNativeService(appInstance)

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := service.ExportUAssets(ctx, testDir, "")
		if !result.Success {
			b.Fatalf("Native export failed: %s", result.Error)
		}

		// Clean up JSON files for next iteration
		cleanupJsonFiles(testDir)
	}
}

// BenchmarkUAssetImport_IPC benchmarks the old IPC-based import.
func BenchmarkUAssetImport_IPC(b *testing.B) {
	// Setup: Create test directory with JSON files
	testDir := setupBenchmarkJsonFiles(b)
	defer os.RemoveAll(testDir)

	// Create IPC service
	app := &App{}
	depsDir := findDepsDir(b)
	service := NewUAssetService(app, depsDir)

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := service.ImportUAssets(ctx, testDir, "")
		if !result.Success {
			b.Fatalf("IPC import failed: %s", result.Error)
		}

		// Clean up .uasset files for next iteration
		cleanupUAssetFiles(testDir)
	}
}

// BenchmarkUAssetImport_Native benchmarks the new NativeAOT-based import.
func BenchmarkUAssetImport_Native(b *testing.B) {
	// Setup: Create test directory with JSON files
	testDir := setupBenchmarkJsonFiles(b)
	defer os.RemoveAll(testDir)

	// Create native service
	app := &App{}
	service := NewUAssetNativeService(app)

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := service.ImportUAssets(ctx, testDir, "")
		if !result.Success {
			b.Fatalf("Native import failed: %s", result.Error)
		}

		// Clean up .uasset files for next iteration
		cleanupUAssetFiles(testDir)
	}
}

// BenchmarkSingleAssetLoad_Native benchmarks loading a single asset (native).
func BenchmarkSingleAssetLoad_Native(b *testing.B) {
	assetPath := findTestAssetForBench(b)
	if assetPath == "" {
		b.Skip("No test asset found")
	}

	api := NewNativeUAssetAPI()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		handle, err := api.LoadAsset(assetPath, EngineVersionUE5_4, 0)
		if err != nil {
			b.Fatalf("LoadAsset failed: %v", err)
		}
		api.FreeAsset(handle)
	}
}

// BenchmarkSingleAssetSerialize_Native benchmarks serializing a single asset (native).
func BenchmarkSingleAssetSerialize_Native(b *testing.B) {
	assetPath := findTestAssetForBench(b)
	if assetPath == "" {
		b.Skip("No test asset found")
	}

	api := NewNativeUAssetAPI()

	// Load once
	handle, err := api.LoadAsset(assetPath, EngineVersionUE5_4, 0)
	if err != nil {
		b.Fatalf("LoadAsset failed: %v", err)
	}
	defer api.FreeAsset(handle)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		json, err := api.SerializeAssetToJson(handle)
		if err != nil {
			b.Fatalf("SerializeAssetToJson failed: %v", err)
		}
		_ = json // Prevent optimization
	}
}

// Benchmark helpers

func setupBenchmarkAssets(b *testing.B) string {
	b.Helper()

	testDir := b.TempDir()

	// Look for a test asset to copy
	assetPath := findTestAssetForBench(b)
	if assetPath == "" {
		b.Skip("No test asset found - skipping benchmark")
	}

	// Copy to temp directory
	destPath := filepath.Join(testDir, "test.uasset")
	copyFile(b, assetPath, destPath)

	// Copy .uexp if it exists
	uexpPath := filepath.Join(filepath.Dir(assetPath), filepath.Base(assetPath[:len(assetPath)-7])+".uexp")
	if _, err := os.Stat(uexpPath); err == nil {
		destUexp := filepath.Join(testDir, "test.uexp")
		copyFile(b, uexpPath, destUexp)
	}

	return testDir
}

func setupBenchmarkJsonFiles(b *testing.B) string {
	b.Helper()

	testDir := b.TempDir()

	// First export to get JSON
	assetPath := findTestAssetForBench(b)
	if assetPath == "" {
		b.Skip("No test asset found - skipping benchmark")
	}

	// Copy asset to temp dir
	tempAsset := filepath.Join(testDir, "test.uasset")
	copyFile(b, assetPath, tempAsset)

	// Export to JSON using native API
	api := NewNativeUAssetAPI()
	handle, err := api.LoadAsset(tempAsset, EngineVersionUE5_4, 0)
	if err != nil {
		b.Fatalf("Failed to load asset for benchmark setup: %v", err)
	}

	json, err := api.SerializeAssetToJson(handle)
	if err != nil {
		api.FreeAsset(handle)
		b.Fatalf("Failed to serialize asset for benchmark setup: %v", err)
	}
	api.FreeAsset(handle)

	// Write JSON
	jsonPath := filepath.Join(testDir, "test.json")
	err = os.WriteFile(jsonPath, []byte(json), 0644)
	if err != nil {
		b.Fatalf("Failed to write JSON for benchmark setup: %v", err)
	}

	// Remove the .uasset so import has clean slate
	os.Remove(tempAsset)

	return testDir
}

func findTestAssetForBench(b *testing.B) string {
	b.Helper()

	candidates := []string{
		"testdata/test.uasset",
		"test_asset.uasset",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func findDepsDir(b *testing.B) string {
	b.Helper()

	// Look for extracted dependencies in multiple locations
	candidates := []string{
		filepath.Join(os.TempDir(), "aris_deps"),
		"dependencies",
	}

	// Add appdata location (where dependencies are now extracted)
	if configDir, err := os.UserConfigDir(); err == nil {
		appDataDeps := filepath.Join(configDir, "ARI-S", "dependencies")
		candidates = append([]string{appDataDeps}, candidates...) // Prepend appdata path
	}

	for _, path := range candidates {
		bridgePath := filepath.Join(path, "UAssetAPI", "UAssetBridge.exe")
		if _, err := os.Stat(bridgePath); err == nil {
			return path
		}
	}

	b.Skip("Dependencies not found - cannot benchmark IPC mode")
	return ""
}

func copyFile(b *testing.B, src, dst string) {
	b.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		b.Fatalf("Failed to read file for copy: %v", err)
	}

	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		b.Fatalf("Failed to write file for copy: %v", err)
	}
}

func cleanupJsonFiles(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			os.Remove(path)
		}
		return nil
	})
}

func cleanupUAssetFiles(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".uasset" || ext == ".uexp" {
				os.Remove(path)
			}
		}
		return nil
	})
}

// TestLatencyComparison provides a user-friendly latency comparison report.
func TestLatencyComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping latency comparison in short mode")
	}

	testDir := setupBenchmarkAssets(t)
	defer os.RemoveAll(testDir)

	app := &App{}
	ctx := context.Background()

	// Warm up
	t.Log("Warming up...")

	// Test IPC
	depsDir := findDepsDir(t)
	ipcService := NewUAssetService(app, depsDir)

	// Test Native
	nativeService := NewUAssetNativeService(app)

	// Measure IPC
	t.Log("Testing IPC mode...")
	ipcStart := time.Now()
	ipcResult := ipcService.ExportUAssets(ctx, testDir, "")
	ipcDuration := time.Since(ipcStart)

	if !ipcResult.Success {
		t.Fatalf("IPC export failed: %s", ipcResult.Error)
	}

	cleanupJsonFiles(testDir)

	// Measure Native
	t.Log("Testing Native mode...")
	nativeStart := time.Now()
	nativeResult := nativeService.ExportUAssets(ctx, testDir, "")
	nativeDuration := time.Since(nativeStart)

	if !nativeResult.Success {
		t.Fatalf("Native export failed: %s", nativeResult.Error)
	}

	// Report
	t.Log("")
	t.Log("=== Latency Comparison Report ===")
	t.Logf("IPC Mode:    %v", ipcDuration)
	t.Logf("Native Mode: %v", nativeDuration)
	t.Log("")

	speedup := float64(ipcDuration) / float64(nativeDuration)
	t.Logf("Speedup: %.2fx faster", speedup)

	if speedup > 1.0 {
		improvement := (1.0 - (1.0 / speedup)) * 100
		t.Logf("Performance improvement: %.1f%%", improvement)
	}
}
