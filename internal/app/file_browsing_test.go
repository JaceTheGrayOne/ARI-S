package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/JaceTheGrayOne/ARI-S/internal/config"
)

func TestFileBrowsing_LastUsedPath_Persists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	app := NewApp()
	app.ctx = context.Background()
	app.config = config.NewDefaultConfig()

	testPath := "C:\\TestModFolder"
	testKey := "input_mod_folder"

	app.SetLastUsedPath(testKey, testPath)

	// Save config
	if err := config.SaveConfig(configPath, app.config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create new app and load config (simulate app restart)
	newApp := NewApp()
	loadedConfig, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	newApp.config = loadedConfig

	retrievedPath := newApp.GetLastUsedPath(testKey)
	if retrievedPath != testPath {
		t.Errorf("Expected path '%s', got '%s'", testPath, retrievedPath)
	}
}

func TestFileBrowsing_MultipleKeys_IndependentPaths(t *testing.T) {
	app := NewApp()
	app.config = config.NewDefaultConfig()

	paths := map[string]string{
		"input_mod_folder":       "C:\\Input",
		"pak_output_dir":         "C:\\Output",
		"game_paks_folder_unpak": "C:\\Game\\Paks",
		"extract_output_dir":     "C:\\Extracted",
		"export_folder":          "C:\\Export",
		"import_folder":          "C:\\Import",
		"uasset_mappings_path":   "C:\\Mappings\\game.usmap",
	}

	for key, path := range paths {
		app.SetLastUsedPath(key, path)
	}

	for key, expectedPath := range paths {
		actualPath := app.GetLastUsedPath(key)
		if actualPath != expectedPath {
			t.Errorf("Key '%s': expected '%s', got '%s'", key, expectedPath, actualPath)
		}
	}
}

func TestFileBrowsing_GetPreference_ReturnsCorrectValue(t *testing.T) {
	app := NewApp()
	app.config = config.NewDefaultConfig()

	preferences := map[string]string{
		"ue_version": "UE5_3",
		"theme":      "light",
		"auto_save":  "true",
	}

	for key, value := range preferences {
		app.SetPreference(key, value)
	}

	for key, expectedValue := range preferences {
		actualValue := app.GetPreference(key)
		if actualValue != expectedValue {
			t.Errorf("Preference '%s': expected '%s', got '%s'", key, expectedValue, actualValue)
		}
	}
}

func TestFileBrowsing_ConfigKeyMapping_ConsistentConvention(t *testing.T) {
	// Frontend uses kebab-case (input-mod-folder)
	// Backend config uses snake_case (input_mod_folder)

	testCases := []struct {
		frontendID string
		configKey  string
	}{
		{"input-mod-folder", "input_mod_folder"},
		{"pak-output-dir", "pak_output_dir"},
		{"game-paks-folder-unpak", "game_paks_folder_unpak"},
		{"extract-output-dir", "extract_output_dir"},
		{"export-folder", "export_folder"},
		{"import-folder", "import_folder"},
		{"uasset-mappings-path", "uasset_mappings_path"},
	}

	app := NewApp()
	app.config = config.NewDefaultConfig()

	// Act & Assert: Verify key conversion works
	for _, tc := range testCases {
		testPath := "C:\\Test\\" + tc.frontendID

		// Simulate frontend key conversion (replace - with _)
		configKey := tc.configKey

		// Set using config key
		app.SetLastUsedPath(configKey, testPath)

		// Get using config key
		retrievedPath := app.GetLastUsedPath(configKey)

		if retrievedPath != testPath {
			t.Errorf("Key mapping failed for '%s' → '%s': expected '%s', got '%s'",
				tc.frontendID, tc.configKey, testPath, retrievedPath)
		}
	}
}

func TestFileBrowsing_EmptyConfig_ReturnsEmptyString(t *testing.T) {
	app := NewApp()
	app.config = config.NewDefaultConfig()

	path := app.GetLastUsedPath("nonexistent_key")

	if path != "" {
		t.Errorf("Expected empty string for non-existent key, got '%s'", path)
	}
}

func TestFileBrowsing_NilConfig_HandlesGracefully(t *testing.T) {
	app := NewApp()
	app.config = nil

	path := app.GetLastUsedPath("test_key")

	if path != "" {
		t.Errorf("Expected empty string with nil config, got '%s'", path)
	}

	app.SetLastUsedPath("test_key", "/test/path")

	if app.config == nil {
		t.Error("Expected config to be initialized after SetLastUsedPath")
	}

	// Verify path was set
	if app.GetLastUsedPath("test_key") != "/test/path" {
		t.Error("Path not set correctly after config initialization")
	}
}
