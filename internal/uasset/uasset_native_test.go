//go:build cgo
// +build cgo

package uasset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNativeUAssetAPI_GetVersion is a simple PoC test to validate the CGO toolchain.
// It calls the C# GetVersion function and verifies the response.
func TestNativeUAssetAPI_GetVersion(t *testing.T) {
	api := NewNativeUAssetAPI()

	version, err := api.GetVersion()
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}

	if version == "" {
		t.Fatal("GetVersion returned empty string")
	}

	t.Logf("UAssetAPI Version: %s", version)

	// Verify it contains expected text
	if !strings.Contains(version, "UAssetAPI") {
		t.Errorf("Version string doesn't contain 'UAssetAPI': %s", version)
	}
}

// TestNativeUAssetAPI_LoadMappings tests loading a .usmap file.
// This test is skipped if no mappings file is available.
func TestNativeUAssetAPI_LoadMappings(t *testing.T) {
	// Look for a test mappings file
	testMappingsPath := findTestMappingsFile(t)
	if testMappingsPath == "" {
		t.Skip("No test .usmap file found - skipping test")
	}

	api := NewNativeUAssetAPI()

	// Load mappings
	handle, err := api.LoadMappings(testMappingsPath)
	if err != nil {
		t.Fatalf("LoadMappings failed: %v", err)
	}

	if handle == 0 {
		t.Fatal("LoadMappings returned null handle")
	}

	t.Logf("Successfully loaded mappings: %s", testMappingsPath)

	// Free mappings
	api.FreeMappings(handle)
	t.Log("Successfully freed mappings handle")
}

// TestNativeUAssetAPI_LoadAsset tests loading a .uasset file.
// This test is skipped if no test asset is available.
func TestNativeUAssetAPI_LoadAsset(t *testing.T) {
	// Look for a test asset file
	testAssetPath := findTestAssetFile(t)
	if testAssetPath == "" {
		t.Skip("No test .uasset file found - skipping test")
	}

	api := NewNativeUAssetAPI()

	// Load asset (without mappings)
	handle, err := api.LoadAsset(testAssetPath, EngineVersionUE5_4, 0)
	if err != nil {
		t.Fatalf("LoadAsset failed: %v", err)
	}

	if handle == 0 {
		t.Fatal("LoadAsset returned null handle")
	}

	t.Logf("Successfully loaded asset: %s", testAssetPath)

	// Get export count
	count, err := api.GetAssetExportCount(handle)
	if err != nil {
		t.Fatalf("GetAssetExportCount failed: %v", err)
	}

	t.Logf("Asset has %d export(s)", count)

	// Free asset
	api.FreeAsset(handle)
	t.Log("Successfully freed asset handle")
}

// TestNativeUAssetAPI_SerializeAsset tests JSON serialization.
// This test is skipped if no test asset is available.
func TestNativeUAssetAPI_SerializeAsset(t *testing.T) {
	testAssetPath := findTestAssetFile(t)
	if testAssetPath == "" {
		t.Skip("No test .uasset file found - skipping test")
	}

	api := NewNativeUAssetAPI()

	// Load asset
	handle, err := api.LoadAsset(testAssetPath, EngineVersionUE5_4, 0)
	if err != nil {
		t.Fatalf("LoadAsset failed: %v", err)
	}
	defer api.FreeAsset(handle)

	// Serialize to JSON
	json, err := api.SerializeAssetToJson(handle)
	if err != nil {
		t.Fatalf("SerializeAssetToJson failed: %v", err)
	}

	if len(json) == 0 {
		t.Fatal("SerializeAssetToJson returned empty string")
	}

	t.Logf("Serialized asset to JSON (%d bytes)", len(json))

	// Verify JSON structure
	if !strings.Contains(json, "{") {
		t.Error("JSON doesn't contain opening brace")
	}
}

// TestNativeUAssetAPI_RoundTrip tests the full load -> serialize -> deserialize -> write cycle.
// This test is skipped if no test asset is available.
func TestNativeUAssetAPI_RoundTrip(t *testing.T) {
	testAssetPath := findTestAssetFile(t)
	if testAssetPath == "" {
		t.Skip("No test .uasset file found - skipping test")
	}

	api := NewNativeUAssetAPI()

	// Step 1: Load asset
	t.Log("Step 1: Loading asset...")
	handle1, err := api.LoadAsset(testAssetPath, EngineVersionUE5_4, 0)
	if err != nil {
		t.Fatalf("LoadAsset failed: %v", err)
	}
	defer api.FreeAsset(handle1)

	// Step 2: Serialize to JSON
	t.Log("Step 2: Serializing to JSON...")
	json, err := api.SerializeAssetToJson(handle1)
	if err != nil {
		t.Fatalf("SerializeAssetToJson failed: %v", err)
	}
	t.Logf("  JSON size: %d bytes", len(json))

	// Step 3: Deserialize from JSON
	t.Log("Step 3: Deserializing from JSON...")
	handle2, err := api.DeserializeAssetFromJson(json)
	if err != nil {
		t.Fatalf("DeserializeAssetFromJson failed: %v", err)
	}
	defer api.FreeAsset(handle2)

	// Step 4: Get export counts to verify
	t.Log("Step 4: Verifying export counts...")
	count1, err := api.GetAssetExportCount(handle1)
	if err != nil {
		t.Fatalf("GetAssetExportCount (original) failed: %v", err)
	}

	count2, err := api.GetAssetExportCount(handle2)
	if err != nil {
		t.Fatalf("GetAssetExportCount (deserialized) failed: %v", err)
	}

	if count1 != count2 {
		t.Errorf("Export count mismatch: original=%d, deserialized=%d", count1, count2)
	}

	t.Logf("  Export count: %d (matches!)", count1)

	// Step 5: Write to temp file
	t.Log("Step 5: Writing to temp file...")
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_roundtrip.uasset")

	err = api.WriteAssetToFile(handle2, outputPath, 0)
	if err != nil {
		t.Fatalf("WriteAssetToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file not created: %s", outputPath)
	}

	t.Logf("  Output written to: %s", outputPath)
	t.Log("Round-trip test PASSED!")
}

// Helper function to find a test mappings file
func findTestMappingsFile(t *testing.T) string {
	// Look in common locations
	candidates := []string{
		"testdata/Grounded.usmap",
		"test_mappings.usmap",
		"../Resources/Grounded.usmap",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// Helper function to find a test asset file
func findTestAssetFile(t *testing.T) string {
	// Look in common locations
	candidates := []string{
		"testdata/test.uasset",
		"test_asset.uasset",
		"../GameFiles/Maine/Content/Test.uasset",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
